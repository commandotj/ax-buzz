//go:build !ios

package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"testing"
)

func TestModifyOptionsForIOSOnOtherPlatformsIsNoop(t *testing.T) {
	options := application.Options{DisableDefaultSignalHandler: true}
	modifyOptionsForIOS(&options)
	if !options.DisableDefaultSignalHandler {
		t.Fatal("non-iOS options were changed")
	}
}
