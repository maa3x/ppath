//go:build darwin

package ppath

import (
	"os"
	"syscall"
	"time"
)

func getTimes(path string) (created, modified, accessed time.Time) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	modified = info.ModTime()
	created = modified

	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		accessed = time.Unix(int64(stat.Atimespec.Sec), int64(stat.Atimespec.Nsec))
		if stat.Birthtimespec.Sec != 0 {
			created = time.Unix(int64(stat.Birthtimespec.Sec), int64(stat.Birthtimespec.Nsec))
		} else if stat.Ctimespec.Sec != 0 {
			created = time.Unix(int64(stat.Ctimespec.Sec), int64(stat.Ctimespec.Nsec))
		}
	}
	return
}
