package main

import "time"

func main() {
	for {
		// burn cpu
		_ = time.Now().UnixNano()
	}
}
