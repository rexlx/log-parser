package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Counter struct {
	Name      string  `json:"unit_name"`
	Occurence int     `json:"occurence"`
	Percent   float64 `json:"percent"`
}

type Application struct {
	WorkLoad [][]*Record
	Data     []*Record
	Counts   []Counter
	Result   map[string]int
	Mx       *sync.RWMutex
}

type Record struct {
	LogID    string      `json:"syslog_identifier"`
	Time     uint64      `json:"_source_realtime_timestamp,string"`
	Command  string      `json:"_cmdline"`
	Binary   string      `json:"_exe"`
	Unit     string      `json:"_systemd_unit"`
	Priority int64       `json:"priority,string"`
	Message  interface{} `json:"message"`
}

var (
	path = flag.String("src", "", "source log or directory")
)

func main() {
	start := time.Now()
	flag.Parse()
	files := flag.Args()
	results := make(map[string]int)

	app := Application{
		Data:   []*Record{},
		Counts: []Counter{},
		Mx:     &sync.RWMutex{},
		Result: results,
	}

	// var wg sync.WaitGroup

	fileList := WalkFiles(files, 10)
	for _, i := range fileList {
		app.getRecords(*path, i)
	}
	// 	wg.Add(1)
	// 	go func(path string, files []string) {
	// 		defer wg.Done()
	// 		app.getRecords(path, files)
	// 	}(*path, i)

	// wg.Wait()

	app.setStep()

	app.run()

	runtime := time.Since(start)

	app.prettyPrint()

	fmt.Printf("\n\nread %v files and processed %v records in %v seconds\n", len(files), app.Result["_total"], runtime.Seconds())
	fmt.Printf("detected %v errors (priority<5)\t%v%v\n", app.Result["_error"], float64(app.Result["_error"])/float64(app.Result["_total"])*100, "%")
}

func (a *Application) setStep() {

	var allData [][]*Record
	total := len(a.Data)
	chunkSize := total / 8
	for i := 0; i < total; i += chunkSize {
		end := i + chunkSize

		if end > total {
			end = total
		}
		allData = append(allData, a.Data[i:end])
	}
	a.WorkLoad = allData
}

func (a *Application) syncResults(result map[string]int) {
	a.Mx.Lock()
	defer a.Mx.Unlock()
	for k, v := range result {
		a.Result[k] += v
	}
}

func (a *Application) getStats(wg *sync.WaitGroup, records []*Record) {
	defer wg.Done()
	stats := make(map[string]int)
	stats["_error"] = 0
	stats["_total"] = 0
	for _, record := range records {
		// _ = InterfaceToByteSlice(record.Message)
		if record.Priority < 5 {
			stats["_error"]++
		}
		stats[record.Unit]++
		stats["_total"]++
	}
	a.syncResults(stats)
	fmt.Println(len(records), "records processed..")
}

func (a *Application) run() {
	var wg sync.WaitGroup
	for _, job := range a.WorkLoad {
		wg.Add(1)
		go a.getStats(&wg, job)
	}
	wg.Wait()
}

func (a *Application) prettyPrint() {
	for k, v := range a.Result {
		a.Counts = append(a.Counts, Counter{
			Name:      k,
			Occurence: v,
			Percent:   (float64(v) / float64(len(a.Data)) * 100),
		})

	}
	SortCounts(a.Counts)
	out, err := json.Marshal(a.Counts)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(string(out))
}

func readInFile(wg *sync.WaitGroup, path string, f func(r []*Record)) {
	// var c int
	defer wg.Done()
	var records []*Record
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		// c++
		var r Record
		if err := json.Unmarshal(line, &r); err != nil {
			fmt.Println("json marshalling issue:", err)
			continue
		}
		records = append(records, &r)
	}
	f(records)
}

// returns a list of files
func (a *Application) getRecords(path string, files []string) {
	fInfo, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	if fInfo.IsDir() && len(files) > 0 {
		var wg sync.WaitGroup
		wg.Add(len(files))
		for _, fh := range files {
			go readInFile(&wg, filepath.Join(path, fh), a.storeRecords)
		}
		wg.Wait()
	} else {
		var wg sync.WaitGroup
		wg.Add(1)
		readInFile(&wg, path, a.storeRecords)
		wg.Wait()
	}
}

func (a *Application) storeRecords(records []*Record) {
	a.Mx.Lock()
	defer a.Mx.Unlock()
	a.Data = append(a.Data, records...)

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
		fmt.Println(string(out), "bleh")
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
