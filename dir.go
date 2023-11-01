package main

import (
	"cmp"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
    "slices"
)

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
