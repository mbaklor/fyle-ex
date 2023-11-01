package main

import (
	"fmt"
	"io/fs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/skratchdot/open-golang/open"
)

func (a *App) createFileList() (*KeybindList, error) {
	tree, err := a.GetFileTree(a.dir)
	if err != nil {
		return nil, err
	}
	lst := NewKeybindList(
		a.ListLength(tree),
		a.ListCreate(),
		a.ListUpdate(tree),
	)
	lst.OnSelected = a.ListSelect(tree)
	return lst, nil
}

func (a *App) UpdateList(lst *KeybindList, tree []fs.DirEntry) {
	lst.Length = a.ListLength(tree)
	lst.UpdateItem = a.ListUpdate(tree)
	lst.OnSelected = a.ListSelect(tree)
	lst.Refresh()
}

func (a *App) ListSelect(tree []fs.DirEntry) func(id int) {
	return func(id int) {
		t := tree[id]
		fmt.Println(t.Name())
		if t.IsDir() {
			a.ChangeDir(t.Name())
		} else {
			open.Start(a.GetAbsPath(t.Name()))
		}
	}
}

func (a *App) ListLength(tree []fs.DirEntry) func() int {
	return func() int {
		return len(tree)
	}
}

func (a *App) ListCreate() func() fyne.CanvasObject {
	return func() fyne.CanvasObject {
		return container.NewBorder(nil, nil, widget.NewIcon(theme.FolderIcon()), nil, widget.NewLabel("--"))
	}
}

func (a *App) ListUpdate(tree []fs.DirEntry) func(lii widget.ListItemID, co fyne.CanvasObject) {
	return func(lii widget.ListItemID, co fyne.CanvasObject) {
		t := tree[lii]
		if t.IsDir() {
			co.(*fyne.Container).Objects[1].(*widget.Icon).SetResource(theme.FolderIcon())
		} else {
			co.(*fyne.Container).Objects[1].(*widget.Icon).SetResource(theme.FileIcon())
		}
		co.(*fyne.Container).Objects[0].(*widget.Label).SetText(t.Name())
	}
}

func main() {
	println("STARTED")
	a := InitApp()

	a.win.ShowAndRun()
}
