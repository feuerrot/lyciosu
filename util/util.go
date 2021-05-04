package util

import (
	"fmt"
	"time"
)

func FSize(i float64) string {
	fix := []string{" ", "k", "M", "G"}
	cnt := 0
	for {
		if i < 1024 {
			return fmt.Sprintf("%.2f%sB", i, fix[cnt])
		}

		i = i / 1024
		cnt += 1
	}
}

func FSizeDurationSpeed(size int64, duration time.Duration) string {
	s := float64(size)
	si := FSize(s)
	du := duration.Seconds()
	sp := FSize(s / du)

	return fmt.Sprintf("%9s %7.2fs %9s/s", si, du, sp)
}
