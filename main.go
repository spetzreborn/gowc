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
	totalWords uint
}

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
	/*
		printWordList and printSummary gets inverted so it makes logical names and logical test
	*/
	printWordList = flag.Bool("no-word-list", false, "Don't print the wordlist")
	printSummary  = flag.Bool("no-summary", false, "Don't print the summary")
)

func main() {

	flag.Parse()

	/*
		printWordList and printSummary gets inverted so it makes logical names and logical test
	*/
	*printWordList = !*printWordList
	*printSummary = !*printSummary

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

	linesChan := make(chan string, 1)
	wordsChan := make(chan string, 1)
	wordsCloser := make(chan string)
	wordCount := wordCounter{words: make(map[string]uint)}
	var totalLines uint

	go linesToWords(linesChan, wordsChan, reg)
	go buildWordMap(wordsChan, wordsCloser, &wordCount)

	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		/*
			words := strings.Fields(strings.ToLower(reg.ReplaceAllString(scanner.Text(), "")))
			wordCount.totalLines++
			for _, w := range words {
				wordCount.words[w]++
				wordCount.totalWords++
			}
		*/
		linesChan <- scanner.Text()
		totalLines++
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Invalid input: %s", err)
	}
	close(linesChan)

	_ = <-wordsCloser // wait untill map is build.

	if *printWordList {
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

	if *printSummary {
		fmt.Printf("\nTotal number of uniq words:%10d\n", len(wordCount.words))
		fmt.Printf("Total number of words:%15d\n", wordCount.totalWords)
		fmt.Printf("Total number of lines:%15d\n", totalLines)
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

func linesToWords(linesChan, wordsChan chan string, reg *regexp.Regexp) {
	for {
		line, ok := <-linesChan
		if !ok {
			fmt.Println("DEBUG: Got close, closing wordsChan")
			close(wordsChan)
			return
		}
		words := strings.Fields(strings.ToLower(reg.ReplaceAllString(line, "")))
		for _, word := range words {
			wordsChan <- word
		}
	}
}

func buildWordMap(wordsChan, wordsCloser chan string, wordCount *wordCounter) {
	for {
		w, ok := <-wordsChan
		if !ok {
			fmt.Println("DEBUG: Got close, closing wordsCloser")
			close(wordsCloser)
			return
		}
		wordCount.words[w]++
		wordCount.totalWords++
	}
}
