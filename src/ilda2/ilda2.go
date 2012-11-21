package main

import (
	"fmt"
	"os"
)

const (
	CONFIG_ERROR  = 1
	CONVERT_ERROR = 2
)

// TODO: GIT
func main() {
	opt, _ := readOpts("ilda2.json")

	// clear target dir
	if err := os.RemoveAll(opt.TargetDir); err != nil {
		fail(CONVERT_ERROR, "Unable remove all from target dir: ", err)
	}

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
