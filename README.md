# GoExplore

GoExplore is a cross-platform desktop file manager built with Wails v2 and Vanilla JS. It provides a unified UI for browsing and managing files across multiple remote and local filesystems.

## Requirements
- Go 1.26
- Wails v2 CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- Node.js & npm
- Platform-specific dependencies for Wails (e.g. `libgtk-3-dev`, `libwebkit2gtk-4.0-dev` on Linux)

## Configuration
Configuration is stored in `~/.config/goexplore/config.yaml`.
All secrets (passwords, keys) are stored securely in the OS keychain using `go-keyring`.

## Build Instructions
1. Run `wails build -tags webkit2_41` to compile the production binary.
2. The output will be located in the `build/bin/` directory.

## Development
To run in development mode with live reload:
```bash
wails dev -tags webkit2_41
```
