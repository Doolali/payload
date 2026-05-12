package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"payload/internal/store"
)

//go:embed all:frontend/dist
var assets embed.FS

// version is overwritten at release-build time via -ldflags "-X main.version=...".
// Dev builds keep the literal "dev" so the updater knows not to compare.
var version = "dev"

func main() {
	s, err := store.New()
	if err != nil {
		log.Fatalf("init store: %v", err)
	}
	app := NewApp(s, version)

	appMenu := menu.NewMenu()
	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("Check for Updates...", nil, func(_ *menu.CallbackData) {
		if app.ctx == nil {
			return
		}
		wruntime.EventsEmit(app.ctx, "menu:check-for-updates")
	})

	err = wails.Run(&options.App{
		Title:  "payload",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		Menu:             appMenu,
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatalf("wails run: %v", err)
	}
}
