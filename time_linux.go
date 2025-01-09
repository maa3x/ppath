//go:build linux

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
		accessed = time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
		if stat.Ctim.Sec != 0 {
			created = time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
		}
	}
	return
}
