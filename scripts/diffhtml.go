package main

import (
	"github.com/documize/html-diff"
	"io/ioutil"
	"fmt"
	"flag"
)

var cfg = &htmldiff.Config{
	Granularity:  5,
	InsertedSpan: []htmldiff.Attribute{{Key: "style", Val: "background-color: palegreen;"}},
	DeletedSpan:  []htmldiff.Attribute{{Key: "style", Val: "background-color: lightpink;"}},
	ReplacedSpan: []htmldiff.Attribute{{Key: "style", Val: "background-color: lightskyblue;"}},
	CleanTags:    []string{""},
}

func main() {
	prevPath := flag.String("prev", ".", "previous file to compare")
	latestPath := flag.String("latest", ".", "latest file to compare")
	flag.Parse()

	bufferPrev, err := ioutil.ReadFile(*prevPath)
	if err != nil {
		fmt.Println(fmt.Errorf("error: %v", err))
		return
	}

	bufferLatest, err := ioutil.ReadFile(*latestPath)
	if err != nil {
		fmt.Println(fmt.Errorf("error: %v", err))
		return
	}

	sizePrev := len(bufferPrev)
	sizeLatest := len(bufferLatest)

	if sizeLatest == sizePrev {
		fmt.Println("Same size")
		return
	}

	previousHTML := string(bufferPrev)
	latestHTML := string(bufferLatest)
	res, err := cfg.HTMLdiff([]string{previousHTML, latestHTML})
	if err != nil {
		fmt.Println(fmt.Errorf("error: %v", err))
		return
	}

	fmt.Println(res[0])
}