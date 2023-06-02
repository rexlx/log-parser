package main

import (
	"bufio"
	"encoding/json"
	"fmt"
)

func ScanStream(scanner *bufio.Scanner) {
	// scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var obj Record
		err := json.Unmarshal([]byte(scanner.Text()), &obj)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(obj.Unit)
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
}
