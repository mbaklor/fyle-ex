package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/skratchdot/open-golang/open"
)

type DirList struct {
	widget.BaseWidget
	list      *widget.List
	cmdInput  *CmdEntry
	topLabel  *widget.Label
	container *fyne.Container

	canvas fyne.Canvas

	currentDir string
	dirCache   map[string]int
	selected   widget.ListItemID
	tree       []fs.DirEntry
	search     string

	OnDirChange func(dir []fs.DirEntry)
	onKey       func(*fyne.KeyEvent)
	onRune      func(rune)
}

func NewDirList(canvas fyne.Canvas) (*DirList, error) {
	dl := &DirList{canvas: canvas}

	dir, err := os.UserHomeDir()
	if err != nil {
		dir = "/"
	}
	dl.currentDir = dir
	dl.dirCache = make(map[string]int)
	dl.selected = -1
	dl.dirCache[dl.currentDir] = -1

	dl.topLabel = widget.NewLabel(dir)

	dl.cmdInput = NewCmdEntry()
	dl.cmdInput.Hide()
	dl.cmdInput.OnHide = func() {
		dl.cmdInput.SetText("")
		dl.cmdInput.SetLabel("")
		if dl.canvas != nil {
			dl.canvas.Focus(dl)
		}
		dl.cmdInput.SetOnSubmitted(nil)
		dl.Refresh()
	}

	dl.tree, err = dl.GetFileTree(dir)
	if err != nil {
		return nil, err
	}
	dl.list = widget.NewList(
		func() int {
			return len(dl.tree)
		},
		func() fyne.CanvasObject {
			return NewKeybindItem("--", widget.NewIcon(theme.FolderIcon()))
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			item, is := co.(*DirListItem)
			if is {
				t := dl.tree[lii]
				item.SetOnTapped(func() {
					dl.list.Select(lii)
				})
				if t.IsDir() {
					item.SetOnDoubleTapped(func() {
						dl.ChangeDir(t.Name())
					})
					item.SetIcon(theme.FolderIcon())
				} else {
					item.SetOnDoubleTapped(func() {
						open.Start(dl.GetAbsPath(t.Name()))
					})
					item.SetIcon(theme.FileIcon())
				}
				item.SetText(t.Name())
			}
		})
	dl.list.OnSelected = func(id widget.ListItemID) {
		dl.selected = id
		dl.dirCache[dl.currentDir] = id
	}

	dl.ExtendBaseWidget(dl)
	return dl, nil
}

func (dl *DirList) CreateRenderer() fyne.WidgetRenderer {
	dl.container = container.NewBorder(dl.topLabel, dl.cmdInput, nil, nil, dl.list)
	return widget.NewSimpleRenderer(dl.container)
}

func (dl *DirList) TypedKey(event *fyne.KeyEvent) {
	switch event.Name {
	case fyne.KeyReturn, fyne.KeyEnter:
		if dl.selected != -1 {
			dl.ListEnter(dl.tree[dl.selected])
		}
	case fyne.KeyBackspace:
		dl.ListBack()
	case fyne.KeyDown:
		dl.IncSelection()
	case fyne.KeyUp:
		dl.DecSelection()
	case fyne.KeyF2:
		dl.typedRename()
	default:
		dl.list.TypedKey(event)
	}
}

func (dl *DirList) TypedRune(r rune) {
	if dl.onRune != nil {
		dl.onRune(r)
	}
	switch r {
	case 'k':
		dl.TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
	case 'j':
		dl.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
	case 'l':
		dl.TypedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
	case 'h':
		dl.TypedKey(&fyne.KeyEvent{Name: fyne.KeyBackspace})
	// case 'q':
	// 	fyne.CurrentApp().Quit()
	case ':':
		dl.SetCmdMode()
	case 'r':
		dl.typedRename()
	case '/':
		dl.typedSearch()
	case 'n':
		dl.searchString(dl.search, true)
	}
}

func (dl *DirList) FocusGained() {
}

func (dl *DirList) FocusLost() {
}

func (dl *DirList) showCmd(text, label string) {
	dl.cmdInput.SetLabel(label)
	dl.cmdInput.SetText(text)
	dl.cmdInput.Show()
	dl.canvas.Focus(dl.cmdInput)
	dl.Refresh()
}

func (dl *DirList) SetCmdMode() {
	dl.showCmd(":", "")
	dl.cmdInput.moveCurser(1)
	dl.cmdInput.SetOnSubmitted(func(s string) {
		sp := strings.Split(s, " ")
		switch sp[0] {
		case ":q", ":quit":
			fyne.CurrentApp().Quit()
		case ":cd", ":cwd":

		}
	})
}

func (dl *DirList) typedSearch() {
	dl.showCmd("/"+dl.search, "search:")
	dl.cmdInput.moveCurser(1)
	dl.cmdInput.selectString(len(dl.search))
	dl.cmdInput.SetOnChanged(func(s string) {
		dl.searchString(s[1:], false)
	})
	dl.cmdInput.SetOnSubmitted(func(s string) {
		dl.search = s[1:]
		dl.cmdInput.SetOnChanged(nil)
		dl.cmdInput.SetOnSubmitted(nil)
	})
}

func (dl *DirList) searchString(s string, findNext bool) {
	if s == "" {
		return
	}
	search := strings.ToLower(s)
	if !findNext && dl.selected != -1 {
		if strings.Contains(strings.ToLower(dl.tree[dl.selected].Name()), search) {
			return
		}
	}
	ids := make([]int, 0, len(dl.tree))
	for i, file := range dl.tree {
		if strings.Contains(strings.ToLower(file.Name()), search) {
			ids = append(ids, i)
		}
	}
	for _, id := range ids {
		if dl.selected < id {
			dl.list.Select(id)
			return
		}
	}
	if len(ids) > 0 {
		dl.list.Select(ids[0])
	}
}

func (dl *DirList) typedRename() {
	if dl.selected == -1 {
		return
	}
	current := dl.tree[dl.selected].Name()
	dl.showCmd(current, "rename:")
	idx := strings.LastIndex(current, ".")
	if idx == -1 {
		idx = len(current)
	}
	dl.cmdInput.selectString(idx)
	dl.cmdInput.SetOnSubmitted(func(s string) {
		dl.renameItem(dl.selected, s)
		dl.cmdInput.SetOnSubmitted(nil)
		dl.ChangeDir(dl.currentDir)
	})
}

func (dl *DirList) IncSelection() {
	if dl.selected != dl.list.Length()-1 {
		dl.list.Select(dl.selected + 1)
	} else {
		dl.list.Select(0)
	}
}

func (dl *DirList) DecSelection() {
	if dl.selected <= 0 {
		dl.list.Select(dl.list.Length() - 1)
	} else {
		dl.list.Select(dl.selected - 1)
	}
}

func (dl *DirList) renameItem(id int, newName string) {
	path := dl.GetAbsPath(dl.tree[id].Name())
	os.Rename(path, dl.GetAbsPath(newName))
}

func (dl *DirList) ListEnter(t fs.DirEntry) {
	if t.IsDir() {
		dl.ChangeDir(t.Name())
	} else {
		open.Start(dl.GetAbsPath(t.Name()))
	}
}

func (dl *DirList) ListBack() {
	dir := filepath.Dir(dl.currentDir)
	current := filepath.Base(dl.currentDir)
	dl.ChangeDir(dir)
	if dl.selected == -1 {
		for i, file := range dl.tree {
			if current == file.Name() {
				dl.selected = i
				dl.list.Select(i)
			}
		}
	}
}
