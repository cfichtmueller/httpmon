// Copyright 2024 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cli

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

type Out struct {
	out io.Writer
	err io.Writer
}

func (o *Out) Errorf(format string, a ...any) {
	fmt.Fprintf(o.err, format, a...)
}

func (o *Out) FailAndExit(err error) {
	fmt.Fprintln(o.err, err)
	os.Exit(1)
}

func (o *Out) FailAndExitf(format string, a ...any) {
	fmt.Fprintf(o.err, format, a...)
	os.Exit(1)
}

func (o *Out) Println(a ...any) {
	fmt.Fprintln(o.out, a...)
}

func (o *Out) Printf(format string, a ...any) {
	fmt.Fprintf(o.out, format, a...)
}

func (o *Out) NewCsvWriter(comma rune) *CsvWriter {
	return newCsvWriter(o.out, comma)
}

func (o *Out) NewTabwriter() *TabWriter {
	return &TabWriter{
		tw: tabwriter.NewWriter(o.out, 10, 1, 3, ' ', tabwriter.TabIndent),
	}
}
