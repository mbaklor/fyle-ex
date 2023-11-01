package main

import (
	"io/fs"
	"os"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	sync.RWMutex
	app fyne.App
	win fyne.Window

	dir         string
	OnDirChange func(dir []fs.DirEntry)
}

func InitApp() *App {
	app := app.NewWithID("com.github.mbaklor.fyle-ex")
	win := app.NewWindow("Fyle-Ex")
	dir, err := os.UserHomeDir()
	if err != nil {
		println("no home??")
		dir = "/"
	}
	a := &App{app: app, win: win, dir: dir}
	a.populateWindow()
	return a
}

func (a *App) populateWindow() {
	path := widget.NewLabel(a.dir)
	filetree, _ := a.createFileList()
	mainView := container.NewBorder(path, nil, nil, nil, filetree)
	a.win.SetContent(mainView)

	a.OnDirChange = func(dir []fs.DirEntry) {
		a.UpdateList(filetree, dir)
		path.SetText(a.dir)
	}
}
