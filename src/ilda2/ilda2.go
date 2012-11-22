package main

import (
	"fmt"
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

func main() {
	profile()
	defer pprof.StopCPUProfile()

	opt, err := readOpts("ilda2.json")
	if err != nil {
		fail(CONFIG_ERROR, "Unable read options:", err)
	}

	clearDir(opt.TargetDir)

	// convert files
	for i, _ := range opt.Files {
		path := wavFileName(opt.TargetDir, opt.Files[i].Name)
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

	listPath := filepath.Join(opt.TargetDir, "list.lst")
	f, err := os.OpenFile(listPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fail(LIST_ERROR, err)
	}
	defer f.Close()

	for i, _ := range opt.Files {
		fmt.Fprintln(f, opt.Files[i].Name+".wav", opt.Files[i].Time)
	}

	fmt.Fprintln(f, "#")
}

func wavFileName(dir, file string) string {
	return filepath.Join(dir, file) + ".wav"
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

func profile() {
	file, err := os.Create(".prof")
	check(err, 5)
	pprof.StartCPUProfile(file)
}
