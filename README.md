<div align="center">
  <img src="build/appicon.png" alt="GoExplore" width="128" />

  # GoExplore

  GoExplore is a powerful, cross-platform file manager designed to simplify how you access and manage your files across various servers and cloud storage providers. Built with speed and modern aesthetics in mind, it provides a seamless unified experience for interacting with local files, cloud buckets, and remote file shares.

</div>

## ✨ Features

- **Multi-Protocol Support:** Seamlessly connect and manage files across multiple storage systems including S3, SMB/CIFS, SFTP, WebDAV, NFS, and your Local Filesystem.
- **Cross-Connection Transfers:** Easily transfer files and folders directly between completely different connections (e.g., from an SMB share straight to an S3 bucket).
- **Secure by Design:** Your connection credentials and secrets are kept safe by utilizing your operating system's native keychain instead of being stored in plain text.
- **Modern Interface:** A stunning, responsive dark-mode UI with robust multi-selection (Shift/Ctrl+Click) and sortable data columns.
- **Auto-Installation:** For Linux users, downloads seamlessly integrate themselves into your application launcher with a desktop icon upon first run.

## 🚀 Downloads

Download the latest version of GoExplore for your operating system:

- **Windows (64-bit)**
  [Download Installer (.exe)](https://apps.jdbnet.co.uk/goexplore/goexplore-windows-installer.exe)
  
- **Linux (AMD64)**
  [Download Binary](https://apps.jdbnet.co.uk/goexplore/goexplore-linux-amd64)

---

## 📖 Getting Started

### Windows
1. Download the **Windows Installer** from the link above.
2. Run the executable. It will automatically install GoExplore, add it to your Start menu, and ensure all necessary dependencies (like WebView2) are seamlessly installed.
3. Launch **GoExplore** from your Start menu!

### Linux
1. Download the **Linux binary** from the link above.
2. Give the file executable permissions via your terminal: `chmod +x goexplore-linux-amd64`
3. Run the binary.
4. **GoExplore** will automatically install itself onto your system, create a desktop shortcut, and place its icon in your application launcher for future use! You can safely delete the downloaded file afterward.

---

## 🛠 For Developers

GoExplore is built using [Wails v2](https://wails.io/) and Vanilla JS.

### Build Instructions

1. Ensure Go 1.26 and Node.js are installed.
2. Install the Wails v2 CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
3. Run the development server with live reload:
   ```bash
   wails dev -tags webkit2_41
   ```
4. To compile a production binary, use:
   ```bash
   wails build -tags webkit2_41
   ```
   *(The output will be placed in the `build/bin/` directory).*
