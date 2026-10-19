package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := runDesktop(createDesktopApp, startWailsApp); err != nil {
		fatalError(err)
	}
}

var fatalError = log.Fatal
var createDesktopApp = newBuzzApp
var startWailsApp = runWailsApp
var newWailsApplication = application.New
var createWailsWindow = (*application.WindowManager).NewWithOptions
var restoreWailsWindow = (*application.WebviewWindow).Restore
var showWailsWindow = (*application.WebviewWindow).Show
var focusWailsWindow = (*application.WebviewWindow).Focus
var invokeWailsRun = (*application.App).Run

func runDesktop(create func() (*application.App, *CLIService), run func(*application.App) error) error {
	app, cliService := create()
	cliService.SetApp(app)
	return run(app)
}

func newBuzzApp() (*application.App, *CLIService) {
	cliService := NewCLIService()
	var mainWindow *application.WebviewWindow
	app := newWailsApplication(application.Options{
		Name:        "Buzz",
		Description: "Mac software discovery and management",
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "com.commandotj.buzz",
			OnSecondInstanceLaunch: func(application.SecondInstanceData) {
				if mainWindow == nil {
					return
				}
				restoreWailsWindow(mainWindow)
				showWailsWindow(mainWindow)
				focusWailsWindow(mainWindow)
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

	mainWindow = createWailsWindow(app.Window, application.WebviewWindowOptions{
		Title:            "Buzz",
		Width:            1160,
		Height:           760,
		BackgroundColour: application.NewRGB(244, 243, 239),
		Mac:              application.MacWindow{TitleBar: application.MacTitleBarHidden},
		URL:              "/",
	})
	return app, cliService
}

func runWailsApp(app *application.App) error {
	return invokeWailsRun(app)
}
