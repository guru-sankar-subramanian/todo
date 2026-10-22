package main

import (
	"embed"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/guru-sankar-subramanian/todo/internal/db"
	"github.com/guru-sankar-subramanian/todo/internal/server"
	"gopkg.in/natefinch/lumberjack.v2"
)

var DEFAULT_ADDR string

var DATA_DIRECTORY = defaultDataDir()
var isProd bool

//go:embed frontend
var embeddedUI embed.FS

func init() {
	runtime.LockOSThread()
}

func main() {
	initialize()
	database := db.StartDb(DATA_DIRECTORY)
	defer database.Db.Close()

	runTray("http://" + DEFAULT_ADDR) // blocks here
}

func initialize() {
	err := setupLogging()
	if err != nil {
		log.Fatal(err)
	}

	setupDefaultConfigs()
	startHttpServer()
}

func startHttpServer() {
	srv := &http.Server{Addr: DEFAULT_ADDR, Handler: server.Init(isProd, embeddedUI)}
	go func() {
		log.Printf("listening on http://%s", DEFAULT_ADDR)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err) // e.g. port already in use
		}
	}()
}

func setupLogging() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	logDir := filepath.Join(home, "Library", "Logs", "todo-manager")
	log.Println("log files are in : ", logDir)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return err
	}

	rotator := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "app.log"),
		MaxSize:    5,    // MB per file
		MaxBackups: 5,    // old files kept
		MaxAge:     30,   // days
		Compress:   true, // gzip old files
	}

	// In a terminal you still see logs; under launchd stdout goes nowhere useful.
	log.SetOutput(io.MultiWriter(os.Stderr, rotator))
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return nil
}

func setupDefaultConfigs() {
	if insideAppBundle() {
		isProd = true
	} else {
		isProd = false
	}

	if isProd {
		DEFAULT_ADDR = "127.0.0.1:8421"
	} else {
		DEFAULT_ADDR = "127.0.0.1:8422"
	}
}

func defaultDataDir() string {
	base, err := os.UserConfigDir() // ~/Library/Application Support on macOS
	if err != nil {
		log.Fatal(err)
	}
	dir := filepath.Join(base, "todo-manager")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	return dir
}

func insideAppBundle() bool {
	exe, err := os.Executable()
	return err == nil && strings.Contains(exe, ".app/Contents/MacOS/")
}
