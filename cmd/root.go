// Copyright 2024 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cmd

import (
	"os"

	"github.com/cfichtmueller/httpmon/cli"
	"github.com/cfichtmueller/httpmon/cmd/monitor"
	"github.com/spf13/cobra"
)

type rootopts struct {
	batch bool
	csv   bool
}

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	mcli := cli.New(
		cli.DefaultFormatter(),
		os.Stdout,
		os.Stderr,
	)

	opts := rootopts{}

	cmd := &cobra.Command{
		Use:   "httpmon",
		Short: "A one-shot tool for monitoring HTTP and HTTPS endpoints.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			mcli.Batch = opts.batch
			mcli.Csv = opts.csv
		},
	}

	persistentFlags := cmd.PersistentFlags()
	persistentFlags.BoolVarP(&opts.batch, "batch", "b", false, "batch mode")
	persistentFlags.BoolVar(&opts.csv, "csv", false, "produce csv output")

	cmd.AddCommand(
		monitor.NewCommand(mcli),
	)

	return cmd
}
