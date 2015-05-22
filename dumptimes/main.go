package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

func main() {
	cycleTimes := map[string][]time.Duration{}
	Check(json.NewDecoder(os.Stdin).Decode(&cycleTimes))

	times := make([]float64, 0, len(cycleTimes)*10)

	for _, ts := range cycleTimes {
		for _, t := range ts {
			if t.Minutes() >= 5.0 {
				times = append(times, t.Hours())
			}
		}
	}

	sort.Float64s(times)

	for _, time := range times {
		fmt.Printf("%.3f\r\n", time)
	}
}

func Check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
