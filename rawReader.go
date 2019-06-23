package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func getRawReader(p runParams) io.Reader {
	if p.file != "" {
		return getFileReader(p.file)
	}

	return bufio.NewReader(os.Stdin)
}

func getFileReader(path string) io.Reader {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return file
}
