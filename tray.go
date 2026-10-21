package main

import (
	_ "embed"
	"log"
	"os"
	"os/exec"

	"github.com/gogpu/systray"
)

//go:embed assets/icons8.png
var iconPNG []byte

func openBrowser(url string) {
	if err := exec.Command("open", url).Start(); err != nil {
		log.Printf("open browser: %v", err)
	}
}

// runTray blocks until the tray's message loop ends. Call it last in main.
func runTray(url string) {
	tray := systray.New()

	menu := systray.NewMenu()
	menu.Add("Open Todo Manager", func() { openBrowser(url) })
	menu.AddSeparator()
	menu.Add("Quit", func() {
		tray.Remove()
		os.Exit(0) // section 2 replaces this with a graceful shutdown
	})

	tray.SetTemplateIcon(iconPNG). // macOS template image: adapts to light/dark
					SetTooltip("Todo Manager").
					SetMenu(menu)
	tray.Show()

	if err := tray.Run(); err != nil { // Run blocks and pumps the platform message loop
		log.Printf("tray: %v", err)
	}
}
