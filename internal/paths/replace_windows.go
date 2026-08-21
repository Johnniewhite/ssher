//go:build windows

package paths

import "golang.org/x/sys/windows"

func replaceFile(source, target string) error { return windows.Rename(source, target) }
