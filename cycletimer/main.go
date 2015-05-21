package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
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
	// confidence := flag.Float64("confidence", 0.95, "for confidence interval calculations")
	flag.Parse()

	log.Printf(
		"Calculating cycle time statistics for project %d using %d past iterations",
		*projectID, *pastIters,
	)

	stories, err := GetStoriesFromPastIterations(*token, *projectID, *pastIters)
	Check(err)

	cycleTimes, err := GetCycleTimes(*token, *projectID, stories)
	Check(err)

	Check(json.NewEncoder(os.Stdout).Encode(&cycleTimes))

	// cycleStats := ComputeCycleTimeStats(cycleTimes, *confidence)

	// for category, cycleStat := range cycleStats {
	// 	log.Printf("%s is from %.2f to %.2f days", category, cycleStat.Optimistic(*confidence)/24.0, cycleStat.Pessimistic(*confidence)/24.0)
	// 	log.Printf("\tMean %.2f hours", cycleStat.Mean)
	// 	log.Printf("\tStdDev %.2f hours", cycleStat.AdjustedStdDev())
	// 	log.Printf("\tSample Size %d", cycleStat.SampleSize)
	// }

	// B := 1999

	// for category, durations := range cycleTimes {
	// 	sortedDurations := durationsToHours(durations)
	// 	sort.Float64s(sortedDurations)
	// 	sampleMeans := make([]float64, 0, B)
	// 	l := len(sortedDurations)
	// 	maxIndex := int64(l - 1)
	// 	// rnd := func() int {
	// 	// 	if maxIndex == 0 {
	// 	// 		return 0
	// 	// 	}
	// 	// 	v, _ := crand.Int(crand.Reader, big.NewInt(maxIndex))
	// 	// 	return int(v.Int64())
	// 	// }
	// 	// β := dst.Beta(2, 5)
	// 	// rnd := func() int {
	// 	// 	return int(xmath.FloatRound(β()*float64(maxIndex), 0))
	// 	// }
	// 	rnd := func() int {
	// 		return int(xmath.FloatRound(float64(maxIndex)*rand.ExpFloat64()/math.MaxFloat64, 0))
	// 	}
	// 	for i := 0; i < B; i++ {
	// 		sample := make([]float64, 0, l)
	// 		for j := 0; j < len(sortedDurations); j++ {
	// 			sample = append(sample, sortedDurations[rnd()])
	// 		}
	// 		sampleMeans = append(sampleMeans, stats.StatsMean(sample))
	// 	}

	// 	sort.Float64s(sampleMeans)

	// 	log.Printf("%s is from %.2f to %.2f days (from %d samples)", category, sampleMeans[100-1]/24.0, sampleMeans[1900-1]/24.0, l)

	// }
}
