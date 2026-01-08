# RegServer - Claude Development Notes

## Project Overview

This is a Go-based software license registration server with:
- **REST API** for programmatic access (admin and public endpoints)
- **Web UI** for administration (Templ + HTMX + DaisyUI)
- **SQLite database** with migrations

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      Echo Router                        │
├──────────────┬──────────────┬──────────────┬────────────┤
│ /api/admin/* │   /api/v1/*  │   /web/*     │  /static/* │
│  Admin API   │  Client API  │   Web UI     │   Assets   │
├──────────────┴──────────────┴──────────────┴────────────┤
│                    Service Layer                        │
│  (customer, product, license, feature, etc.)            │
├─────────────────────────────────────────────────────────┤
│                   Repository Layer                      │
├─────────────────────────────────────────────────────────┤
│                        SQLite                           │
└─────────────────────────────────────────────────────────┘
```

### Key Design Patterns

1. **Domain-Driven Structure**: Each domain (customer, product, etc.) has its own package with service and repository
2. **Service Facade**: `internal/http/admin/service.go` aggregates all domain services for the admin API
3. **View Models**: `internal/viewmodels/` contains types for templates, separate from domain models to avoid import cycles
4. **Conversion Functions**: `internal/http/web/models.go` converts domain models to view models

## Important Directories

| Path | Purpose |
|------|---------|
| `cmd/regserver/` | Application entry point |
| `internal/server/` | Server builder - wires everything together |
| `internal/http/admin/` | Admin REST API (service facade + handlers) |
| `internal/http/client/` | Client API handlers (activation, product version) |
| `internal/http/web/` | Web UI handlers |
| `internal/activation/` | Activation service (orchestrates license activation) |
| `internal/viewmodels/` | Template view models (breaks import cycle) |
| `templates/` | Templ templates (.templ files) |
| `internal/sqlite/` | Database connection, migrations, schema |

## Key Files to Understand

1. **`internal/server/server.go`** - Server setup, dependency injection, route registration
2. **`internal/http/admin/service.go`** - Service facade aggregating all domain services
3. **`internal/http/client/handler.go`** - Client API handlers (activate, productver endpoints)
4. **`internal/http/web/handler.go`** - Web UI handlers (~780 lines, all CRUD operations)
5. **`internal/http/web/models.go`** - Domain-to-viewmodel conversion functions
6. **`internal/activation/service.go`** - Activation orchestration (license checks, machine registration)
7. **`internal/viewmodels/models.go`** - View model types used by templates

## Domain Models

| Domain | Key Fields |
|--------|------------|
| `customer` | CustomerID, CustomerName, ContactName, Phone, Email, Notes |
| `product` | ProductID, ProductName, ProductGUID, LatestVersion, DownloadURL |
| `license` | CustomerID, ProductID, LicenseKey, LicenseCount, IsSubscription, LicenseTerm, StartDate, ExpirationDate, MaintExpirationDate, MaxProductVersion |
| `feature` | FeatureID, ProductID, FeatureName, FeatureType (0=Integer, 1=String, 2=Values), AllowedValues (pipe-delimited), DefaultValue |
| `featurevalue` | CustomerID, ProductID, FeatureID, FeatureValue |
| `machine` | MachineID, CustomerID, MachineCode, UserName |
| `registration` | MachineID, ProductID, RegistrationHash, ExpirationDate, FirstRegistrationDate, LastRegistrationDate |

### GUID Conventions

All GUIDs are normalized to **lowercase** on insert/update (in repo layer) and stored with `COLLATE NOCASE` for case-insensitive lookups. Users can enter either case.

### Feature Value Override Pattern

The `license_feature` table uses an **override-only storage pattern**:

- **Default values** are defined on the `feature` table (product-level)
- **Override values** are stored in `license_feature` (customer-level)
- When displaying/returning features, the system joins both tables and uses `COALESCE(override, default)`

**Benefits:**
- Database stays lean (no redundant default values per customer)
- Changing a product's default affects all customers who haven't overridden it
- Easy to identify which customers have custom configurations

**SQL Pattern:**
```sql
-- Get effective feature values (used in getEnrichedFeatureValues)
SELECT f.*, COALESCE(lf.feature_value, f.default_value) AS effective_value
FROM feature f
LEFT JOIN license_feature lf
  ON f.feature_id = lf.feature_id
  AND lf.customer_id = ? AND lf.product_id = ?
WHERE f.product_id = ?
```

**UPSERT for Updates:**
```sql
-- featurevalue/sql.go uses INSERT ON CONFLICT to handle both insert and update
INSERT INTO license_feature (customer_id, product_id, feature_id, feature_value)
VALUES (?, ?, ?, ?)
ON CONFLICT (customer_id, product_id, feature_id) DO UPDATE SET feature_value = excluded.feature_value
```

This UPSERT is necessary because the row may not exist if the customer has never overridden the default.

## Template Architecture

Templates use the `viewmodels` package to avoid import cycles:

```
templates/
├── layouts/base.templ      # HTML shell with nav, logout button, modal container
├── pages/
│   ├── index.templ         # Registrations (customer selector + registrations)
│   ├── login.templ         # Login page (standalone, no base layout)
│   ├── customers.templ     # Customer list page
│   └── products.templ      # Product catalog page
└── components/
    ├── customers_table.templ      # Customer CRUD
    ├── products_table.templ       # Product CRUD
    ├── licenses_table.templ       # License management
    ├── features_table.templ       # Feature value editing
    ├── machines_modal.templ       # Machine registration list
    └── empty_state.templ          # Empty state placeholder
```

### Import Pattern for Templates

```go
import (
    vm "winsbygroup.com/regserver/internal/viewmodels"
)

templ SomeComponent(data []vm.Customer) {
    // ...
}
```

**Never import `internal/http/web` in templates** - this creates import cycles.

## Client API Routes

Client API routes are under `/api/v1/` prefix. Authentication uses `X-License-Key` header with the customer's license key (for `/activate`), or license key in URL (for `/license/:license_key`).

| Route | Handler | Description |
|-------|---------|-------------|
| POST `/api/v1/activate` | Activate | Activate product for a machine |
| GET `/api/v1/productver/:guid` | GetProductVersion | Get product version info (returns 404 for unknown GUID) |
| GET `/api/v1/license/:license_key` | GetLicenseInfo | Get license info and availability (LicensesAvailable counts only non-expired registrations) |
| PUT `/api/v1/license/:license_key` | UpdateLicenseInfo | Update installed version for a machine |

Key files:
- `internal/http/client/handler.go` - Request handlers
- `internal/http/client/routes.go` - Route registration
- `internal/activation/service.go` - Activation business logic

## Web Routes

All web routes are under `/web/` prefix. Authentication uses cookie-based sessions (login page) or `X-API-Key` header.

| Route | Handler | Description |
|-------|---------|-------------|
| GET `/web/login` | LoginPage | Login form |
| POST `/web/login` | Login | Validate key, set cookie |
| POST `/web/logout` | Logout | Clear cookie, redirect to login |
| GET `/web/` | Index | Registrations |
| GET `/web/customers` | ListCustomers | Customer page |
| GET `/web/customers/new` | NewCustomerForm | HTMX partial |
| POST `/web/customers` | CreateCustomer | Create + return table |
| GET `/web/customers/:id/edit` | EditCustomerForm | HTMX partial |
| PUT `/web/customers/:id` | UpdateCustomer | Update + return table |
| DELETE `/web/customers/:id` | DeleteCustomer | Delete + return table |
| Similar patterns for products, registrations, features, machines... |

## HTMX Patterns Used

1. **Table refresh**: Forms target `#<table>-container` and swap innerHTML
2. **Modal forms**: Buttons target `#modal-content`, JS opens modal on htmx:afterSwap
3. **Row deletion**: Delete buttons use hx-confirm for confirmation
4. **Partial updates**: Most handlers return just the component, not full page

## Build Commands

```bash
# Generate templates (required before build)
templ generate

# Build
go build -o ./dist/regserver ./cmd/regserver

# Or use Task
task build      # generate + build
task run        # go run without build
task generate   # just template generation
```

## Database

- **Location**: Configured in config.yaml or `DB_PATH` env var
- **Default**: `./registrations.db`
- **Migrations**: `internal/sqlite/migrate.go` using darwin

### SQLite Configuration

The following PRAGMAs are set on connection (`internal/server/server.go`):

| PRAGMA | Value | Purpose |
|--------|-------|---------|
| `journal_mode` | WAL | Better concurrency, crash recovery |
| `foreign_keys` | ON | Enforce referential integrity |

**Note:** SQLite disables foreign keys by default for backwards compatibility. With `foreign_keys=ON`, inserts/updates that violate foreign key constraints will fail. Tests also enable this via `internal/testutil/testdb.go`.

### Database Validation

The database uses SQLite's `application_id` pragma to prevent accidentally using another application's database:

| Constant | Value | Meaning |
|----------|-------|---------|
| `ApplicationID` | `0x52454753` | "REGS" in ASCII |

**Validation logic** (in `VerifyApplicationID`):

| Condition | Result |
|-----------|--------|
| `application_id = 0x52454753` | ✓ Accept (regserver database) |
| `application_id = 0` AND no tables | ✓ Accept (new database) |
| `application_id = 0` AND has tables | ✗ Reject |
| `application_id = other` | ✗ Reject |

This runs before migrations to prevent corrupting another application's database.

## Common Tasks

### Adding a New Field to a Domain

1. Update domain model in `internal/<domain>/<domain>.go`
2. Update repository SQL in `internal/<domain>/repository.go`
3. Update view model in `internal/viewmodels/models.go`
4. Update conversion function in `internal/http/web/models.go`
5. Update templates that display the field
6. Run `templ generate && go build`

### Adding a New Feature

1. Add service method in appropriate domain package
2. Add handler in `internal/http/web/handler.go`
3. Add route in `internal/http/web/routes.go`
4. Create/update templates as needed
5. Run `templ generate && go build`

### Debugging Template Issues

1. Check that templates import `viewmodels` not `web`
2. Run `templ generate` to regenerate `*_templ.go` files
3. Check for import cycle errors in build output

## Testing

```bash
task test           # Run all tests
task test:errors    # Show only failures (platform-specific)

# Run tests for a specific package
go test ./internal/http/client/... -v
go test ./internal/activation/... -v
go test ./internal/product/... -v
```

Test files use `internal/testutil/testdb.go` which creates an in-memory SQLite database with migrations applied. Tests follow the `_test` package convention (e.g., `package client_test`).

## Platform Notes

- **Cross-platform**: Works on Windows, Linux, macOS
- **Taskfile**: Uses bash commands (`mkdir -p`, `rm -f`) - requires Git Bash on Windows
- **Path separators**: All code uses forward slashes

## Authentication

- **Admin API** (`/api/admin/*`): `X-API-Key` header validated against `ADMIN_API_KEY` env var
- **Web UI** (`/web/*`): Session-based auth with CSRF protection (see below)
- **Client API** (`/api/v1/*`): `X-License-Key` header with customer's license key

### Web UI Auth Flow

1. User visits `/web/login` and enters their `ADMIN_API_KEY`
2. Server validates key, creates an in-memory session, returns session ID in cookie
3. Subsequent requests validated via session ID lookup (not the API key itself)
4. Sessions expire after 7 days; cleared on logout or server restart
5. CSRF token required for all POST/PUT/DELETE requests

**Key points:**
- Session IDs are UUIDs stored in an in-memory map (no database persistence)
- The API key is never stored in the cookie - only the session ID
- `X-API-Key` header still works for programmatic access to `/web/*` routes
- CSRF tokens passed via `X-CSRF-Token` header (HTMX) or `_csrf` form field

### Middleware Package Structure

```
internal/middleware/
├── auth.go      # WebAuth, AdminAPIKeyAuth, LicenseKeyAuth, ValidateAdminKey
├── session.go   # Session struct, CreateSession, GetSession, DeleteSession
├── csrf.go      # CSRF middleware, GetCSRF helper (for templates)
├── context.go   # Theme, GetTheme, Version, GetVersion
```

### Key Files

- `internal/middleware/auth.go` - Auth middlewares
- `internal/middleware/session.go` - In-memory session management
- `internal/server/server.go` - Startup check, CSRF middleware configuration
- `internal/http/web/handler.go` - Login/Logout handlers
- `templates/pages/login.templ` - Login page template
- `templates/layouts/base.templ` - CSRF meta tag and hidden field for logout form

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `ADMIN_API_KEY` | **Yes** | Server refuses to start without this |
| `REGISTRATION_SECRET` | **Yes** | Secret key appended before hashing registration data |
| `DB_PATH` | No | Database path (default: `./registrations.db`) |

**Warning:** Changing `REGISTRATION_SECRET` invalidates all existing registrations - clients will need to re-activate.

## Theme System

The Web UI supports Light and Dark themes using DaisyUI.

### How It Works

1. **Cookie**: User preference stored in `theme` cookie (light/dark)
2. **Middleware**: `Theme()` middleware reads cookie, adds to request context
3. **Template**: `base.templ` reads theme from context via `middleware.GetTheme(ctx)`
4. **Rendering**: Server renders correct `data-theme` attribute - no client-side JS needed for initial load

### Key Files

| File | Purpose |
|------|---------|
| `internal/middleware/context.go` | `Theme()` middleware, `GetTheme()` helper |
| `templates/layouts/base.templ` | Reads theme from context, renders `data-theme` attribute |
| `static/js/app.js` | `toggleTheme()` for switching, `updateThemeUI()` for button state |

### Adding Theme to Context

Templates access theme via context. The middleware adds it automatically for `/web/*` routes:

```go
// In server.go
webGroup.Use(mwsvc.Theme())

// In templates (base.templ)
import "winsbygroup.com/regserver/internal/middleware"
data-theme={ middleware.GetTheme(ctx) }
```

## Tom Select (Searchable Dropdowns)

Large dropdowns (like customer selection) use Tom Select for type-to-search functionality.

### CDN Setup

Loaded in `base.templ`:
```html
<link href="https://cdn.jsdelivr.net/npm/tom-select@2.3.1/dist/css/tom-select.css" rel="stylesheet"/>
<script src="https://cdn.jsdelivr.net/npm/tom-select@2.3.1/dist/js/tom-select.complete.min.js"></script>
```

### Initialization

Tom Select is initialized in `static/js/app.js` on DOMContentLoaded:

```js
var customerSelect = document.getElementById('customer-select');
if (customerSelect && typeof TomSelect !== 'undefined') {
    new TomSelect(customerSelect, {
        create: false,
        sortField: { field: 'text', direction: 'asc' },
        onChange: function(value) {
            loadCustomerLicenses(value);
        }
    });
}
```

### Styling

Custom CSS in `base.templ` integrates Tom Select with DaisyUI theme colors (`.ts-control`, `.ts-dropdown`, etc.).

### Adding to New Dropdowns

1. Add a plain `<select>` with a unique ID (no DaisyUI select classes)
2. Initialize in `app.js` with `new TomSelect('#your-select-id', { options })`
3. Use `onChange` callback instead of inline `onchange` attribute

## Gotchas

1. **Import cycles**: Templates must use `viewmodels`, not `web` package
2. **Template regeneration**: Always run `templ generate` after changing `.templ` files
3. **Templ version**: Keep CLI and go.mod version in sync (`go get -u github.com/a-h/templ`)
4. **HTMX responses**: Most web handlers return partials, check `isHTMX()` for full page vs partial
5. **Context types**: Helper methods in handler.go use `context.Context`, not `echo.Context`
