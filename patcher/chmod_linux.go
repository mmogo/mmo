// +build linux

package main

import "os"

func chmod(file *os.File, mode os.FileMode) error {
	return file.Chmod(mode)
}
