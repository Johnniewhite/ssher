package cmd

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/store"
	"github.com/johnniewhite/ssher/internal/ui"
)

var pingTimeout int
var pingParallel int

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "TCP-check the SSH port on every saved server",
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		servers := saved.Vault.SortedServers()
		results := tcpPingAll(servers, time.Duration(pingTimeout)*time.Second, pingParallel)
		for _, r := range results {
			label := fmt.Sprintf("%s (%s:%d)", r.Server.Name, r.Server.Host, r.Server.Port)
			if r.Err == nil {
				fmt.Printf("%s  %s  %s\n", ui.Success.Render("up"), label, r.RTT.Round(time.Millisecond))
			} else {
				fmt.Printf("%s  %s  %s\n", ui.ErrorSty.Render("down"), label, r.Err)
			}
		}
		return nil
	},
}

func init() {
	pingCmd.Flags().IntVar(&pingTimeout, "timeout", 5, "TCP connect timeout (seconds)")
	pingCmd.Flags().IntVarP(&pingParallel, "parallel", "p", 16, "max concurrent probes")
	rootCmd.AddCommand(pingCmd)
}

type pingResult struct {
	Server store.Server
	RTT    time.Duration
	Err    error
}

func tcpPingAll(servers []store.Server, timeout time.Duration, parallelism int) []pingResult {
	if parallelism <= 0 {
		parallelism = 8
	}
	sem := make(chan struct{}, parallelism)
	results := make([]pingResult, len(servers))
	var wg sync.WaitGroup
	for i, s := range servers {
		wg.Add(1)
		i, s := i, s
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			start := time.Now()
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(s.Host, strconv.Itoa(s.Port)), timeout)
			results[i] = pingResult{Server: s, RTT: time.Since(start), Err: err}
			if conn != nil {
				_ = conn.Close()
			}
		}()
	}
	wg.Wait()
	return results
}
