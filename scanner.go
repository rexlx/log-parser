package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"time"
)

func (a *Application) scanStream(scanner *bufio.Scanner, stalk string) {
	var records []*Record
	tick := time.NewTicker(666 * time.Millisecond)
	end := time.After(666 * time.Minute)
	fmt.Println("scanning...")
	for scanner.Scan() {
		select {
		case <-tick.C:
			a.storeRecords(records)
			a.createWorkload(10)
			a.processWorkload(6)
			records = nil
			if stalk == "" {
				a.summarizeResults(a.Result, 25)
			} else {
				fmt.Print("\033[2J")
				a.stalkService(stalk)
				a.summarizeResults(a.ServiceDetails, 5)
				// a.summarizeResults(6)
			}
		case <-end:
			fmt.Println("im too old for this shit...")
		default:
			// fmt.Printf("\rgot: %d", len(a.Data))
			var obj Record
			err := json.Unmarshal([]byte(scanner.Text()), &obj)
			if err != nil {
				fmt.Println(err)
				continue
			}
			records = append(records, &obj)
			// time.Sleep(25 * time.Millisecond)
			// tick.Stop()
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
}
