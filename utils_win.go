//go:build windows

package main

import "errors"

func ioctl(fd, op, arg uintptr) error {
	return errors.New("no tiocgwinsz on windows")
}
