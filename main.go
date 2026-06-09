package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

var Version string = "dev"

func installLinux() {
	if runtime.GOOS != "linux" {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	targetDir := filepath.Join(home, ".local", "bin")
	targetExe := filepath.Join(targetDir, "goexplore")

	// If we are already running from the target path, do nothing.
	if exe == targetExe {
		return
	}

	// Ensure target directories exist
	os.MkdirAll(targetDir, 0755)

	// Copy the executable
	data, err := os.ReadFile(exe)
	if err == nil {
		os.WriteFile(targetExe, data, 0755)
	}

	// Write the application icon
	iconDir := filepath.Join(home, ".local", "share", "icons")
	os.MkdirAll(iconDir, 0755)
	iconPath := filepath.Join(iconDir, "goexplore.png")
	os.WriteFile(iconPath, icon, 0644)

	// Write the desktop shortcut file
	appDir := filepath.Join(home, ".local", "share", "applications")
	os.MkdirAll(appDir, 0755)
	desktopContent := fmt.Sprintf(`[Desktop Entry]
Name=GoExplore
Exec=%s
Icon=%s
Type=Application
Terminal=false
Categories=Utility;FileTools;`, targetExe, iconPath)
	os.WriteFile(filepath.Join(appDir, "goexplore.desktop"), []byte(desktopContent), 0644)

	// Execute the newly installed binary and exit the current running process
	cmd := exec.Command(targetExe)
	cmd.Start()
	os.Exit(0)
}

func main() {
	installLinux()

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:            "GoExplore",
		Width:            1024,
		Height:           768,
		WindowStartState: options.Normal,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 10, G: 10, B: 10, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			Icon: icon,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
