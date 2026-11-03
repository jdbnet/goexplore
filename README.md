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

- **Debian and Ubuntu**

  ```bash
  curl -fsSL https://apt.jdbnet.co.uk/install/stable.sh | sudo bash
  sudo apt install goexplore
  ```

- **Windows (64-bit)**
  [Download Installer (.exe)](https://github.com/jdbnet/goexplore/releases/latest/download/goexplore-windows-installer.exe)

- **Linux (AMD64 binary)**
  [Download Binary](https://github.com/jdbnet/goexplore/releases/latest/download/goexplore-linux-amd64)

---

## 📖 Getting Started

### Windows
1. Download the **Windows Installer** from the link above.
2. Run the executable. It will automatically install GoExplore, add it to your Start menu, and ensure all necessary dependencies (like WebView2) are seamlessly installed.
3. Launch **GoExplore** from your Start menu!

### Linux
Install from the apt repository:

```bash
curl -fsSL https://apt.jdbnet.co.uk/install/stable.sh | sudo bash
sudo apt install goexplore
```

The package installs GoExplore to `/usr/bin/goexplore` and adds a desktop entry.

Or download the binary from [GitHub Releases](https://github.com/jdbnet/goexplore/releases/latest):

1. Download `goexplore-linux-amd64`.
2. `chmod +x goexplore-linux-amd64`
3. Run the binary once. It copies itself to `~/.local/bin/goexplore`, installs icons and a menu entry under `~/.local`, then re-launches from there.
4. You can delete the downloaded file afterward; use the application menu or `~/.local/bin/goexplore`.

---

## 🛠 For Developers

GoExplore is built using [Wails v3](https://v3.wails.io/) and Vanilla JS.

### Build Instructions

1. Ensure Go 1.27.3 and Node.js are installed.
2. Install the Wails v3 CLI: `go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0`
3. Install Linux build deps (GTK 4, WebKitGTK 6): e.g. on Debian/Ubuntu, `libgtk-4-dev` and `libwebkitgtk-6.0-dev`.
4. Generate build assets once (if `build/Taskfile.yml` is missing): `wails3 update build-assets -name GoExplore -binaryname goexplore -config ./wails.json -dir build`
5. App icon source: `frontend/src/assets/appicon.png` (copied to `build/appicon.png` for Wails packaging; CI regenerates `windows/icon.ico` from it).
6. Run the development server with live reload:
   ```bash
   wails3 dev
   ```
7. To compile a production binary, use:
   ```bash
   wails3 build
   ```
   *(Binaries are written to the `bin/` directory.)*
