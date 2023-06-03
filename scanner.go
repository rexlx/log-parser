package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"time"
)

func (a *Application) scanStream(scanner *bufio.Scanner, stalk string, amount int) {
	var records []*Record
	// var c int
	tick := time.NewTicker(666 * time.Millisecond)
	end := time.After(666 * time.Minute)
	fmt.Println("scanning...")
	for scanner.Scan() {
		select {
		case <-tick.C:
			a.storeRecords(records)
		case <-end:
			fmt.Println("im too old for this shit...")
		default:
			// c++
			// fmt.Println("TOTAL:", c)
			var obj Record
			err := json.Unmarshal([]byte(scanner.Text()), &obj)
			if err != nil {
				fmt.Println(err)
				continue
			}
			records = append(records, &obj)

			if len(a.Data) > 9 {
				a.createWorkload(10)
				a.processWorkload(6)
				if stalk == "" {
					fmt.Print("\033[2J")
					a.summarizeResults(a.Result, 25)
				} else {
					fmt.Print("\033[2J")
					a.stalkService(stalk, amount)
					// a.summarizeResults(a.ServiceDetails, 5)
					// a.summarizeResults(6)
				}
				records = nil
			}
			// time.Sleep(25 * time.Millisecond)
			// tick.Stop()
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
}
