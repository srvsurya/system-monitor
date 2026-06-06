package main

import (
	"math"
	"time"
)

func main() {
	// Phase 1 — behave normally for 60 seconds (builds baseline)
	normalEnd := time.Now().Add(60 * time.Second)
	for time.Now().Before(normalEnd) {
		// light work — small math loop
		x := 0.0
		for i := 0; i < 10000; i++ {
			x += math.Sqrt(float64(i))
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Phase 2 — spike hard (triggers anomaly detection)
	for {
		// max CPU — tight loop with no sleep
		x := 0.0
		for i := 0; i < 100000000; i++ {
			x += math.Sqrt(float64(i))
		}
		_ = x
	}
}
