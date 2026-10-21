# Todo Manager

A tiny personal todo app that lives in the macOS menu bar. A Go backend serves an
HTML + JS UI on `localhost` and stores tasks in a local SQLite database.
Everything ships as one self-contained app: the UI is embedded in the binary.

- Menu-bar only (no Dock icon): **Open Todo Manager** and **Quit**
- Go backend with the UI embedded (`go:embed`)
- SQLite storage (`modernc.org/sqlite`, pure Go, no cgo)
- Menu-bar icon via `gogpu/systray` (pure Go, no cgo)
- Rotating log files
- Apple Silicon (arm64), macOS 11 or newer

## Install (from a zip)

1. Unzip `TodoManager-<version>.zip`.
2. Drag `TodoManager.app` into `/Applications`.
3. Open it for the first time (see "First launch" below).
4. Look for the icon in the menu bar, click it, and choose **Open Todo Manager**.
   The UI is at <http://127.0.0.1:8421>.

Optional: check the download first.

```bash
shasum -a 256 TodoManager-<version>.zip     # compare with the checksum you were sent
```

### First launch

The app is ad-hoc signed, not notarized by Apple, so macOS blocks the first open.
Use either way:

- **Settings route:** try to open the app once, then go to
  System Settings -> Privacy & Security, scroll down, and click **Open Anyway**.
  (On older macOS you can instead right-click the app -> Open -> Open.)
- **Terminal route:** remove the quarantine flag, then open it normally:

  ```bash
  xattr -dr com.apple.quarantine /Applications/TodoManager.app 
  ```
  then try running the app. 

### Start at login

Not built in yet. To do it manually: System Settings -> General -> Login Items ->
add `TodoManager` under "Open at Login".

## Where things live

| What | Location |
|---|---|
| Database | `~/Library/Application Support/todo-manager/data.db` |
| Logs | `~/Library/Logs/todo-manager/` |
| UI | embedded in the app (nothing to install) |
| Address | `http://127.0.0.1:8421` (this Mac only) |

The server listens on `127.0.0.1` only, so nothing on your network can reach it.

## Build from source

Requirements: macOS on Apple Silicon, Go 1.25 or newer, Xcode command line tools
(for `codesign` and `ditto`).

```bash
git clone <repo-url> todo
cd todo
./build.sh 1.0.0
```

Output:

- `build/TodoManager.app`
- `build/TodoManager-1.0.0.zip` (the file to share) and its SHA-256 checksum

For Intel Macs, change `GOARCH=arm64` to `amd64` in `build.sh`.

## Development

```bash
go run .
```

When run outside an `.app` bundle the app is in dev mode and serves the UI from
`./frontend` on disk, so you can edit HTML/JS and just refresh. Inside the bundle
it always uses the embedded copy. If the installed app is running, quit it first
when both would use the same port.

### Project layout:

```
main.go        startup, logging, database, tray hand-off
server.go      HTTP routes (UI + /api/tasks, /api/health)
tray.go        menu-bar icon and menu
frontend/      HTML, JS, CSS (embedded at build time)
assets/        tray icon (icon.png)
build.sh       builds, signs and zips TodoManager.app
```

## Uninstall

1. Click the menu-bar icon -> **Quit**.
2. Delete `/Applications/TodoManager.app`.
3. Optional, removes your data and logs:

   ```bash
   rm -rf ~/Library/Application\ Support/todo-manager ~/Library/Logs/todo-manager
   ```

## Troubleshooting

- **No menu-bar icon:** run it from a terminal to see errors:
  `/Applications/TodoManager.app/Contents/MacOS/TodoManager`
- **"Address already in use":** another copy holds the port. Check with
  `lsof -i :8421`, quit it, and relaunch.
- **404 page:** the server is up but cannot find the UI. Make sure you are running
  the built app (not an old copy), and quit and reopen it after updating.
- **"App is damaged" or will not open:** run the `xattr` command from "First launch".