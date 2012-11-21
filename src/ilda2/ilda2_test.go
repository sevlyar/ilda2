package main

import (
	"bytes"
	"os"
	"testing"
)

func TestWav(test *testing.T) {
	file, err := os.OpenFile("snd.wav",
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		test.Fatal(err)
	}
	defer file.Close()

	samples := &bytes.Buffer{}
	samples.Write([]byte{0, 0, 0, 0})

	err = WriteWav(file, samples, samples.Len(), 2, 44100)
	if err != nil {
		test.Fatal(err)
	}
}

func TestOpt(test *testing.T) {
	_, err := readOpts("ilda2.json")
	if err != nil {
		test.Fatal(err)
	}
}
