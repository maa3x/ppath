//go:build windows

package ppath

import (
	"time"

	"golang.org/x/sys/windows"
)

func getTimes(path string) (created, modified, accessed time.Time) {
	handle, err := openHandle(path)
	if err != nil {
		return
	}
	defer windows.CloseHandle(handle)

	var cTime, aTime, wTime windows.Filetime
	err = windows.GetFileTime(handle, &cTime, &aTime, &wTime)
	if err != nil {
		return
	}
	return time.Unix(0, cTime.Nanoseconds()), time.Unix(0, wTime.Nanoseconds()), time.Unix(0, aTime.Nanoseconds())
}

func openHandle(path string) (windows.Handle, error) {
	pointer, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, windows.ERROR_ABANDONED_WAIT_0
	}
	return windows.CreateFile(
		pointer,
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
}
