package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
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
