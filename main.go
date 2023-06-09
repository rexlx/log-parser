package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

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
	// mark the start time
	start := time.Now()

	// get the flags
	flag.Parse()

	// the application type will need two maps like this
	results := make(map[string]int)
	serviceDetails := make(map[string]int)

	// init the app
	app := Application{
		Data:           []*Record{},
		Counts:         []Counter{},
		Mx:             &sync.RWMutex{},
		Result:         results,
		ServiceDetails: serviceDetails,
	}

	app.Stats.Start = time.Now()

	// the scanner will run forever unless it catches a sigint or sigterm,
	// in which case it exits with 0; the code below *should* not run (...)
	if *scan {
		scanner := bufio.NewScanner(os.Stdin)
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		app.scanStream(sigs, scanner)
	} else {
		fileList := WalkFiles(*path, *rate)

		for _, i := range fileList {
			// we block here instead using go routines for the sake of the cpu
			app.getRecords(i)
		}

		app.createWorkload(*step)
		app.processWorkload(*level)
		app.stalkService(*stalk, *amount)
		app.printToScreen(SummarizeResults(app.Result, *amount))

		fmt.Printf("\nprocessed %v records in %v seconds\n", app.Result["_total"], time.Since(start).Seconds())
	}

	fmt.Println("done.")
}
