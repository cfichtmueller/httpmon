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

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	mcli := cli.New(
		cli.DefaultFormatter(),
		os.Stdout,
		os.Stderr,
	)

	cmd := &cobra.Command{
		Use:   "httpmon",
		Short: "A one-shot tool for monitoring HTTP and HTTPS endpoints.",
	}

	cmd.AddCommand(
		monitor.NewCommand(mcli),
	)

	return cmd
}
