package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
)

func GetMaxLen(c []Counter) int {
	if len(c) > 0 {
		l := c[0]
		for _, s := range c {
			if len(s.Name) > len(l.Name) {
				l = s
			}
		}
		return len(l.Name)
	}
	return 0
}

func InterfaceToByteSlice(i interface{}) []byte {
	arr, ok := i.([]interface{})
	str, ko := i.(string)
	if ok {
		var out []byte
		for _, item := range arr {
			x, ok := item.(float64)
			if ok {
				out = append(out, uint8(x))
			}
		}
		return out
	}
	if ko {
		return []byte(str)
	}
	return []byte{}
}

func SortCounts(counts []Counter) {
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].Occurence > counts[j].Occurence
	})
}

func WalkFiles(files []string, step int) [][]string {
	var fileList [][]string
	if len(files) <= step {
		fileList = append(fileList, files)
		return fileList
	}
	total := len(files)
	for i := 0; i < total; i += step {
		end := i + step

		if end > total {
			end = total
		}
		fileList = append(fileList, files[i:end])
	}
	return fileList
}

func readInFile(wg *sync.WaitGroup, path string, f func(r []*Record)) {
	defer wg.Done()
	var records []*Record
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		var r Record
		if err := json.Unmarshal(line, &r); err != nil {
			fmt.Println("json marshalling issue:", err)
			continue
		}
		if r.Unit == "" {
			r.Unit = r.LogID
		}
		records = append(records, &r)
	}
	f(records)
}

func SummarizeResults(results map[string]int, amount int) string {
	var msg string
	if len(results) < 1 {
		fmt.Println("empty results")
		return "empty results for this period"
	}
	var counts []Counter
	var total int

	for k, v := range results {
		if k == "_total" || k == "_error" {
			continue
		}
		total += v
	}

	for k, v := range results {
		if k == "_total" || k == "_error" {
			continue
		}
		p := (float64(v) / float64(total) * 100)

		counts = append(counts, Counter{
			Name:      k,
			Occurence: v,
			Percent:   p,
		})

	}
	SortCounts(counts)

	if amount > len(counts) {
		amount = len(counts)
	}
	maxLen := GetMaxLen(counts[0:amount])
	if maxLen > 31 {
		maxLen = 31
	}
	for _, i := range counts[0:amount] {
		msg += fmt.Sprintf("%-*s %*v > %v\n", maxLen, i.Name, 8, i.Occurence, i.Percent)
	}
	return msg
}
