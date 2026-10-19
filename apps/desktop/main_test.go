package main

import (
	"errors"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestMainDispatchesAndReportsFailure(t *testing.T) {
	oldCreate, oldRun, oldFatal := createDesktopApp, startWailsApp, fatalError
	defer func() { createDesktopApp, startWailsApp, fatalError = oldCreate, oldRun, oldFatal }()
	app := &application.App{}
	service := NewCLIService()
	createDesktopApp = func() (*application.App, *CLIService) { return app, service }
	var gotFatal error
	startWailsApp = func(got *application.App) error {
		if got != app {
			t.Fatal("runner received different app")
		}
		if service.app != app {
			t.Fatal("CLI service not bound to app")
		}
		return errors.New("run failed")
	}
	fatalError = func(values ...any) { gotFatal = values[0].(error) }
	main()
	if gotFatal == nil || gotFatal.Error() != "run failed" {
		t.Fatalf("fatal error = %v", gotFatal)
	}
}

func TestRunDesktopSuccess(t *testing.T) {
	app := &application.App{}
	service := NewCLIService()
	called := false
	err := runDesktop(func() (*application.App, *CLIService) { return app, service }, func(got *application.App) error {
		called = true
		if got != app || service.app != app {
			t.Fatal("app binding missing")
		}
		return nil
	})
	if err != nil || !called {
		t.Fatalf("runDesktop() = %v, called=%t", err, called)
	}
}

func TestRunWailsAppDelegates(t *testing.T) {
	old := invokeWailsRun
	defer func() { invokeWailsRun = old }()
	app := &application.App{}
	called := false
	want := errors.New("expected")
	invokeWailsRun = func(got *application.App) error { called = got == app; return want }
	if got := runWailsApp(app); !errors.Is(got, want) || !called {
		t.Fatalf("runWailsApp() = %v, called=%t", got, called)
	}
}

func TestNewBuzzAppConfiguration(t *testing.T) {
	oldNew, oldWindow := newWailsApplication, createWailsWindow
	oldRestore, oldShow, oldFocus := restoreWailsWindow, showWailsWindow, focusWailsWindow
	defer func() {
		newWailsApplication, createWailsWindow = oldNew, oldWindow
		restoreWailsWindow, showWailsWindow, focusWailsWindow = oldRestore, oldShow, oldFocus
	}()
	var options application.Options
	var windowOptions application.WebviewWindowOptions
	app := &application.App{}
	newWailsApplication = func(got application.Options) *application.App {
		options = got
		got.SingleInstance.OnSecondInstanceLaunch(application.SecondInstanceData{})
		return app
	}
	window := &application.WebviewWindow{}
	createWailsWindow = func(_ *application.WindowManager, opts application.WebviewWindowOptions) *application.WebviewWindow {
		windowOptions = opts
		return window
	}
	var restored, shown, focused int
	restoreWailsWindow = func(got *application.WebviewWindow) {
		if got == window {
			restored++
		}
	}
	showWailsWindow = func(got *application.WebviewWindow) application.Window {
		if got == window {
			shown++
		}
		return nil
	}
	focusWailsWindow = func(got *application.WebviewWindow) {
		if got == window {
			focused++
		}
	}
	app, service := newBuzzApp()
	if app == nil || service == nil || service.app != app || options.SingleInstance == nil {
		t.Fatal("newBuzzApp returned unbound app service")
	}
	options.SingleInstance.OnSecondInstanceLaunch(application.SecondInstanceData{})
	if restored != 1 || shown != 1 || focused != 1 {
		t.Fatalf("window activation calls = %d/%d/%d", restored, shown, focused)
	}
	if windowOptions.Title != "Buzz" || windowOptions.Width != 1160 || windowOptions.Height != 760 || windowOptions.URL != "/" {
		t.Fatalf("window options = %#v", windowOptions)
	}
}
