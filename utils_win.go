//go:build windows

package main

import "errors"

const tiocgwinsz = 0x5413

func ioctl(fd, op, arg uintptr) error {
	return errors.New("no tiocgwinsz on windows")
}
