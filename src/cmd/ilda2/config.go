package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type ConvOpt struct {
	TargetDir string        `json:"target-dir"`
	GenList   bool          `json:"gen-list"`
	Default   FileConvOpt   `json:"def-opts"`
	Files     []FileConvOpt `json:"convert"`
}

func readOpts(path string) (opt *ConvOpt, err error) {
	// open config
	var file *os.File
	if file, err = os.Open(path); err != nil {
		return
	}
	defer file.Close()

	// set defaults
	opt = new(ConvOpt)
	opt.Default = FileConvOpt{"", 15, 20000, "Y|-1X|B|C", nil, "*"}

	// read config
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(opt); err != nil {
		opt = nil
		return
	}

	// check options
	if err = opt.Default.parseChans(); err != nil {
		opt = nil
		return
	}

	// expand defaults on options
	if err = opt.expand(); err != nil {
		opt = nil
		return
	}

	return
}

func (opt *ConvOpt) expand() error {

	if opt.Default.Fps < 1 {
		return fmt.Errorf("wrong value of default option fps: %d",
			opt.Default.Fps)
	}
	if opt.Default.Pps < 1 {
		return fmt.Errorf("wrong value of default option pps: %d",
			opt.Default.Pps)
	}

	for i, _ := range opt.Files {
		// check file
		if _, err := os.Stat(opt.Files[i].Name); err != nil {
			return err
		}

		opt.Default.expandOn(&opt.Files[i])

		if err := opt.Files[i].parseChans(); err != nil {
			return err
		}
	}

	return nil
}

type FileConvOpt struct {
	Name  string      `json:"name"`
	Fps   int         `json:"fps"`
	Pps   int         `json:"pps"`
	Order string      `json:"order"`
	chans []chanDescr `json:"-"`
	Time  string      `json:"time"`
}

func (opt *FileConvOpt) parseChans() (err error) {

	strs := strings.Split(opt.Order, "|")
	switch len(strs) {
	case 2, 4:
		break
	default:
		err = fmt.Errorf("wrong order format: %s", opt.Order)
		return
	}

	opt.chans = make([]chanDescr, len(strs))

	for i, _ := range strs {
		if !opt.chans[i].parse(strs[i]) {
			err = fmt.Errorf("wrong order format: %s", opt.Order)
			return
		}
	}

	return
}

func (def *FileConvOpt) expandOn(opt *FileConvOpt) {

	expandParam(&opt.Fps, def.Fps)
	expandParam(&opt.Pps, def.Pps)

	if len(opt.Order) == 0 {
		opt.Order = def.Order
	}
	if len(opt.Time) == 0 {
		opt.Time = def.Time
	}
}

func expandParam(dest *int, src int) {
	if *dest < 1 {
		*dest = src
	}
}

type chanDescr struct {
	data byte
	mult float32
}

func (descr *chanDescr) parse(s string) bool {
	switch len(s) {
	case 0:
		return false
	case 1:
		descr.mult = 1
	default:
		mul, err := strconv.ParseFloat(s[:len(s)-1], 32)
		if err != nil {
			return false
		}
		descr.mult = float32(mul)
	}

	descr.data = s[len(s)-1]
	switch descr.data {
	case 'X', 'Y', 'B', 'C', 'Z':
		break
	default:
		return false
	}

	return true
}
