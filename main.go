package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

var Version string = "dev"

const (
	linuxApplicationID = "uk.co.jdbnet.goexplore"
	linuxIconThemeName = "uk.co.jdbnet.goexplore"
)

func linuxIconPath(home string) string {
	return filepath.Join(home, ".local", "share", "icons", "hicolor", "256x256", "apps", linuxIconThemeName+".png")
}

// installLinuxIcon writes the same PNG already embedded in the binary (see icon var)
// to disk. Linux desktop shells do not read icons from inside the ELF; Wails on GTK4
// also ignores Options.Icon for the window/taskbar. The .desktop Icon= path is required.
func installLinuxIcon(home string) {
	dir := filepath.Join(home, ".local", "share", "icons", "hicolor", "256x256", "apps")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}
	_ = os.WriteFile(linuxIconPath(home), icon, 0644)
}

func refreshLinuxDesktopDatabase(home string) {
	if upd, err := exec.LookPath("update-desktop-database"); err == nil {
		appDir := filepath.Join(home, ".local", "share", "applications")
		_ = exec.Command(upd, appDir).Run()
	}
}

func installLinux() {
	if runtime.GOOS != "linux" {
		return
	}

	for _, arg := range os.Args {
		if arg == "--debug" {
			return
		}
	}

	exe, err := os.Executable()
	if err != nil {
		return
	}

	if installedUnderUsr(exe) {
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	targetDir := filepath.Join(home, ".local", "bin")
	targetExe := filepath.Join(targetDir, "goexplore")

	// Already installed. Refresh the menu entry so it keeps the WebKit workaround.
	if exe == targetExe {
		installLinuxIcon(home)
		writeDesktopEntry(home, targetExe)
		refreshLinuxDesktopDatabase(home)
		return
	}

	// Ensure target directories exist
	os.MkdirAll(targetDir, 0755)

	// Copy the executable
	data, err := os.ReadFile(exe)
	if err == nil {
		os.WriteFile(targetExe, data, 0755)
	}

	installLinuxIcon(home)
	writeDesktopEntry(home, targetExe)
	refreshLinuxDesktopDatabase(home)

	// Execute the newly installed binary and exit the current running process
	cmd := exec.Command(targetExe)
	cmd.Start()
	os.Exit(0)
}

func installedUnderUsr(exe string) bool {
	candidates := []string{exe}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		candidates = append(candidates, resolved)
	}
	for _, path := range candidates {
		clean := filepath.Clean(path)
		if clean == "/usr" || strings.HasPrefix(clean, "/usr/") {
			return true
		}
	}
	return false
}

func writeDesktopEntry(home, targetExe string) {
	appDir := filepath.Join(home, ".local", "share", "applications")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return
	}
	iconPath := linuxIconPath(home)
	// env= on Exec avoids Plasma/WebKit NVIDIA explicit-sync black flash on Wayland.
	// Embedded icon (Options.Icon) is used on Windows/macOS. On Linux GTK4, Wails does
	// not set a window icon from bytes; Plasma reads Icon= from this desktop entry.
	desktopContent := fmt.Sprintf(`[Desktop Entry]
Version=1.0
Type=Application
Name=GoExplore
Comment=File manager for local and remote storage
Exec=env __NV_DISABLE_EXPLICIT_SYNC=1 %s
Icon=%s
Terminal=false
Categories=Utility;FileTools;
StartupNotify=false
StartupWMClass=%s
`, targetExe, iconPath, linuxApplicationID)
	_ = os.WriteFile(filepath.Join(appDir, "goexplore.desktop"), []byte(desktopContent), 0644)
}

// WebKitGTK enables Wayland explicit sync but does not set an acquire point.
// KWin then closes the connection, or the window flashes black, on NVIDIA.
// Force the workaround even if the desktop launcher already set the variable.
func applyWaylandNVIDIAWorkaround() {
	if runtime.GOOS != "linux" {
		return
	}
	if os.Getenv("XDG_SESSION_TYPE") != "wayland" && os.Getenv("WAYLAND_DISPLAY") == "" {
		return
	}
	if _, err := os.Stat("/sys/module/nvidia"); err != nil {
		return
	}
	_ = os.Setenv("__NV_DISABLE_EXPLICIT_SYNC", "1")
}

func main() {
	applyWaylandNVIDIAWorkaround()
	installLinux()

	appService := NewApp()

	wailsApp := application.New(application.Options{
		Name:        "GoExplore",
		Description: "Multi-protocol file explorer",
		Icon:        icon,
		Services: []application.Service{
			application.NewService(appService),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(assets),
		},
		Linux: application.LinuxOptions{
			ApplicationID: linuxApplicationID,
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "GoExplore",
		Width:            1024,
		Height:           768,
		BackgroundColour: application.NewRGB(10, 10, 10),
		URL:              "/",
	})

	if err := wailsApp.Run(); err != nil {
		println("Error:", err.Error())
	}
}
