package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	cliService := NewCLIService()
	app := application.New(application.Options{
		Name:        "Buzz",
		Description: "Mac software discovery and management",
		Services: []application.Service{
			application.NewService(cliService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})
	cliService.SetApp(app)

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Buzz Installer",
		Width:            1160,
		Height:           760,
		BackgroundColour: application.NewRGB(244, 243, 239),
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
