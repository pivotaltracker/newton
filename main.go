package main

import (
	"flag"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/mceldeen/aero/ratelimit"
)

var client *http.Client = nil

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	rlConfig := ratelimit.NewBursty(300, time.Minute, 15)
	client = &http.Client{
		Transport: ratelimit.NewHttpTransport(rlConfig),
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	projectID := flag.Int("project", 0, "The project id to load")
	token := flag.String("token", "", "The tok")
	pastIters := flag.Int("past", 3, "Past iterations to use for cycle time calculations")
	confidence := flag.Float64("confidence", 0.95, "for confidence interval calculations")
	flag.Parse()

	log.Printf(
		"Calculating cycle time statistics for project %d using %d past iterations",
		*projectID, *pastIters,
	)

	stories, err := GetStoriesFromPastIterations(*token, *projectID, *pastIters)
	Check(err)

	cycleTimes, err := GetCycleTimes(*token, *projectID, stories)
	Check(err)

	cycleStats := ComputeCycleTimeStats(cycleTimes, *confidence)

	α := 1.0 - *confidence
	for category, cycleStat := range cycleStats {
		log.Printf("%s is from %.2f to %.2f days", category, cycleStat.Optimistic(α)/24.0, cycleStat.Pessimistic(α)/24.0)
		log.Printf("\tMean %.2f hours", cycleStat.Mean)
		log.Printf("\tStdDev %.2f hours", cycleStat.AdjustedStdDev())
		log.Printf("\tSample Size %d", cycleStat.SampleSize)
	}
}

func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
