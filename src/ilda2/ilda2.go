package main

import (
	"encoding/base32"
	"flag"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
)

const (
	CONFIG_ERROR  = 1
	CONVERT_ERROR = 2
	LIST_ERROR    = 3
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
)

func main() {

	flag.Parse()
	if *cpuprofile != "" {
		file, err := os.Create(*cpuprofile)
		check(err, 5)
		pprof.StartCPUProfile(file)
		defer pprof.StopCPUProfile()
	}

	opt, err := readOpts("ilda2.json")
	if err != nil {
		fail(CONFIG_ERROR, "Unable read options:", err)
	}

	clearDir(opt.TargetDir)

	// convert files
	for i, _ := range opt.Files {
		path := wavFileName(opt.TargetDir, &opt.Files[i])
		for stat := range ilda2wavGo(&opt.Files[i], path) {
			if stat.err != nil {
				fmt.Println()
				fail(CONVERT_ERROR, "Unable convert file",
					opt.Files[i].Name, stat.err)
			}
			status(&opt.Files[i], stat.percent)
		}
		fmt.Println()
	}

	if !opt.GenList {
		return
	}

	// playlist generation
	listPath := filepath.Join(opt.TargetDir, "list.lst")
	f, err := os.OpenFile(listPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fail(LIST_ERROR, err)
	}
	defer f.Close()

	for i, _ := range opt.Files {
		fmt.Fprintln(f, genWavName(&opt.Files[i]), opt.Files[i].Time)
	}

	fmt.Fprintln(f, "#")
}

func genWavName(opt *FileConvOpt) string {
	name := opt.Name
	pInd := strings.LastIndex(name, ".")
	if pInd >= 0 {
		name = name[:pInd]
	}

	buf := []byte(opt.Name)
	buf = append(buf, []byte(opt.Order)...)
	n := uint16(opt.Pps + opt.Fps)
	buf = append(buf, []byte{byte(n >> 8), byte(n)}...)
	crc := crc32.ChecksumIEEE(buf)
	crc16 := uint16(crc) ^ uint16(crc>>16)
	hash := base32.StdEncoding.EncodeToString([]byte{byte(crc16 >> 8), byte(crc16)})
	hash = hash[:4]
	return opt.Name[:4] + hash + ".wav"
}

func wavFileName(dir string, opt *FileConvOpt) string {
	return filepath.Join(dir, genWavName(opt))
}

func clearDir(path string) {
	const msg = "Unable remove all from target dir:"
	dir, err := os.Open(path)
	if err != nil {
		fail(CONVERT_ERROR, msg, err)
	}
	children, err := dir.Readdirnames(-1)
	if err != nil {
		fail(CONVERT_ERROR, msg, err)
	}
	for _, child := range children {
		if err = os.RemoveAll(filepath.Join(path, child)); err != nil {
			fail(CONVERT_ERROR, msg, err)
		}
	}
}

func goToZeroPos() {
	fmt.Print("\x1b[0G")
}

func printProgress(p int) {
	fmt.Printf("[%s%s] %3d%%",
		strings.Repeat("-", p/2),
		strings.Repeat(" ", 50-p/2),
		p)
}

func status(opt *FileConvOpt, persent int) {
	const WIDTH = 15

	name := opt.Name
	if len(name) > WIDTH {
		name = name[:WIDTH-3] + "..."
	}
	goToZeroPos()
	fmt.Printf("%15s ", name)
	printProgress(persent)
}

func check(err error, code int) {
	if err != nil {
		fail(code, err)
	}
}

func fail(code int, a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(code)
}
