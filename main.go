package main

import (
	"embed"

	"EmailTools/internal/app"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	mailApp := app.NewMailApp()

	err := wails.Run(&options.App{
		Title:     "EmailTools - 多邮箱查看",
		Width:     1200,
		Height:    800,
		MinWidth:  900,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 245, G: 247, B: 250, A: 1},
		OnStartup:        mailApp.Startup,
		Bind: []interface{}{
			mailApp,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
