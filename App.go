package main

import (
	"io/fs"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type CmdEntry struct {
	widget.BaseWidget
	entry     *widget.Entry
	label     *widget.RichText
	container *fyne.Container
	OnHide    func()
}

func NewCmdEntry() *CmdEntry {
	ce := &CmdEntry{}
	ce.entry = widget.NewEntry()
	ce.label = widget.NewRichText()
	ce.ExtendBaseWidget(ce)
	return ce
}

func (ce *CmdEntry) CreateRenderer() fyne.WidgetRenderer {
	ce.container = container.NewBorder(nil, nil, ce.label, nil, ce.entry)
	return widget.NewSimpleRenderer(ce.container)
}

func (ce *CmdEntry) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyEscape:
		if !ce.Hidden {
			ce.SetOnChanged(nil)
			ce.SetOnSubmitted(nil)
			ce.Hide()
		}
		return
	default:
		ce.entry.TypedKey(key)
	}
}

func (ce *CmdEntry) moveCurser(cols int) {
	for i := 0; i < cols; i++ {
		ce.entry.TypedKey(&fyne.KeyEvent{Name: fyne.KeyRight})
	}
}

func (ce *CmdEntry) selectString(cols int) {
	entry := ce.entry
	entry.KeyDown(&fyne.KeyEvent{Name: desktop.KeyShiftLeft})
	ce.moveCurser(cols)
	entry.KeyUp(&fyne.KeyEvent{Name: desktop.KeyShiftLeft})
}

func (ce *CmdEntry) TypedRune(r rune) {
	ce.entry.TypedRune(r)
}

func (ce *CmdEntry) FocusGained() {
	ce.entry.FocusGained()
}

func (ce *CmdEntry) FocusLost() {
	ce.entry.FocusLost()
}

func (ce *CmdEntry) Hide() {
	ce.BaseWidget.Hide()
	if ce.OnHide != nil {
		ce.OnHide()
	}
}

func (ce *CmdEntry) SetOnSubmitted(f func(string)) {
	ce.entry.OnSubmitted = func(s string) {
		if f != nil {
			f(s)
		}
		ce.Hide()
	}
}

func (ce *CmdEntry) SetOnChanged(f func(string)) {
	ce.entry.OnChanged = f
}

func (ce *CmdEntry) SetText(text string) {
	ce.entry.SetText(text)
	ce.Refresh()
}

func (ce *CmdEntry) SetLabel(text string) {
	if text == "" {
		ce.label.Hide()
	} else {
		ce.label.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: text, Style: widget.RichTextStyle{SizeName: theme.SizeNameCaptionText}}}
		ce.label.Show()
	}
	ce.Refresh()
}

type App struct {
	sync.RWMutex
	app fyne.App
	win fyne.Window
}

func InitApp() *App {
	app := app.NewWithID("com.github.mbaklor.fyle-ex")
	win := app.NewWindow("Fyle-Ex")
	win.Resize(fyne.NewSize(800, 600))
	a := &App{app: app, win: win}
	a.populateWindow()
	return a
}

// func (a *App) setCanvasBindings() {
// 	c := a.win.Canvas()
// 	fmt.Println("adding keybinds")
// 	c.SetOnTypedKey(func(ke *fyne.KeyEvent) {
// 		if ke.Name == fyne.KeyEscape {
// 			if !a.cmdInput.Hidden {
// 				a.cmdInput.Hide()
// 				c.Refresh(a.win.Content())
// 			}
// 		}
// 	})
// 	c.SetOnTypedRune(func(r rune) {
// 		if r == ':' {
// 			if a.cmdInput.Hidden {
// 				a.cmdInput.Show()
// 				c.Focus(a.cmdInput)
// 				c.Refresh(a.win.Content())
// 			}
// 		}
// 	})
// }

func (a *App) createFileList() (*DirList, error) {
	ls, err := NewDirList(a.win.Canvas())
	if err != nil {
		return nil, err
	}
	return ls, nil
}

func (a *App) populateWindow() {
	center := container.NewStack()
	ls, err := a.createFileList()
	if err != nil {
		center.Add(container.NewCenter(widget.NewLabel(err.Error())))
	} else {
		center.Add(ls)
	}
	path := widget.NewLabel(ls.currentDir)
	mainView := container.NewBorder(nil, nil, nil, nil, center)
	a.win.SetContent(mainView)
	a.win.Canvas().Focus(ls)

	ls.OnDirChange = func(dir []fs.DirEntry) {
		path.SetText(ls.currentDir)
	}
}
