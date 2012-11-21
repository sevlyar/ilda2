package main

import (
	"bytes"
	"ilda"
	"io"
	"math"
	"os"
	"path/filepath"
)

type opStatus struct {
	err     error // operation error
	percent int   // complete status
}

func sendError(status chan<- *opStatus, err error) {
	status <- &opStatus{err, -1}
	close(status)
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

	file, err := os.OpenFile(opt.Name, os.O_RDONLY, 0644)
	if err != nil {
		sendError(status, err)
		return
	}
	defer file.Close()

	ani, err := ilda.ReadAnimation(file)
	if err != nil {
		sendError(status, err)
		return
	}

	var stream bytes.Buffer

	l := len(ani.Frames)
	for i, frame := range ani.Frames {
		repeat := opt.Pps / (opt.Fps * len(frame.Points))

		convertFrame(&stream, frame, opt.chans, repeat)
		status <- newStatusPercent(100 * (i + 1) / l)
	}

	wav, err := os.OpenFile(wavFileName(targetDir, opt.Name),
		os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		sendError(status, err)
		return
	}
	defer wav.Close()

	err = WriteWav(wav, &stream, stream.Len(), len(opt.chans), opt.Pps)
	if err != nil {
		sendError(status, err)
		return
	}

	close(status)
}

func wavFileName(dir, file string) string {
	return filepath.Join(dir, file) + ".wav"
}

func convertFrame(w io.Writer, f *ilda.Table, chans []chanDescr, repeat int) {
	p := make([]byte, 2)

	// TODO: cache sequence and copy
	for repeat > 0 {
		repeat--

		for i := 0; i < len(f.Points); i++ {
			for j := 0; j < len(chans); j++ {
				var v int16

				switch chans[j].data {
				case 'X':
					v = int16(chans[j].mult * float32(f.Points[i].X))
				case 'Y':
					v = int16(chans[j].mult * float32(f.Points[i].Y))
				case 'Z':
					v = int16(chans[j].mult * float32(f.Points[i].Z))
				case 'B':
					if !f.Points[i].Status.IsBlank() {
						v = math.MaxInt16
					}
				case 'C':
					v = int16(f.Points[i].Status.GetColor())
				}

				p[0] = byte(v >> 8)
				p[1] = byte(v)
				w.Write(p)
			}
		}
	}
}
