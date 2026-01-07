# Hybrid Windows + WSL Development Workflow (GoLand 2025.3)

## Overview
This document defines a **hybrid development workflow** designed for developers who:

- Work primarily on **Windows**
- Keep all source code under a single Windows directory
- Need to build, test, and debug **Linux-native Go binaries**
- Use **GoLand 2025.3** (which no longer supports WSL toolchains)
- Want a clean, reproducible, and stable setup without Remote Dev or VM shared-folder issues

This workflow keeps the project on Windows while running Linux-native tooling inside WSL.

**Note:** Using the Delve debugger is not a requirement for development. Those Delve-specific
instructions can be skipped if all you want is to edit in GoLand Windows and build and deploy
from WSL.

---

## Architecture Summary

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              WINDOWS HOST                                   │
│  ┌───────────────────────────────────┐    ┌──────────────────────────────┐  │
│  │           GoLand IDE              │    │     Windows Filesystem       │  │
│  │  ┌─────────────────────────────┐  │    │                              │  │
│  │  │  • Code editing             │  │    │  C:\Users\dougw\Source\      │  │
│  │  │  • File watching            │  │    │    Products\regserver\       │  │
│  │  │  • Git integration          │  │    │      ├── cmd/                │  │
│  │  │  • Search & refactoring     │  │    │      ├── internal/           │  │
│  │  │  • Go Remote debug config   │──┼────│      ├── templates/          │  │
│  │  └─────────────────────────────┘  │    │      └── ...                 │  │
│  └───────────────────────────────────┘    └──────────────────────────────┘  │
│                    │                                    ▲                   │
│                    │ TCP :40000                         │                   │
│                    │ (Debug attach)                     │ /mnt/c/           │
│                    ▼                                    │ (filesystem)      │
│  ┌──────────────────────────────────────────────────────┴───────────────┐   │
│  │                           WSL (Ubuntu)                               │   │
│  │  ┌────────────────┐  ┌─────────────────┐  ┌───────────────────────┐  │   │
│  │  │  Go Toolchain  │  │ Delve Debugger  │  │    Other Tools        │  │   │
│  │  │  • go build    │  │ • Headless mode │  │  • Taskfile           │  │   │
│  │  │  • go test     │  │ • Port 40000    │  │  • Caddy              │  │   │
│  │  │  • go run      │  │ • API v2        │  │  • SQLite             │  │   │
│  │  └────────────────┘  └─────────────────┘  └───────────────────────┘  │   │
│  │                                                                      │   │
│  │  ~/regserver -> /mnt/c/Users/<username>/Source/regserver             │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘

Data Flow:
  1. Edit code in GoLand (Windows) → saved to Windows filesystem
  2. WSL accesses same files via /mnt/c mount
  3. Delve builds & runs Linux binary, listens on :40000
  4. GoLand attaches via "Go Remote" config over TCP
  5. Breakpoints, variables, and stack traces work seamlessly
```

### Windows
- **GoLand IDE** (full UI, file watching, editing, navigation)
- **Source code lives entirely on Windows**  
  Example:
  ```
  C:\Users\<username>\Source\regserver
  ```
- Git, search, refactoring, and editor features run at full speed

### WSL (Ubuntu)
- **Go toolchain** (Linux-native)
- **Delve debugger**
- **Taskfile**
- **Caddy**, **SQLite**, and any other Linux tools
- All Linux commands run against the Windows-mounted project directory

### Communication
- GoLand attaches to Delve over TCP using a **Go Remote** configuration
- Delve runs inside WSL in headless mode
- Breakpoints, variables, goroutines, and stack traces work normally

## Why This Workflow

### Keep all source code on Windows  
Easy backups, easy navigation, one place for everything.

### Linux-native builds/tests/debugging  
Matches production behavior without moving the repo.

### No JetBrains Gateway required  
GoLand runs locally and attaches to Delve manually.

### No WSL plugin required  
GoLand 2025.3 removed WSL toolchain support; this workflow replaces it cleanly.

### No VM shared-folder performance issues  
WSL access to `/mnt/c` is significantly faster and more stable than VMware shared folders (for example).

## WSL Setup

### 1. Create a symlink to your Windows project (optional but recommended)
```bash
ln -s /mnt/c/Users/jimmy/Source/regserver ~/regserver
```

### 2. Install Go inside WSL
```bash
sudo apt update
sudo apt install golang-go
```

Or install a specific version under `/usr/local/go`.

### 3. Install Delve inside WSL
```bash
sudo apt install dlv
```
or:
```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

## Running Delve in WSL

From WSL:

```bash
cd ~/registration-server
dlv debug --headless --listen=:40000 --api-version=2 --accept-multiclient
```

This:

- Builds the Go binary with debug flags  
- Runs it inside Linux  
- Opens a Delve server on port 40000  
- Waits for GoLand to attach

## GoLand Configuration (Windows)

### 1. Create a **Go Remote** configuration
In GoLand:

```
Run ? Edit Configurations ? + ? Go Remote
```

Set:

- **Host:** `localhost`
- **Port:** `40000`
- **Mode:** `Connect`
- **Debugger:** Delve

Save the configuration.

### 2. Attach the debugger
- Start Delve in WSL  
- Run the **Go Remote** config in GoLand  
- Debug as usual (breakpoints, variables, goroutines, etc.)

## Taskfile, Caddy, and Other Tools

All Linux-native tools run inside WSL using the symlinked project directory:

```bash
cd ~/registration-server
task build
task test
caddy run
sqlite3 my.db
```

GoLand's terminal can be configured to open WSL by default:

```
Settings ? Tools ? Terminal ? Shell path: wsl.exe ~
```

## Limitations

- GoLand cannot run Go builds/tests/debugging *directly* inside WSL  
  (WSL toolchain support was removed in 2025.3)
- File watching inside WSL does not work on `/mnt/c`, but this does not affect GoLand  
  (GoLand watches the Windows filesystem natively)

## Benefits

- Full Windows-native IDE performance  
- Full Linux-native runtime behavior  
- No need to move the repo  
- No need for Remote Dev  
- No need for VMware shared folders  
- Clean, explicit, reproducible workflow  
- Works perfectly with GoLand 2025.3+
