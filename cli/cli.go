// Copyright 2024 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cli

import (
	"io"
)

type Cli struct {
	Csv       bool
	Batch     bool
	Formatter Formatter
	In        *In
	Out       *Out
}

func New(
	formatter Formatter,
	out, err io.Writer,
) *Cli {
	return &Cli{
		Formatter: formatter,
		In:        &In{},
		Out: &Out{
			out: out,
			err: err,
		},
	}
}
