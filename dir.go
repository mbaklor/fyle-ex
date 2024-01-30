package main

import (
	"cmp"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
)

func (dl *DirList) GetAbsPath(rel string) string {
	return filepath.Join(dl.currentDir, rel)
}

func (dl *DirList) ChangeDir(location string) (err error) {
	if filepath.IsAbs(location) {
		dl.currentDir = location
	} else {
		path := dl.GetAbsPath(location)
		if _, err := os.Stat(path); err == nil {
			dl.currentDir = path
		}
	}
	dl.tree, err = dl.GetFileTree(dl.currentDir)
	if err != nil {
		return err
	}
	dl.topLabel.SetText(dl.currentDir)
	dl.Refresh()
	if val, ok := dl.dirCache[dl.currentDir]; ok && val != -1 {
		dl.list.Select(val)
		dl.selected = val
	} else {
		dl.list.UnselectAll()
		dl.selected = -1
		dl.dirCache[dl.currentDir] = -1
	}
	if dl.OnDirChange != nil {
		dl.OnDirChange(dl.tree)
	}
	return nil
}

func (dl *DirList) GetFileTree(location string) ([]fs.DirEntry, error) {
	dir, err := os.ReadDir(location)
	if err != nil {
		return nil, err
	}
	tree := make([]fs.DirEntry, 0, len(dir))
	for _, t := range dir {
		protected, err := dl.CheckFileProtected(t.Name())
		if err != nil {
			fmt.Println("get tree:", err)
			continue
		}
		if !protected {
			tree = append(tree, t)
		}
	}
	dl.SortTree(tree)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (dl *DirList) SortTree(tree []fs.DirEntry) []fs.DirEntry {
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

func (dl *DirList) CheckFileProtected(path string) (bool, error) {
	path, err := filepath.Abs(dl.GetAbsPath(path))
	if err != nil {
		fmt.Println("protected:", err)
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
