package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"os"
	"runtime"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	tracks := flag.Int("tracks", 1, "The number of development tracks")
	confidenceIntervalsFile := flag.String("ci", "", "JSON file specifiying confidence intervals for story types")
	flag.Parse()

	stories := []string{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		stories = append(stories, scanner.Text())
	}
	Check(scanner.Err())

	confidenceIntervals := map[string]ConfidenceInterval{}
	ciFH, err := os.Open(*confidenceIntervalsFile)
	Check(err)
	Check(json.NewDecoder(ciFH).Decode(&confidenceIntervals))

	oBuckets := make([]float64, *tracks)
	oStories := make([][]string, *tracks)
	pBuckets := make([]float64, *tracks)
	pStories := make([][]string, *tracks)

	for _, story := range stories {
		oMin := MinFloat64Index(oBuckets)
		pMin := MinFloat64Index(pBuckets)

		oBuckets[oMin] += confidenceIntervals[story].Optimistic
		oStories[oMin] = append(oStories[oMin], story)

		pBuckets[pMin] += confidenceIntervals[story].Pessimistic
		pStories[pMin] = append(pStories[pMin], story)
	}

	log.Printf("Stories should take between %.2f and %.2f days", MaxFloat64(oBuckets), MaxFloat64(pBuckets))

	log.Println("Optimistic tracks:")
	for i, stories := range oStories {
		log.Printf("\tTrack %d (%.2f days)", i+1, oBuckets[i])
		for _, story := range stories {
			log.Printf("\t\t%s", story)
		}
	}

	log.Println("Pessimistic tracks:")
	for i, stories := range pStories {
		log.Printf("\tTrack %d (%.2f days)", i+1, pBuckets[i])
		for _, story := range stories {
			log.Printf("\t\t%s", story)
		}
	}
}

type ConfidenceInterval struct {
	Optimistic  float64 `json:"optimistic"`
	Pessimistic float64 `json:"pessimistic"`
}

func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func MinFloat64Index(xs []float64) int {
	i := 0
	for j, x := range xs {
		if x < xs[i] {
			i = j
		}
	}
	return i
}

func MaxFloat64(xs []float64) float64 {
	m := xs[0]
	for _, x := range xs {
		if x > m {
			m = x
		}
	}
	return m
}
