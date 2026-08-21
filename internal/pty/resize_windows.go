//go:build windows

package pty

import (
	"os"
	"time"

	"github.com/charmbracelet/x/xpty"
	"golang.org/x/term"
)

func watchTerminalResize(terminal *os.File, pseudoterminal xpty.Pty) func() {
	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		lastWidth, lastHeight := 0, 0
		for {
			select {
			case <-ticker.C:
				width, height, err := term.GetSize(int(terminal.Fd()))
				if err == nil && (width != lastWidth || height != lastHeight) {
					_ = pseudoterminal.Resize(width, height)
					lastWidth, lastHeight = width, height
				}
			case <-done:
				return
			}
		}
	}()
	return func() {
		close(done)
		<-stopped
	}
}
