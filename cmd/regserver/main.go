package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"winsbygroup.com/regserver/internal/config"
	"winsbygroup.com/regserver/internal/server"
	"winsbygroup.com/regserver/internal/version"
)

func main() {
	fmt.Println(version.Banner())

	//
	// Flags
	//
	configPath := flag.String("config", "config.yaml", "path to config file")
	routesFlag := flag.Bool("routes", false, "print routes and exit")
	demoFlag := flag.Bool("demo", false, "load sample data on new database (for demos)")
	flag.Parse()

	//
	// Load configuration
	//
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	cfg.DemoMode = *demoFlag

	//
	// Build server (Echo, DB, services, etc.)
	//
	srv, err := server.Build(cfg)
	if err != nil {
		log.Fatalf("failed to build server: %v", err)
	}
	defer srv.DB.Close()

	//
	// Routes inspection mode
	//
	if *routesFlag {
		routes := srv.Echo.Routes()
		sort.Slice(routes, func(i, j int) bool {
			return routes[i].Path < routes[j].Path
		})

		for _, r := range routes {
			fmt.Printf("%-6s %s\n", r.Method, r.Path)
		}

		os.Exit(0)
	}

	//
	// Normal server startup
	//
	go func() {
		if err := srv.Echo.StartServer(srv.HTTP); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srv.Echo.Logger.Fatalf("server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Echo.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
