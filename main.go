package main

import (
	"cmp"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/skratchdot/open-golang/open"
)

type KeybindList struct {
	widget.List
}

func NewKeybindList(length func() int, createItem func() fyne.CanvasObject, updateItem func(widget.ListItemID, fyne.CanvasObject)) *KeybindList {
	kl := &KeybindList{}
	kl.Length = length
	kl.CreateItem = createItem
	kl.UpdateItem = updateItem
	kl.ExtendBaseWidget(kl)
	return kl
}

func (kl *KeybindList) UpdateList(list []any) {
}

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
	lst.Select(0)
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

func (a *App) GetAbsPath(rel string) string {
	return filepath.Join(a.dir, rel)
}

func (a *App) ChangeDir(location string) error {
	if filepath.IsAbs(location) {
		a.dir = location
	} else {
		path := a.GetAbsPath(location)
		if _, err := os.Stat(path); err == nil {
			a.dir = path
		}
	}
	tree, err := a.GetFileTree(a.dir)
	if err != nil {
		return err
	}
	if a.OnDirChange != nil {
		a.OnDirChange(tree)
	}

	return nil
}

func (a *App) GetFileTree(location string) ([]fs.DirEntry, error) {
	dir, err := os.ReadDir(location)
	if err != nil {
		return nil, err
	}
	tree := make([]fs.DirEntry, 0, len(dir))
	for _, t := range dir {
		protected, err := a.CheckFileProtected(t.Name())
		if err != nil {
			fmt.Println(err)
			continue
		}
		if !protected {
			tree = append(tree, t)
		}
	}
	a.SortTree(tree)
	// for _, t := range tree {
	// 	fmt.Println(t.IsDir(), t.Name())
	// }
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (a *App) SortTree(tree []fs.DirEntry) []fs.DirEntry {
	slices.SortFunc(tree, func(a, b fs.DirEntry) int {
		an, bn := 0, 0
		if !a.IsDir() {
			an = 1
		}
		if !b.IsDir() {
			bn = 1
		}
		if n := cmp.Compare(an, bn); n != 0 {
			return n
		}
		as, bs := strings.ToLower(a.Name()), strings.ToLower(b.Name())
		return cmp.Compare(as, bs)
	})
	return tree
}

func (a *App) CheckFileProtected(path string) (bool, error) {
	path, err := filepath.Abs(a.GetAbsPath(path))
	if err != nil {
		fmt.Println(err)
	}
	pointer, err := syscall.UTF16PtrFromString(`\\?\` + path)
	if err != nil {
		return false, err
	}
	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false, err
	}
	// 4 is a little magic number-y but it seems to be protected os files
	return attributes&4 != 0, nil

}

func main() {
	println("STARTED")
	a := InitApp()

	a.win.ShowAndRun()
}
