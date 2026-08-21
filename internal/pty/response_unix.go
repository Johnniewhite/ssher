//go:build !windows

package pty

func formatPromptResponse(password string) string { return password + "\n" }
