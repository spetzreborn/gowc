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

func main() {

	reg, err := regexp.Compile("[^a-zA-Z0-9 ]+")
	if err != nil {
		log.Fatal(err)
	}

	wordCount := make(map[string]int)
	var wordsTotal int
	var linesTotal int

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		words := strings.Fields(strings.ToLower(reg.ReplaceAllString(scanner.Text(), "")))
		linesTotal++
		for _, w := range words {
			wordCount[w]++
			wordsTotal++
		}
	}

	if scanner.Err() != nil {
		// Handle error.
	}

	type kv struct {
		Key   string
		Value int
	}

	var ss []kv
	for k, v := range wordCount {
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

	fmt.Printf("\nTotal number of uniq words:%4d\n", len(wordCount))
	fmt.Printf("Total number of words:%4d\n", wordsTotal)
	fmt.Printf("Total number of lines:%4d\n", linesTotal)
}
