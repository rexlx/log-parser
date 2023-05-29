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
	LogID    string `json:"syslog_identifier"`
	Time     uint64 `json:"_source_realtime_timestamp,string"`
	Command  string `json:"_cmdline"`
	Binary   string `json:"_exe"`
	Unit     string `json:"_systemd_unit"`
	Priority int64  `json:"priority,string"`
	Message  string `json:"message"`
}

type Lib struct {
	TotalRecords int
	Mx           *sync.RWMutex
	Data         []*Record
}

var (
	path = flag.String("src", "", "source log or directory")
	// out  = flag.Bool("out", false, "")
)

func main() {
	start := time.Now()
	flag.Parse()
	files := flag.Args()
	results := make(map[string]int)
	library := Lib{
		Data: []*Record{},
		Mx:   &sync.RWMutex{},
	}

	data, err := GetRecords(*path, files, &library)
	if err != nil {
		log.Fatal(err)
	}

	library.TotalRecords = len(data)

	app := Application{
		Data:   data,
		Counts: []Counter{},
		Mx:     &sync.RWMutex{},
		Result: results,
	}

	app.setStep()
	app.Run()
	runtime := time.Since(start)
	app.prettyPrint()
	fmt.Printf("\n\nread %v files and processed %v records in %v seconds\n", len(files), library.TotalRecords, runtime.Seconds())
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
	for _, record := range records {
		if record.Binary == "" && record.LogID != "" {
			record.Binary = record.LogID
		}
		stats[record.Binary]++
	}
	a.syncResults(stats)
	fmt.Println(len(records), "records processed..")
}

func (a *Application) Run() {
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

func readInFile(wg *sync.WaitGroup, path string, lib *Lib) {
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
	lib.storeRecords(records)
}

// func multiRead(records []*Record, err error) {}

// returns a list of files
func GetRecords(path string, files []string, lib *Lib) ([]*Record, error) {
	fInfo, err := os.Stat(path)
	if err != nil {
		return []*Record{}, err
	}
	if fInfo.IsDir() && len(files) > 0 {
		var wg sync.WaitGroup
		wg.Add(len(files))
		for _, fh := range files {
			go readInFile(&wg, filepath.Join(path, fh), lib)
		}
		wg.Wait()
	} else {
		var wg sync.WaitGroup
		wg.Add(1)
		readInFile(&wg, path, lib)
		wg.Wait()
	}
	return lib.Data, nil
}

func SortCounts(counts []Counter) {
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].Occurence > counts[j].Occurence
	})
}

func (l *Lib) storeRecords(records []*Record) {
	l.Mx.Lock()
	defer l.Mx.Unlock()
	l.Data = append(l.Data, records...)

}
