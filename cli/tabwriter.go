// Copyright 2024 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cli

import (
	"strings"
	"text/tabwriter"
)

type TabWriter struct {
	tw *tabwriter.Writer
}

func (w *TabWriter) Write(record ...string) error {
	if _, err := w.tw.Write([]byte(strings.Join(record, "\t") + "\n")); err != nil {
		return err
	}
	return nil
}

func (w *TabWriter) Flush() {
	if err := w.tw.Flush(); err != nil {
		panic(err)
	}
}
