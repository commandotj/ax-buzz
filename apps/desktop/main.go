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
	var mainWindow *application.WebviewWindow
	app := application.New(application.Options{
		Name:        "Buzz",
		Description: "Mac software discovery and management",
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "com.commandotj.buzz",
			OnSecondInstanceLaunch: func(application.SecondInstanceData) {
				if mainWindow == nil {
					return
				}
				mainWindow.Restore()
				mainWindow.Show()
				mainWindow.Focus()
			},
		},
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

	mainWindow = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Buzz",
		Width:            1160,
		Height:           760,
		BackgroundColour: application.NewRGB(244, 243, 239),
		Mac:              application.MacWindow{TitleBar: application.MacTitleBarHidden},
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
