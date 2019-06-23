package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

const workerCount = 500

type Record []string
type Filter func(r Record) bool

func getCsvReader(r io.Reader) *csv.Reader {
	return csv.NewReader(r)
}

func getRecords(r *csv.Reader) <-chan Record {
	out := make(chan Record)
	go func() {
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			out <- record
		}
		close(out)
	}()
	return out
}

func getFilteredRecords(in <-chan Record, filter Filter) <-chan Record {
	out := make(chan Record)
	go func() {
		for r := range in {
			if filter(r) {
				out <- r
			}
		}
		close(out)
	}()
	return out
}

func mergeFilteredRecords(channels []<-chan Record) <-chan Record {
	var wg sync.WaitGroup
	out := make(chan Record)

	output := func(c <-chan Record) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(channels))
	for _, c := range channels {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func writeRecordsStream(in <-chan Record, w *csv.Writer) {
	for record := range in {
		if err := w.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
}

func buildFilter(p runParams) Filter {
	collectCols := func(r Record) string {
		if 0 == len(p.cols) {
			return strings.Join(r, ",")
		}
		var res string
		for _, cell := range p.cols {
			res += r[cell] + ","
		}
		return res
	}

	re, _ := regexp.Compile(p.pattern)
	regExMatch := func(s string) bool {
		return re.MatchString(s)
	}

	return func(r Record) bool {
		res := regExMatch(collectCols(r))
		if p.invert {
			res = !res
		}
		return res
	}
}

func main() {
	params := initParams()
	r := getCsvReader(getRawReader(params))
	w := csv.NewWriter(os.Stdout)
	f := buildFilter(params)

	var filteredRecordChannels []<-chan Record
	recordsChannel := getRecords(r)
	for i := 0; i < workerCount; i++ {
		filteredRecordChannels = append(filteredRecordChannels, getFilteredRecords(recordsChannel, f))
	}

	writeRecordsStream(mergeFilteredRecords(filteredRecordChannels), w)
}
