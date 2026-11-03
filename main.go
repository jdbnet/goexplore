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
	linuxApplicationID  = "uk.co.jdbnet.goexplore"
	linuxIconThemeName  = "uk.co.jdbnet.goexplore"
	linuxIconLegacyName = "goexplore"
)

func linuxIconPath(home string) string {
	return filepath.Join(home, ".local", "share", "icons", "hicolor", "256x256", "apps", linuxIconThemeName+".png")
}

func installLinuxIcon(home string) {
	for _, size := range []string{"48x48", "128x128", "256x256"} {
		dir := filepath.Join(home, ".local", "share", "icons", "hicolor", size, "apps")
		if err := os.MkdirAll(dir, 0755); err != nil {
			continue
		}
		for _, name := range []string{linuxIconThemeName, linuxIconLegacyName} {
			_ = os.WriteFile(filepath.Join(dir, name+".png"), icon, 0644)
		}
	}
	if cache, err := exec.LookPath("gtk-update-icon-cache"); err == nil {
		hicolor := filepath.Join(home, ".local", "share", "icons", "hicolor")
		_ = exec.Command(cache, "-f", "-t", hicolor).Run()
	}
}

func refreshLinuxDesktopDatabase(home string) {
	if upd, err := exec.LookPath("update-desktop-database"); err == nil {
		appDir := filepath.Join(home, ".local", "share", "applications")
		_ = exec.Command(upd, appDir).Run()
	}
}

// Sidecar launcher for the release binary (e.g. ~/Downloads). KDE on Wayland does not
// use hicolor icons for raw executables; it needs a .desktop with an absolute Icon path.
func writePortableDesktop(sourceExe string) {
	dir := filepath.Dir(sourceExe)
	iconFile := filepath.Join(dir, "goexplore.png")
	_ = os.WriteFile(iconFile, icon, 0644)
	desktopPath := filepath.Join(dir, "GoExplore.desktop")
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
`, sourceExe, iconFile, linuxApplicationID)
	_ = os.WriteFile(desktopPath, []byte(desktopContent), 0644)
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

	writePortableDesktop(exe)

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
	// env= runs before the process starts. A menu launch otherwise inherits Plasma's
	// NVIDIA explicit-sync setting, and WebKit's DMA-BUF renderer flashes black.
	// GTK4 taskbar icons come from the .desktop file (absolute Icon= is most reliable
	// on KDE) and StartupWMClass matching the Wayland app id, not Options.Icon.
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
