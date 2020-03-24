# csvGrep

grep like util for csv

## Getting Started

Build the executable with `go build`.

`./csvGrep --help` will give you this:
```shell script
Usage of ./csvGrep:
  -c value
        list of columns to operate on
  -e string
        regex pattern to match
  -f string
        path to input file instead of stdin
  -s string
        separator character (defaults to a comma) (default ",")
  -v    invert (like -v in grep) return only the rows that *don't* fulfill the pattern
```

## Example
Find bad strings (????? instead of names) in the customers file:
```shell script
./csvGrep -e="[?]{4,}" -f="./customers.csv"
```


## Built With

* [Golang](https://golang.org/) - The go language