package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"payload/internal/store"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	s, err := store.New()
	if err != nil {
		log.Fatalf("init store: %v", err)
	}
	app := NewApp(s)

	err = wails.Run(&options.App{
		Title:  "payload",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
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
