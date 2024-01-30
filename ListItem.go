package main

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type DirListItem struct {
	widget.BaseWidget
	label *widget.Label
	icon  *widget.Icon

	container *fyne.Container

	lock        sync.RWMutex
	onDblTapped func()
	onTapped    func()
	tappedAt    int64
}

func NewKeybindItem(text string, icon *widget.Icon) *DirListItem {
	label := widget.NewLabel(text)
	ki := &DirListItem{label: label, icon: icon}
	ki.ExtendBaseWidget(ki)
	return ki
}

func (ki *DirListItem) CreateRenderer() fyne.WidgetRenderer {
	ki.container = container.NewBorder(nil, nil, ki.icon, nil, ki.label)
	return widget.NewSimpleRenderer(ki.container)
}

func (ki *DirListItem) SetIcon(r fyne.Resource) {
	ki.icon.SetResource(r)
	ki.Refresh()
}

func (ki *DirListItem) SetText(s string) {
	ki.label.SetText(s)
	ki.Refresh()
}

func (ki *DirListItem) SetOnTapped(tapped func()) {
	ki.lock.Lock()
	ki.onTapped = tapped
	ki.lock.Unlock()
}

func (ki *DirListItem) SetOnDoubleTapped(dblTapped func()) {
	ki.lock.Lock()
	ki.onDblTapped = dblTapped
	ki.lock.Unlock()
}

func (ki *DirListItem) Tapped(e *fyne.PointEvent) {
	ki.lock.RLock()
	tapped := ki.onTapped
	dblTapped := ki.onDblTapped
	ki.lock.RUnlock()
	prevTap := ki.tappedAt
	ki.tappedAt = time.Now().UnixMilli()

	if ki.tappedAt-prevTap < 300 {
		if dblTapped != nil {
			dblTapped()
		}
	} else {
		if tapped != nil {
			tapped()
		}
	}
}
