// Copyright 2024 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cli

import (
	"fmt"
	"strconv"
	"time"
)

type Formatter interface {
	FormatInt(i int) string
	FormatPercentage(p float64) string
	FormatTime(t time.Time) string
	// FormatDurationms formats a duration as milliseconds
	FormatDurationms(d time.Duration) string
	// FormatDurations formats a duration as seconds
	FormatDurations(d time.Duration) string
}

type defaultFormatter struct{}

func DefaultFormatter() Formatter {
	return &defaultFormatter{}
}

func (f *defaultFormatter) FormatInt(i int) string {
	return strconv.Itoa(i)
}

func (f *defaultFormatter) FormatPercentage(p float64) string {
	return fmt.Sprintf("%.2f%%", p)
}

func (f *defaultFormatter) FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func (f *defaultFormatter) FormatDurationms(d time.Duration) string {
	return strconv.FormatInt(d.Milliseconds(), 10)
}

func (f *defaultFormatter) FormatDurations(d time.Duration) string {
	return strconv.FormatInt(d.Milliseconds()/1000, 10)
}
