package main

import (
	"time"
)

type opStatus struct {
	err     error // operation error
	percent int   // complete status
}

func newStatusError(err error) *opStatus {
	return &opStatus{err, -1}
}

func newStatusPercent(percent int) *opStatus {
	return &opStatus{nil, percent}
}

func ilda2wavGo(opt *FileConvOpt, targetDir string) <-chan *opStatus {
	status := make(chan *opStatus, 8)
	go ilda2wav(opt, targetDir, status)
	return status
}

func ilda2wav(opt *FileConvOpt, targetDir string, status chan<- *opStatus) {
	for i := 1; i <= 100; i++ {
		status <- newStatusPercent(i)
		time.Sleep(25 * time.Millisecond)
	}
	close(status)
}
