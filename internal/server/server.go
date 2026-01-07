package server

import (
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	mwecho "github.com/labstack/echo/v4/middleware"
	mwsvc "winsbygroup.com/regserver/internal/middleware"

	"winsbygroup.com/regserver/internal/activation"
	"winsbygroup.com/regserver/internal/config"
	"winsbygroup.com/regserver/internal/customer"
	"winsbygroup.com/regserver/internal/demodata"
	"winsbygroup.com/regserver/internal/feature"
	"winsbygroup.com/regserver/internal/featurevalue"
	"winsbygroup.com/regserver/internal/license"
	"winsbygroup.com/regserver/internal/machine"
	"winsbygroup.com/regserver/internal/product"
	"winsbygroup.com/regserver/internal/registration"
	"winsbygroup.com/regserver/internal/sqlite"
	"winsbygroup.com/regserver/static"

	adminhttp "winsbygroup.com/regserver/internal/http/admin"
	clienthttp "winsbygroup.com/regserver/internal/http/client"
	webhttp "winsbygroup.com/regserver/internal/http/web"
)

type Server struct {
	Echo *echo.Echo
	HTTP *http.Server
	DB   *sqlx.DB
}

func Build(cfg *config.Config) (*Server, error) {
	//
	// Validate required environment variables
	//
	if os.Getenv("ADMIN_API_KEY") == "" {
		return nil, errors.New("ADMIN_API_KEY environment variable is required")
	}
	if cfg.RegistrationSecret == "" {
		return nil, errors.New("REGISTRATION_SECRET environment variable is required")
	}

	//
	// Database
	//
	isNewDB := false
	if _, err := os.Stat(cfg.DBPath); os.IsNotExist(err) {
		isNewDB = true
		log.Printf("Creating database '%s' (from %s setting)", cfg.DBPath, cfg.DBPathSource)
	} else {
		log.Printf("Opening database '%s' (from %s setting)", cfg.DBPath, cfg.DBPathSource)
	}
	db, err := sqlx.Connect("sqlite3", cfg.DBPath)
	if err != nil {
		return nil, err
	}

	// WAL mode is only required once after creating the database, but
	// doesn't hurt to set it each time
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		return nil, err
	}

	// Foreign key support is required each time the database is open and
	// is required by the program for cascade deletes
	if _, err := db.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		return nil, err
	}

	// Verify foreign keys are supported and enabled
	var fkEnabled int
	if err := db.QueryRow(`PRAGMA foreign_keys;`).Scan(&fkEnabled); err != nil {
		return nil, errors.New("SQLite foreign key support check failed: " + err.Error())
	}
	if fkEnabled != 1 {
		return nil, errors.New("SQLite foreign keys not supported (requires SQLite 3.6.19+ compiled without SQLITE_OMIT_FOREIGN_KEY)")
	}

	if err := sqlite.RunMigrations(db.DB); err != nil {
		return nil, err
	}

	// Load demo data if requested and database is new
	if cfg.DemoMode && isNewDB {
		if err := demodata.Load(db.DB); err != nil {
			return nil, errors.New("failed to load demo data: " + err.Error())
		}
		log.Print("Demo data loaded")
	}

	//
	// Domain services
	//
	customerSvc := customer.NewService(db)
	productSvc := product.NewService(db)
	licenseSvc := license.NewService(db)
	featureSvc := feature.NewService(db)
	featureValueSvc := featurevalue.NewService(db)
	machineSvc := machine.NewService(db)
	registrationSvc := registration.NewService(db)

	activationSvc := activation.NewService(
		db,
		cfg.RegistrationSecret,
		customerSvc,
		machineSvc,
		registrationSvc,
		licenseSvc,
		productSvc,
		featureSvc,
		featureValueSvc,
	)

	//
	// Handlers
	//
	clientHandler := clienthttp.NewHandler(
		activationSvc,
		registrationSvc,
		productSvc,
		licenseSvc,
		machineSvc,
		featureSvc,
		featureValueSvc,
		customerSvc,
	)

	adminSvc := adminhttp.NewService(
		customerSvc,
		productSvc,
		licenseSvc,
		featureSvc,
		featureValueSvc,
		machineSvc,
		registrationSvc,
	)
	adminHandler := adminhttp.NewHandler(adminSvc)

	webHandler := webhttp.NewHandler(
		adminSvc,
		productSvc,
		featureSvc,
		featureValueSvc,
		machineSvc,
		registrationSvc,
		activationSvc,
	)

	//
	// Echo
	//
	e := echo.New()
	e.HideBanner = true

	// Health endpoints
	e.GET("/livez", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	e.GET("/readyz", func(c echo.Context) error {
		if err := db.Ping(); err != nil {
			return c.String(http.StatusServiceUnavailable, "DB not ready")
		}
		return c.String(http.StatusOK, "Ready")
	})

	// Middleware
	e.Use(mwecho.Logger())
	e.Use(mwecho.Recover())

	// Client API
	clientGroup := e.Group("/api/v1")
	clienthttp.RegisterRoutes(clientGroup, clientHandler, mwsvc.LicenseKeyAuth(db))

	// Admin API
	adminGroup := e.Group("/api/admin")
	adminGroup.Use(mwsvc.AdminAPIKeyAuth())
	adminhttp.RegisterRoutes(adminGroup, adminHandler)

	// Web UI
	webGroup := e.Group("/web")
	webGroup.Use(mwsvc.Theme())            // Read theme cookie into context
	webGroup.Use(mwsvc.Version())          // Add app version to context
	webGroup.Use(mwsvc.DemoMode(cfg.DemoMode)) // Add demo mode flag to context
	webGroup.Use(mwsvc.WebAuth())
	webGroup.Use(mwecho.CSRFWithConfig(mwecho.CSRFConfig{
		TokenLookup:    "header:X-CSRF-Token,form:_csrf",
		CookieName:     "_csrf",
		CookiePath:     "/",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
		Skipper: func(c echo.Context) bool {
			// Skip CSRF for login page (user not authenticated yet)
			return strings.HasPrefix(c.Path(), "/web/login")
		},
	}))
	webGroup.Use(mwsvc.CSRF()) // Copy CSRF token to request context for templates
	webhttp.RegisterRoutes(webGroup, webHandler)

	// Static files (embedded)
	jsFS, _ := fs.Sub(static.Files, "js")
	e.GET("/static/js/*", echo.WrapHandler(http.StripPrefix("/static/js/", http.FileServer(http.FS(jsFS)))))
	cssFS, _ := fs.Sub(static.Files, "css")
	e.GET("/static/css/*", echo.WrapHandler(http.StripPrefix("/static/css/", http.FileServer(http.FS(cssFS)))))

	//
	// HTTP server
	//
	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      e,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Server{
		Echo: e,
		HTTP: srv,
		DB:   db,
	}, nil
}
