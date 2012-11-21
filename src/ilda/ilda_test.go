package ilda

import (
	"os"
	"testing"
)

func TestIlda(test *testing.T) {
	file, err := os.Open("ani.ild")
	if err != nil {
		test.Fatal("Not correct environment: ", err)
	}
	defer file.Close()

	ani, err := ReadAnimation(file)
	if err != nil {
		test.Fatal("Unable read animation: ", err)
	}

	test.Error("len: ", len(ani.Frames))
	test.Log(ani.String())
	for _, f := range ani.Frames {
		test.Log(f.String())
		for i, _ := range f.Points {
			test.Log(f.Points[i].String())
		}
	}
}
