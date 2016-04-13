package main

import (
	"testing"
	"time"
)

func TestA(t *testing.T) {
	now := time.Now()
	sec := now.Unix()
	nano := now.UnixNano()
	nanos := now.Nanosecond()
	mills := (now.Unix() * 1000 + int64(now.Nanosecond() / 1e6))

	print(sec, " ", nano, " ", nanos, " ", mills, "\n")
	print(mills / 1e3, "\n")
}