// +build windows

package main

import (
	"github.com/hectane/go-acl"
	"os"
)

func chmod(file *os.File, mode os.FileMode) error {
	return acl.Chmod(file.Name(), 0755)
}
