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
	"sync"
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

	linesChan := make(chan string, 10)
	wordsCloser := make(chan interface{})
	wordCount := wordCounter{words: make(map[string]uint)}
	var totalLines uint

	c1 := linesToWords(linesChan, reg)
	c2 := linesToWords(linesChan, reg)
	c3 := linesToWords(linesChan, reg)
	wordsChan := merge(c1, c2, c3)

	go buildWordMap(wordsChan, wordsCloser, &wordCount)

	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		linesChan <- scanner.Text()
		totalLines++
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Invalid input: %s", err)
	}
	close(linesChan)

	<-wordsCloser // wait untill map is build.

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

func linesToWords(linesChan chan string, reg *regexp.Regexp) <-chan string {
	out := make(chan string)

	fmt.Println("DEBUG: starting goroutine")
	go func() {
		for line := range linesChan {
			words := strings.Fields(strings.ToLower(reg.ReplaceAllString(line, "")))
			for _, word := range words {
				out <- word
			}
		}
		fmt.Println("DEBUG: No more input, leaving gorutine")
		close(out)
	}()
	return out
}

func buildWordMap(wordsChan <-chan string, wordsCloser chan interface{}, wordCount *wordCounter) {
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

func merge(cs ...<-chan string) <-chan string {
	var wg sync.WaitGroup
	out := make(chan string, 200)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan string) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
