package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

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
		iconPath := filepath.Join(home, ".local", "share", "icons", "goexplore.png")
		writeDesktopEntry(targetExe, iconPath)
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

	writeDesktopEntry(targetExe, iconPath)

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

func writeDesktopEntry(targetExe, iconPath string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	appDir := filepath.Join(home, ".local", "share", "applications")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return
	}
	// env= runs before the process starts. A menu launch otherwise inherits Plasma's
	// NVIDIA explicit-sync setting, and WebKit's DMA-BUF renderer flashes black.
	desktopContent := fmt.Sprintf(`[Desktop Entry]
Name=GoExplore
Exec=env __NV_DISABLE_EXPLICIT_SYNC=1 %s
Icon=%s
Type=Application
Terminal=false
Categories=Utility;FileTools;
StartupNotify=false
`, targetExe, iconPath)
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
			// A non-nil Linux options value defaults GPU policy to OnDemand.
			// On NVIDIA Wayland that accelerated WebKit surface flashes black.
			WebviewGpuPolicy: linux.WebviewGpuPolicyNever,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
