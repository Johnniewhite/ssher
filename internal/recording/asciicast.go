// Package recording reads and writes asciicast v2 files. The format is the
// de-facto standard for terminal session recordings and is what the Python
// version's recordings can be migrated to (it shipped a custom format; we
// don't support reading those — `ssher record replay` only handles v2).
//
// Asciicast v2 spec: https://docs.asciinema.org/manual/asciicast/v2/
//
//	{"version":2,"width":80,"height":24,"timestamp":1700000000,"title":"prod"}
//	[0.123, "o", "ls\r\n"]
//	[0.456, "o", "file1\nfile2\n"]
package recording

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/johnniewhite/ssher/internal/paths"
)

// Header is line 1 of an asciicast v2 file.
type Header struct {
	Version   int    `json:"version"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Timestamp int64  `json:"timestamp"`
	Title     string `json:"title,omitempty"`
}

// Writer streams output bytes to disk as asciicast v2 events. Callers wrap
// this with an io.MultiWriter to tee the SSH session output into both the
// user's terminal and the recording.
//
// Not safe for concurrent Write calls.
type Writer struct {
	f     *os.File
	enc   *json.Encoder
	start time.Time
	mu    sync.Mutex
}

// NewWriter creates ~/.ssher/recordings/<title>-<timestamp>.cast and writes
// the header. Returns the writer and the final path.
func NewWriter(title string, width, height int) (*Writer, string, error) {
	dir, err := paths.RecordingsDir()
	if err != nil {
		return nil, "", err
	}
	if err := os.MkdirAll(dir, paths.DirMode); err != nil {
		return nil, "", err
	}
	stamp := time.Now().UTC().Format("20060102-150405.000000000")
	path := filepath.Join(dir, fmt.Sprintf("%s-%s.cast", sanitize(title), stamp))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, paths.FileMode)
	if err != nil {
		return nil, "", err
	}
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(Header{
		Version:   2,
		Width:     width,
		Height:    height,
		Timestamp: time.Now().Unix(),
		Title:     title,
	}); err != nil {
		_ = f.Close()
		return nil, path, err
	}
	return &Writer{f: f, enc: enc, start: time.Now()}, path, nil
}

func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delta := time.Since(w.start).Seconds()
	if err := w.enc.Encode([]any{delta, "o", string(p)}); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *Writer) Close() error { return w.f.Close() }

// Replay reads an asciicast v2 file and prints its output stream to dst,
// honouring the original timing. Stops on EOF or write error.
func Replay(path string, dst io.Writer, speed float64) error {
	if speed <= 0 {
		speed = 1.0
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	br := bufio.NewReader(f)
	headerLine, err := br.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}
	var h Header
	if err := json.Unmarshal(headerLine, &h); err != nil {
		return fmt.Errorf("parse header: %w", err)
	}
	if h.Version != 2 {
		return fmt.Errorf("unsupported asciicast version %d", h.Version)
	}

	var lastT float64
	for {
		line, err := br.ReadBytes('\n')
		if len(line) > 0 {
			var ev []json.RawMessage
			if err := json.Unmarshal(line, &ev); err != nil {
				return fmt.Errorf("parse event: %w", err)
			}
			if len(ev) != 3 {
				continue
			}
			var t float64
			var kind string
			var data string
			_ = json.Unmarshal(ev[0], &t)
			_ = json.Unmarshal(ev[1], &kind)
			_ = json.Unmarshal(ev[2], &data)
			if kind != "o" {
				continue
			}
			delay := time.Duration(((t - lastT) / speed) * float64(time.Second))
			if delay > 0 {
				time.Sleep(delay)
			}
			lastT = t
			if _, werr := io.WriteString(dst, data); werr != nil {
				return werr
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// List returns recording paths sorted oldest-first.
func List() ([]string, error) {
	dir, err := paths.RecordingsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".cast" {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(out)
	return out, nil
}

func sanitize(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			out = append(out, r)
		case r == '-' || r == '_' || r == '.':
			out = append(out, r)
		default:
			out = append(out, '_')
		}
	}
	if len(out) == 0 {
		return "session"
	}
	return string(out)
}
