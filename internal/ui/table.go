package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"github.com/johnniewhite/ssher/internal/store"
)

// TerminalWidth returns the current terminal width, defaulting to 100 when
// stdout isn't a TTY (so piped output still wraps reasonably).
func TerminalWidth() int {
	w, _, err := term.GetSize(0)
	if err != nil || w <= 0 {
		return 100
	}
	return w
}

// RenderServerTable produces a sorted, numbered, lipgloss-styled table of
// servers. Auto-switches to two-column layout on terminals >= 140 columns,
// matching the Python tool's behaviour.
func RenderServerTable(servers []store.Server) string {
	if len(servers) == 0 {
		return Muted.Render("  (no servers yet -- run `ssher add` to add one)")
	}

	rows := make([]string, len(servers))
	for i, s := range servers {
		rows[i] = renderRow(i+1, s)
	}

	if TerminalWidth() >= 140 && len(rows) >= 4 {
		return renderTwoColumn(rows)
	}
	return strings.Join(rows, "\n")
}

func renderRow(n int, s store.Server) string {
	star := "  "
	if s.IsFavorite {
		star = Favorite.Render("* ")
	}
	idx := Subtle.Render(fmt.Sprintf("%3d.", n))
	name := Title.Render(padRight(s.Name, 20))
	host := s.Host
	if s.Port != 0 && s.Port != 22 {
		host = fmt.Sprintf("%s:%d", s.Host, s.Port)
	}
	target := fmt.Sprintf("%s@%s", s.User, host)
	auth := Muted.Render(fmt.Sprintf("(%s)", string(s.AuthType)))
	tail := []string{}
	if s.Group != "" && s.Group != "default" {
		tail = append(tail, Accent.Render("#"+s.Group))
	}
	if len(s.Tags) > 0 {
		tail = append(tail, Muted.Render(strings.Join(s.Tags, ",")))
	}
	right := strings.Join(tail, " ")
	return fmt.Sprintf("%s%s %s %s %s %s", star, idx, name, target, auth, right)
}

func renderTwoColumn(rows []string) string {
	half := (len(rows) + 1) / 2
	left := rows[:half]
	right := rows[half:]
	var b strings.Builder
	colWidth := lipgloss.Width(longest(left)) + 4
	for i := 0; i < half; i++ {
		l := left[i]
		l = padDisplay(l, colWidth)
		var r string
		if i < len(right) {
			r = right[i]
		}
		b.WriteString(l)
		b.WriteString(r)
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

func padDisplay(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

func longest(xs []string) string {
	var best string
	for _, x := range xs {
		if lipgloss.Width(x) > lipgloss.Width(best) {
			best = x
		}
	}
	return best
}
