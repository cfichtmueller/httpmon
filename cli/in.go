// Copyright 2024 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cli

import (
	"strconv"
	"time"
)

type In struct{}

func (i *In) ParseInt(in string) (int, error) {
	return strconv.Atoi(in)
}

func (i *In) ParseDurationms(in string) (time.Duration, error) {
	return i.parseDuration(in, time.Millisecond)
}

func (i *In) ParseDurations(in string) (time.Duration, error) {
	return i.parseDuration(in, time.Second)
}

func (i *In) parseDuration(in string, multiplier time.Duration) (time.Duration, error) {
	v, err := strconv.Atoi(in)
	if err != nil {
		return 0, err
	}
	return time.Duration(v) * multiplier, nil
}

func (i *In) ParseTime(in string) (time.Time, error) {
	return time.Parse(time.RFC3339, in)
}
