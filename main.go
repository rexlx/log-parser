package main

import (
	"bufio"
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
	WorkLoad       [][]*Record
	ServiceDetails map[string]int
	Data           []*Record
	Counts         []Counter
	Result         map[string]int
	Mx             *sync.RWMutex
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
	rate   = flag.Int("read", 5, "how many files to read in at once")
	path   = flag.String("src", "", "source log or directory")
	stalk  = flag.String("stalk", "", "systemd unit to follow")
	step   = flag.Int("step", 8, "chunk size")
	amount = flag.Int("show", 25, "amount of stats to show")
	level  = flag.Int64("level", 5, "error level")
	scan   = flag.Bool("scan", false, "read from log stream")
)

func main() {
	start := time.Now()

	flag.Parse()
	files := flag.Args()

	results := make(map[string]int)
	serviceDetails := make(map[string]int)

	app := Application{
		Data:           []*Record{},
		Counts:         []Counter{},
		Mx:             &sync.RWMutex{},
		Result:         results,
		ServiceDetails: serviceDetails,
	}

	if !*scan {
		fileList := WalkFiles(files, *rate)

		for _, i := range fileList {
			app.getRecords(*path, i)
		}

		app.createWorkload(*step)
		app.processWorkload(*level)
		app.stalkService(*stalk)
		app.summarizeResults(app.Result, *amount)

		fmt.Printf("\n\nread %v files and processed %v records in %v seconds\n", len(files), app.Result["_total"], time.Since(start).Seconds())
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		app.scanStream(scanner, *stalk)
	}
}

func (a *Application) createWorkload(size int) {
	var allData [][]*Record
	total := len(a.Data)
	if total < size {
		allData = append(allData, a.Data)
	} else {
		chunkSize := total / size
		for i := 0; i < total; i += chunkSize {
			end := i + chunkSize

			if end > total {
				end = total
			}
			allData = append(allData, a.Data[i:end])
		}
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

func (a *Application) getStats(records []*Record, level int64) {
	stats := make(map[string]int)
	stats["_error"] = 0
	stats["_total"] = 0

	for _, record := range records {
		// _ = InterfaceToByteSlice(record.Message)
		if record.Priority < level {
			stats["_error"]++
		}
		stats[record.Unit]++
		stats["_total"]++
	}
	a.syncResults(stats)
	// fmt.Println(len(records), "records processed..")
}

func (a *Application) processWorkload(level int64) {
	var wg sync.WaitGroup
	for _, job := range a.WorkLoad {
		wg.Add(1)
		go func(job []*Record, level int64) {
			defer wg.Done()
			a.getStats(job, level)
		}(job, level)
	}
	wg.Wait()
}

func GetMaxLen(c []Counter) int {
	l := c[0]
	for _, s := range c {
		if len(s.Name) > len(l.Name) {
			l = s
		}
	}
	return len(l.Name)
}

func MaxLen(arr []string) int {
	x := arr[0]
	for _, item := range arr {
		if len(item) > len(x) {
			x = item
		}
	}
	return len(x)
}

func (a *Application) summarizeResults(results map[string]int, amount int) {
	if len(results) < 1 {
		return
	}
	var counts []Counter
	for k, v := range results {
		p := (float64(v) / float64(len(a.Data)) * 100)

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
	if maxLen > 16 {
		maxLen = 16
	}
	for _, i := range counts[0:amount] {
		fmt.Printf("%-*s %v\t%v\n", maxLen, i.Name, i.Occurence, i.Percent)
	}
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

func (a *Application) syncServiceCounter(sc map[string]int) {
	a.Mx.Lock()
	defer a.Mx.Unlock()
	for k, v := range sc {
		a.ServiceDetails[k] += v
	}
}

func (a *Application) stalkService(service string) {
	if service != "" {
		var wg sync.WaitGroup
		for _, load := range a.WorkLoad {
			wg.Add(1)
			go func(l []*Record, s string) {
				defer wg.Done()
				vals := make(map[string]int)
				for _, i := range l {
					if i.Unit == s {
						vals[fmt.Sprintf("%v -%v", i.Priority, string(InterfaceToByteSlice(i.Message)))]++
					}
				}
				a.syncServiceCounter(vals)
			}(load, service)
		}
		wg.Wait()
	}
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
		records = append(records, &r)
	}
	f(records)
}
