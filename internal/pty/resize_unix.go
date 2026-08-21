//go:build !windows

package pty

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/x/xpty"
	"golang.org/x/term"
)

func watchTerminalResize(terminal *os.File, pseudoterminal xpty.Pty) func() {
	resize := func() {
		if width, height, err := term.GetSize(int(terminal.Fd())); err == nil {
			_ = pseudoterminal.Resize(width, height)
		}
	}
	changes := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(changes, syscall.SIGWINCH)
	go func() {
		for {
			select {
			case <-changes:
				resize()
			case <-done:
				return
			}
		}
	}()
	resize()
	return func() {
		signal.Stop(changes)
		close(done)
	}
}
