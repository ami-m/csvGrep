package main

import (
	"flag"
	"strconv"
	"strings"
)

type runParams struct {
	pattern   string
	file      string
	separator rune
	invert    bool
	cols      []int
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ", ")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func initParams() runParams {
	var res runParams
	var pattern string
	var file string
	var separator string
	var invert bool
	var cols arrayFlags

	flag.StringVar(&pattern, "e", "", "regex pattern to match")
	flag.StringVar(&file, "f", "", "path to input file instead of stdin")
	flag.StringVar(&separator, "s", ",", "separator character (defaults to a comma)")
	flag.BoolVar(&invert, "v", false, "invert (like -v in grep) return only the rows that *don't* fulfill the pattern")
	flag.Var(&cols, "c", "list of columns to operate on")

	flag.Parse()

	res.pattern = pattern
	res.file = file
	res.invert = invert

	var actualCols []int
	for _, v := range cols {
		if index, err := strconv.Atoi(v); err == nil {
			actualCols = append(actualCols, index)
		}
	}
	res.cols = actualCols

	separatorRunes := []rune(separator)
	res.separator = separatorRunes[0]

	return res
}
