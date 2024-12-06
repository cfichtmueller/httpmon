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

	var writer Writer

	if mcli.Csv {
		writer = mcli.Out.NewCsvWriter(';')
	} else {
		writer = mcli.Out.NewTabwriter()
	}

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

	if !mcli.Batch {
		writer.Write(
			"MONITOR",
			"URL",
			"STATUS",
			"TIMESTAMP",
			"CODE",
			"MESSAGE",
			"DNS",
			"CONNECTION",
			"TLS",
			"TTFB",
			"DOWNLOAD",
			"RESPONSE",
			"CERT VALIDITY",
		)
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
		Retries:             2,
		RetryInterval:       10,
		ConnectTimeout:      5 * time.Second,
		ResponseTimeout:     5 * time.Second,
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
		formatter.FormatTime(ping.Timestamp),
		strconv.Itoa(ping.StatusCode),
		ping.Message,
		formatter.FormatDurationms(ping.DNSTime),
		formatter.FormatDurationms(ping.ConnectionTime),
		formatter.FormatDurationms(ping.TLSTime),
		formatter.FormatDurationms(ping.TTFB),
		formatter.FormatDurationms(ping.DownloadTime),
		formatter.FormatDurationms(ping.TotalResponseTime),
		formatter.FormatDurations(ping.CertRemainingValidity),
	)
	wg.Done()
}

type Writer interface {
	Write(record ...string) error
	Flush()
}
