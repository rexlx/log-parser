package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"time"
)

func (a *Application) scanStream(scanner *bufio.Scanner) {
	var records []*Record
	tick := time.NewTicker(666 * time.Millisecond)
	end := time.After(666 * time.Minute)
	for scanner.Scan() {
		select {
		case <-tick.C:
			a.storeRecords(records)
			records = nil
			a.getStats(a.Data, 5)
			// a.summarizeResults(25)
		case <-end:
			fmt.Println("im too old for this shit...")
		default:
			fmt.Printf("\rgot: %d", len(a.Data))
			var obj Record
			err := json.Unmarshal([]byte(scanner.Text()), &obj)
			if err != nil {
				fmt.Println(err)
				continue
			}
			records = append(records, &obj)
			// continue
			// tick.Stop()
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
}
