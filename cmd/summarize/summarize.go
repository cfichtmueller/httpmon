// Copyright 2024 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package summarize

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cfichtmueller/httpmon/cli"
	"github.com/cfichtmueller/httpmon/engine"
	"github.com/spf13/cobra"
)

type summarizeopts struct {
	file                 string
	ignoreInvalidRecords bool
}

func NewCommand(mcli *cli.Cli) *cobra.Command {

	opts := summarizeopts{}

	cmd := &cobra.Command{
		Use:   "summarize",
		Short: "Summarize monitoring results",
		Run: func(cmd *cobra.Command, args []string) {
			var r io.Reader
			if opts.file != "" {
				f, err := os.Open(opts.file)
				if err != nil {
					mcli.Out.FailAndExit(err)
				}
				r = f
				defer f.Close()
			} else {
				r = os.Stdin
			}
			if err := runSummarize(mcli, opts, r); err != nil {
				mcli.Out.FailAndExit(err)
			}
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.file, "file", "f", "", "Read from file")
	flags.BoolVarP(&opts.ignoreInvalidRecords, "ignore", "i", false, "Ignore invalid records")

	return cmd
}

func runSummarize(mcli *cli.Cli, opts summarizeopts, r io.Reader) error {
	var reader Reader
	if mcli.Csv {
		cr := csv.NewReader(r)
		cr.Comma = ';'
		reader = cr
	} else {
		return fmt.Errorf("unsupported format")
	}
	line := 0
	pings := make([]*engine.Ping, 0)
	for {
		line += 1
		record, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			if opts.ignoreInvalidRecords {
				continue
			}
			return err
		}
		if record == nil {
			break
		}

		if len(record) != 13 {
			if opts.ignoreInvalidRecords {
				continue
			}
			return fmt.Errorf("invalid record on line %d", line)
		}
		p, err := parsePing(mcli, record)
		if err != nil {
			if opts.ignoreInvalidRecords {
				continue
			}
			return err
		}
		pings = append(pings, p)
	}
	allStats := engine.Summarize(pings)
	w := mcli.Out.NewTabwriter()
	w.Write(
		"URL",
		"AVAILABILITY",
		"AVG RT",
		"MEDIAN RT",
		"LONGEST RT",
		"WORST MONITOR",
		"MEASUREMENTS",
		"FAILED MEASUREMENTS",
	)
	for _, stats := range allStats {
		w.Write(
			stats.Endpoint,
			mcli.Formatter.FormatPercentage(stats.Availability),
			mcli.Formatter.FormatDurationms(stats.AvgResponseTime),
			mcli.Formatter.FormatDurationms(stats.MedianResponseTime),
			mcli.Formatter.FormatDurationms(stats.LongestResponseTime),
			stats.WorstMonitor,
			mcli.Formatter.FormatInt(stats.NumberOfMeasurements),
			mcli.Formatter.FormatInt(stats.NumberOfFailedMeasurements),
		)
	}
	w.Flush()
	return nil
}

type Reader interface {
	Read() ([]string, error)
}

func parsePing(mcli *cli.Cli, record []string) (*engine.Ping, error) {
	timestamp, err := mcli.In.ParseTime(record[3])
	if err != nil {
		return nil, err
	}
	statusCode, err := mcli.In.ParseInt(record[4])
	if err != nil {
		return nil, err
	}
	dnsTime, err := mcli.In.ParseDurationms(record[6])
	if err != nil {
		return nil, err
	}
	connectionTime, err := mcli.In.ParseDurationms(record[7])
	if err != nil {
		return nil, err
	}
	tlsTime, err := mcli.In.ParseDurationms(record[8])
	if err != nil {
		return nil, err
	}
	ttfb, err := mcli.In.ParseDurationms(record[9])
	if err != nil {
		return nil, err
	}
	downloadTime, err := mcli.In.ParseDurationms(record[10])
	if err != nil {
		return nil, err
	}
	totalResponseTime, err := mcli.In.ParseDurationms(record[11])
	if err != nil {
		return nil, err
	}
	certRemainingValidity, err := mcli.In.ParseDurations(record[12])
	if err != nil {
		return nil, err
	}
	return &engine.Ping{
		Name:                  record[0],
		URL:                   record[1],
		Status:                record[2],
		Timestamp:             timestamp,
		StatusCode:            statusCode,
		Message:               record[5],
		DNSTime:               dnsTime,
		ConnectionTime:        connectionTime,
		TLSTime:               tlsTime,
		TTFB:                  ttfb,
		DownloadTime:          downloadTime,
		TotalResponseTime:     totalResponseTime,
		CertRemainingValidity: certRemainingValidity,
	}, nil

}
