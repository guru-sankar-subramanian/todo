#!/usr/bin/env bash
# Builds TodoManager.app (Apple Silicon) and zips it for sharing.
# Usage: ./build.sh [version]     e.g. ./build.sh 1.0.1
set -euo pipefail
cd "$(dirname "$0")"

APP="TodoManager"
DISPLAY_NAME="Todo Manager"
VERSION="${1:-1.0.0}"
BUNDLE_ID="dev.guru.todomanager"
OUT="build/${APP}.app"
ZIP="build/${APP}-${VERSION}.zip"

# --- preflight ---------------------------------------------------------------
for tool in go codesign ditto shasum plutil; do
  command -v "$tool" >/dev/null || { echo "❌ missing tool: $tool"; exit 1; }
done
[ -d frontend ]        || { echo "❌ ./frontend not found (go:embed needs it)"; exit 1; }
[ -f assets/icons8.png ] || { echo "❌ ./assets/icons8.png not found (tray icon)"; exit 1; }

# --- build -------------------------------------------------------------------
rm -rf build
mkdir -p "${OUT}/Contents/MacOS"

echo "→ go vet"
go vet ./...

echo "→ go build"
# gogpu/systray is pure Go, so no cgo and no Xcode tools are needed.
# The UI is embedded, so the result is one self-contained binary.
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
  go build -trimpath -ldflags "-s -w" \
  -o "${OUT}/Contents/MacOS/${APP}" .

# --- bundle ------------------------------------------------------------------
cat > "${OUT}/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleName</key>               <string>${APP}</string>
    <key>CFBundleDisplayName</key>        <string>${DISPLAY_NAME}</string>
    <key>CFBundleIdentifier</key>         <string>${BUNDLE_ID}</string>
    <key>CFBundleExecutable</key>         <string>${APP}</string>
    <key>CFBundlePackageType</key>        <string>APPL</string>
    <key>CFBundleShortVersionString</key> <string>${VERSION}</string>
    <key>CFBundleVersion</key>            <string>${VERSION}</string>
    <key>LSMinimumSystemVersion</key>     <string>11.0</string>
    <key>NSHighResolutionCapable</key>    <true/>
    <key>LSUIElement</key>                <true/>
</dict>
</plist>
EOF
plutil -lint "${OUT}/Contents/Info.plist" >/dev/null

# --- sign + verify -----------------------------------------------------------
# Ad-hoc signature (free): seals Info.plist together with the binary.
codesign --force --deep --sign - "${OUT}"
codesign --verify --deep --strict "${OUT}"

# --- package -----------------------------------------------------------------
# ditto keeps the app's metadata intact; plain zip can break it.
ditto -c -k --keepParent "${OUT}" "${ZIP}"

echo
echo "✅ Built:    ${OUT}"
echo "   Zip:      ${ZIP}"
echo "   Version:  ${VERSION}"
echo "   Checksum: $(shasum -a 256 "${ZIP}" | cut -d' ' -f1)"