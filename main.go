package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

type wordCounter struct {
	words      map[string]uint
	totalLines uint
	totalWords uint
}

func main() {

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

	type kv struct {
		Key   string
		Value uint
	}

	var ss []kv
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

	fmt.Printf("\nTotal number of uniq words:%10d\n", len(wordCount.words))
	fmt.Printf("Total number of words:%15d\n", wordCount.totalWords)
	fmt.Printf("Total number of lines:%15d\n", wordCount.totalLines)
}
