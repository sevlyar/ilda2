package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	CONFIG_ERROR  = 1
	CONVERT_ERROR = 2
)

func main() {
	opt, err := readOpts("ilda2.json")
	if err != nil {
		fail(CONFIG_ERROR, "Unable read options:", err)
	}

	// clear target dir
	if err = os.RemoveAll(opt.TargetDir); err != nil {
		fail(CONVERT_ERROR, "Unable remove all from target dir:", err)
	}

	// convert files
	for i, _ := range opt.Files {
		for stat := range ilda2wavGo(&opt.Files[i], opt.TargetDir) {
			if stat.err != nil {
				fmt.Println()
				fail(CONVERT_ERROR, "Unable convert file",
					opt.Files[i].Name, stat.err)
			}
			status(&opt.Files[i], stat.percent)
		}
		fmt.Println()
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
