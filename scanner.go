package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"syscall"
	"time"
)

func (a *Application) scanStream(sigs chan os.Signal, scanner *bufio.Scanner, stalk string, amount int) {
	var records []*Record
	tick := time.NewTicker(666 * time.Millisecond)
	fmt.Println("scanning...")
	for scanner.Scan() {
		select {
		case <-tick.C:
			a.storeRecords(records)
		case sig := <-sigs:
			fmt.Println("received a sign that it is time to die")
			switch sig {
			case syscall.SIGINT:
				fmt.Print("sigint")
			case syscall.SIGTERM:
				fmt.Println("sigterm")
			}
			os.Exit(0)
		default:
			var obj Record
			err := json.Unmarshal([]byte(scanner.Text()), &obj)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if obj.Unit == "" {
				obj.Unit = obj.LogID
			}
			records = append(records, &obj)

			if len(a.Data) > 9 {
				a.createWorkload(10)
				a.processWorkload(6)
				if stalk == "" {
					fmt.Print("\033[2J")
					SummarizeResults(a.Result, 25)
				} else {
					fmt.Print("\033[2J")
					a.stalkService(stalk, amount)
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
