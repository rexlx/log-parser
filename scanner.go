package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"syscall"
	"time"
)

func (a *Application) scanStream(sigs chan os.Signal, scanner *bufio.Scanner) {
	var records []*Record
	tick := time.NewTicker(150 * time.Millisecond)
	fmt.Println("scanning...")
	for scanner.Scan() {
		select {
		case <-tick.C:
			if len(records) > 0 {
				a.Data = append(a.Data, records...)
				records = nil
			}
		case sig := <-sigs:
			fmt.Println("\nreceived a sign that it is time to die")
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

			if len(a.Data) > 10 {
				a.createWorkload(10)
				a.processWorkload(*level)
				if *stalk == "" {
					// fmt.Print("\033[2J")
					a.printToScreen(SummarizeResults(a.Result, *amount))
				} else {
					// fmt.Print("\033[2J")
					a.stalkService(*stalk, *amount)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
}
