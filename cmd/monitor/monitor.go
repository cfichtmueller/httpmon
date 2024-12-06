// Copyright 2024 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package monitor

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cfichtmueller/httpmon/cli"
	"github.com/cfichtmueller/httpmon/engine"
	"github.com/spf13/cobra"
)

type monitoropts struct {
	file string
	name string
	urls []string
}

func NewCommand(mcli *cli.Cli) *cobra.Command {
	opts := monitoropts{}

	cmd := &cobra.Command{
		Use:   "monitor [URL]...",
		Short: "Monitor HTTP endpoints",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				opts.urls = args
			}
			if err := runMonitor(mcli, opts); err != nil {
				mcli.Out.FailAndExit(err)
			}
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.file, "file", "f", "", "file to read URLs from")
	flags.StringVarP(&opts.name, "name", "n", "", "name of the monitor")

	return cmd
}

func runMonitor(mcli *cli.Cli, opts monitoropts) error {
	name := opts.name
	if name == "" {
		n, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("unable to determine hostname: %v", err)
		}
		name = n
	}

	writer := mcli.Out.NewCsvWriter(';')
	wait := &sync.WaitGroup{}
	urls := opts.urls

	if opts.file != "" && len(opts.urls) > 0 {
		return fmt.Errorf("cannot use URLs from file and arguments simultaneously")
	}
	if opts.file != "" {
		b, err := os.ReadFile(os.Args[2])
		if err != nil {
			return fmt.Errorf("unable to read file %s: %v", os.Args[2], err)
		}
		urls = strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n")
	}

	invalid := false
	for _, ru := range urls {
		if ru == "" {
			continue
		}
		u, err := url.Parse(ru)
		if err != nil {
			mcli.Out.Errorf("Invalid url '%s': %v\n", ru, err)
			invalid = true
		} else if u.Scheme != "http" && u.Scheme != "https" {
			mcli.Out.Errorf("Invalid url '%s'\n", ru)
			invalid = true
		}
	}
	if invalid {
		os.Exit(1)
	}

	for _, u := range urls {
		if u == "" {
			continue
		}
		wait.Add(1)
		go pingUrl(writer, mcli.Formatter, wait, name, u)
	}

	wait.Wait()
	writer.Flush()
	return nil
}

func pingUrl(w Writer, formatter cli.Formatter, wg *sync.WaitGroup, name, url string) {
	monitor := &engine.Monitor{
		Name:                name,
		URL:                 url,
		Retries:             3,
		RetryInterval:       10,
		ConnectTimeout:      5 * time.Second,
		ResponseTimeout:     10 * time.Second,
		MaxRedirects:        3,
		AcceptedStatusCodes: []int{200, 201, 202, 204},
		HTTPMethod:          "GET",
		Headers:             map[string]string{"User-Agent": "HTTP-Monitor-Agent"},
	}
	ping := engine.ExecutePing(monitor)

	w.Write(
		ping.Name,
		ping.URL,
		ping.Status,
		ping.Timestamp.Format(time.RFC3339),
		strconv.Itoa(ping.StatusCode),
		ping.Message,
		formatter.FormatTimems(ping.DNSTime),
		formatter.FormatTimems(ping.ConnectionTime),
		formatter.FormatTimems(ping.TLSTime),
		formatter.FormatTimems(ping.TTFB),
		formatter.FormatTimems(ping.DownloadTime),
		formatter.FormatTimems(ping.TotalResponseTime),
		formatter.FormatTimes(ping.CertRemainingValidity),
	)
	wg.Done()
}

type Writer interface {
	Write(record ...string) error
	Flush()
}
