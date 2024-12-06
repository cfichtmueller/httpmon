// Copyright 2024 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cli

import (
	"encoding/csv"
	"io"
)

type CsvWriter struct {
	writer *csv.Writer
}

func newCsvWriter(w io.Writer, comma rune) *CsvWriter {
	writer := csv.NewWriter(w)
	writer.Comma = comma

	return &CsvWriter{
		writer: writer,
	}
}

func (w *CsvWriter) Write(record ...string) error {
	return w.writer.Write(record)
}

func (w *CsvWriter) Flush() {
	w.writer.Flush()
}
