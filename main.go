package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
)

type wordCounter struct {
	words      map[string]uint
	totalLines uint
	totalWords uint
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
var noWordList = flag.Bool("no-word-list", false, "Don't print the wordlist")
var noSummery = flag.Bool("no-summery", false, "Don't print the summary")

func main() {

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	reg, err := regexp.Compile("[^a-zA-Z0-9 ]+")
	if err != nil {
		log.Fatal(err)
	}

	wordCount := wordCounter{words: make(map[string]uint)}

	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		words := strings.Fields(strings.ToLower(reg.ReplaceAllString(scanner.Text(), "")))
		wordCount.totalLines++
		for _, w := range words {
			wordCount.words[w]++
			wordCount.totalWords++
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Invalid input: %s", err)
	}

	if *noWordList == false {
		type kv struct {
			Key   string
			Value uint
		}

		ss := make([]kv, 0, len(wordCount.words))
		for k, v := range wordCount.words {
			ss = append(ss, kv{k, v})
		}

		sort.Slice(ss, func(i, j int) bool {

			if ss[i].Value != ss[j].Value {
				return ss[i].Value < ss[j].Value
			}
			return ss[i].Key < ss[j].Key

		})

		for _, kv := range ss {
			fmt.Printf("%d: %s\n", kv.Value, kv.Key)
		}

	}

	if *noSummery == false {
		fmt.Printf("\nTotal number of uniq words:%10d\n", len(wordCount.words))
		fmt.Printf("Total number of words:%15d\n", wordCount.totalWords)
		fmt.Printf("Total number of lines:%15d\n", wordCount.totalLines)
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}
