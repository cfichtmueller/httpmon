// Copyright 2024 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cli

import (
	"strconv"
	"time"
)

type Formatter interface {
	// FormatTimems formats a time as milliseconds
	FormatTimems(d time.Duration) string
	// FormatTimes formats a time as seconds
	FormatTimes(d time.Duration) string
}

type defaultFormatter struct{}

func DefaultFormatter() Formatter {
	return &defaultFormatter{}
}

func (f *defaultFormatter) FormatTimems(d time.Duration) string {
	return strconv.FormatInt(d.Milliseconds(), 10)
}

func (f *defaultFormatter) FormatTimes(d time.Duration) string {
	return strconv.FormatInt(d.Milliseconds()/1000, 10)
}
