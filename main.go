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

func getRecords(done <-chan bool, r *csv.Reader) <-chan Record {
	out := make(chan Record)
	go func() {
		defer close(out)

		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			select {
			case <-done:
				return
			case out <- record:

			}

		}
	}()
	return out
}

func getFilteredRecords(done <-chan bool, in <-chan Record, filter Filter) <-chan Record {
	out := make(chan Record)
	go func() {
		defer close(out)

		for r := range in {
			select {
			case <-done:
				return
			default:
				if filter(r) {
					out <- r
				}
			}
		}
	}()
	return out
}

func mergeFilteredRecords(done <-chan bool, channels []<-chan Record) <-chan Record {
	var wg sync.WaitGroup
	out := make(chan Record)

	output := func(c <-chan Record) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:

			case <-done:
				return
			}
		}
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

func writeRecordsStream(done <-chan bool, in <-chan Record, w *csv.Writer) {
	for record := range in {
		select {
		case <-done:
			log.Println("stopped printing")
		default:
		}
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

	done := make(chan bool)
	defer close(done)

	var filteredRecordChannels []<-chan Record
	recordsChannel := getRecords(done, r)
	for i := 0; i < workerCount; i++ {
		filteredRecordChannels = append(filteredRecordChannels, getFilteredRecords(done, recordsChannel, f))
	}

	writeRecordsStream(done, mergeFilteredRecords(done, filteredRecordChannels), w)
}
