package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Application struct {
	WorkLoad       [][]*Record
	ServiceDetails map[string]int
	Data           []*Record
	Counts         []Counter
	Result         map[string]int
	Stats          Stats
	Mx             *sync.RWMutex
}

type Stats struct {
	Start   time.Time
	Now     time.Time
	Runtime time.Duration
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

func (a *Application) createWorkload(size int) {
	a.Mx.Lock()
	defer a.Mx.Unlock()
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
	a.Data = nil
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
		if record.Priority <= level {
			stats["_error"]++
		}
		stats[record.Unit]++
		stats["_total"]++
	}
	a.syncResults(stats)
	// fmt.Println(len(records), "records processed..")
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

func (a *Application) stalkService(service string, amount int) {
	if service != "" {
		var msg string
		var pairs []Counter
		var wg sync.WaitGroup
		for _, load := range a.WorkLoad {
			wg.Add(1)
			go func(load []*Record, svc string) {
				defer wg.Done()
				vals := make(map[string]int)
				for _, i := range load {
					if i.Unit == svc {
						vals[fmt.Sprintf("%v -%v", i.Priority, string(InterfaceToByteSlice(i.Message)))]++
					}
				}
				a.syncServiceCounter(vals)
			}(load, service)
		}
		wg.Wait()
		// sort.Slice(counts, func(i, j int) bool {
		// 	return counts[i].Occurence > counts[j].Occurence
		// })
		for k, v := range a.ServiceDetails {
			pairs = append(pairs, Counter{
				Name:      k,
				Occurence: v,
			})
		}
		SortCounts(pairs)
		if amount > len(pairs) {
			amount = len(pairs)
		}
		maxLen := GetMaxLen(pairs[0:amount])
		if maxLen > 16 {
			maxLen = 16
		}
		for _, i := range pairs[0:amount] {
			msg += fmt.Sprintf("%-*s %v\n", maxLen, i.Name, i.Occurence)
		}
		a.printToScreen(msg)
	}
}

func (a *Application) printToScreen(msg string) {
	a.Stats.Now = time.Now()
	a.Stats.Runtime = time.Since(a.Stats.Start)
	fmt.Print("\033[2J")
	header := fmt.Sprintf("Initialized: %v | Runtime: %v | Date: %v\n\n", a.Stats.Start.Format(time.RFC822), a.Stats.Runtime, a.Stats.Now.Format(time.RFC822))
	footer := strings.Repeat("_", len(header))
	fmt.Printf("\n%v\n%v\n%v", header, msg, footer)
}
