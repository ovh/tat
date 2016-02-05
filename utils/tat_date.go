package utils

import (
	"math"
	"time"
)

// SplitFloatForTimeUnix returns a.b a and b, b with 9 numbers
// then, time.Unix(a,b) for initialize date from a float64
func SplitFloatForTimeUnix(in float64) (int64, int64) {
	a := float64(int64(in))
	c := Round((in - a) * 1000000000)
	return int64(a), int64(c)
}

// DateFromFloat returns a time.Time from a float
func DateFromFloat(in float64) time.Time {
	sec, nsec := SplitFloatForTimeUnix(in)
	return time.Unix(sec, nsec)
}

// TatTSFromDate returns a timestamp with float.
func TatTSFromDate(in time.Time) float64 {
	return float64(in.UnixNano()) / 1000000000
}

// TatTSFromNow returns a timestamp with float for timeNow()
func TatTSFromNow() float64 {
	return TatTSFromDate(time.Now())
}

// Round rounds float
func Round(f float64) float64 {
	return math.Floor(f + .5)
}
