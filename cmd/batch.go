package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/store"
	"github.com/johnniewhite/ssher/internal/ui"
)

// CSV / JSON import / export. Plaintext exports DO NOT include passwords —
// the user can re-add them via `edit`. This matches the Python tool's
// security stance and is intentional.

var importCSVCmd = &cobra.Command{
	Use:   "import-csv <path>",
	Short: "Import servers from a CSV file",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		f, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer f.Close()
		r := csv.NewReader(f)
		rows, err := r.ReadAll()
		if err != nil {
			return fmt.Errorf("parse csv: %w", err)
		}
		if len(rows) < 2 {
			return fmt.Errorf("csv has no data rows")
		}
		header := rows[0]
		idx := func(name string) int {
			for i, h := range header {
				if strings.EqualFold(strings.TrimSpace(h), name) {
					return i
				}
			}
			return -1
		}
		nameI, hostI, userI := idx("name"), idx("host"), idx("user")
		portI, authI, groupI := idx("port"), idx("auth_type"), idx("group")
		tagsI, notesI := idx("tags"), idx("notes")
		if nameI < 0 || hostI < 0 || userI < 0 {
			return fmt.Errorf("csv must have at least name,host,user columns")
		}
		added := 0
		for rowNum, row := range rows[1:] {
			if nameI >= len(row) || hostI >= len(row) || userI >= len(row) {
				fmt.Println(ui.Warnf(fmt.Sprintf("row %d is missing a required column", rowNum+2)))
				continue
			}
			s := store.Server{
				Name: row[nameI],
				Host: row[hostI],
				User: row[userI],
			}
			if value, ok := csvCell(row, portI); ok {
				if p, err := strconv.Atoi(value); err == nil {
					s.Port = p
				}
			}
			if value, ok := csvCell(row, authI); ok {
				s.AuthType = store.AuthType(value)
			}
			if value, ok := csvCell(row, groupI); ok {
				s.Group = value
			}
			if value, ok := csvCell(row, tagsI); ok {
				s.Tags = splitCSV(value)
			}
			if value, ok := csvCell(row, notesI); ok {
				s.Notes = value
			}
			if err := saved.Vault.AddServer(s); err != nil {
				fmt.Println(ui.Warnf(err.Error()))
				continue
			}
			added++
		}
		if err := saved.SaveAndRefreshSession(); err != nil {
			return err
		}
		fmt.Println(ui.Successf(fmt.Sprintf("imported %d server(s)", added)))
		return nil
	},
}

var importJSONCmd = &cobra.Command{
	Use:   "import-json <path>",
	Short: "Import servers from a JSON file",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		b, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		var batch struct {
			Servers []store.Server `json:"servers"`
		}
		// Accept either {"servers": [...]} or a bare array.
		if err := json.Unmarshal(b, &batch); err != nil || batch.Servers == nil {
			var bare []store.Server
			if err2 := json.Unmarshal(b, &bare); err2 != nil {
				return fmt.Errorf("parse json: %w", err)
			}
			batch.Servers = bare
		}
		added := 0
		for _, s := range batch.Servers {
			if err := saved.Vault.AddServer(s); err != nil {
				fmt.Println(ui.Warnf(err.Error()))
				continue
			}
			added++
		}
		if err := saved.SaveAndRefreshSession(); err != nil {
			return err
		}
		fmt.Println(ui.Successf(fmt.Sprintf("imported %d server(s)", added)))
		return nil
	},
}

var exportCSVOut string

var exportCSVCmd = &cobra.Command{
	Use:   "export-csv",
	Short: "Export servers to CSV (passwords excluded)",
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		w := os.Stdout
		if exportCSVOut != "" {
			f, err := openPrivateOutput(exportCSVOut)
			if err != nil {
				return err
			}
			defer f.Close()
			w = f
		}
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"name", "host", "user", "port", "auth_type", "group", "tags", "notes"}); err != nil {
			return err
		}
		for _, s := range saved.Vault.Servers {
			if err := cw.Write([]string{
				s.Name, s.Host, s.User,
				strconv.Itoa(s.Port),
				string(s.AuthType),
				s.Group,
				strings.Join(s.Tags, ","),
				s.Notes,
			}); err != nil {
				return err
			}
		}
		cw.Flush()
		if err := cw.Error(); err != nil {
			return err
		}
		if exportCSVOut != "" {
			fmt.Fprintln(os.Stderr, ui.Successf("exported "+exportCSVOut))
		}
		return nil
	},
}

var exportJSONOut string
var exportJSONIncludePasswords bool

var exportJSONCmd = &cobra.Command{
	Use:   "export-json",
	Short: "Export servers to JSON (passwords excluded by default)",
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		export := make([]store.Server, len(saved.Vault.Servers))
		copy(export, saved.Vault.Servers)
		if !exportJSONIncludePasswords {
			for i := range export {
				export[i].Password = ""
			}
		}
		b, err := json.MarshalIndent(map[string]any{"servers": export}, "", "  ")
		if err != nil {
			return err
		}
		if exportJSONOut == "" {
			_, err := os.Stdout.Write(append(b, '\n'))
			return err
		}
		f, err := openPrivateOutput(exportJSONOut)
		if err != nil {
			return err
		}
		if _, err := f.Write(b); err != nil {
			_ = f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		fmt.Println(ui.Successf("exported " + exportJSONOut))
		return nil
	},
}

func csvCell(row []string, index int) (string, bool) {
	if index < 0 || index >= len(row) {
		return "", false
	}
	return row[index], true
}

func openPrivateOutput(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, err
	}
	if err := f.Chmod(0o600); err != nil {
		_ = f.Close()
		return nil, err
	}
	return f, nil
}

func init() {
	exportCSVCmd.Flags().StringVarP(&exportCSVOut, "output", "o", "", "write to FILE instead of stdout")
	exportJSONCmd.Flags().StringVarP(&exportJSONOut, "output", "o", "", "write to FILE instead of stdout")
	exportJSONCmd.Flags().BoolVar(&exportJSONIncludePasswords, "include-passwords", false,
		"include plaintext passwords (DANGEROUS — use only for trusted backups)")
	rootCmd.AddCommand(importCSVCmd, importJSONCmd, exportCSVCmd, exportJSONCmd)
}
