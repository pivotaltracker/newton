package main

import (
	crand "crypto/rand"
	"encoding/json"
	"log"
	"math/big"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"time"

	"code.google.com/p/probab/dst"

	"simonwaldherr.de/go/golibs/xmath"

	"github.com/GaryBoone/GoStats/stats"
)

func init() {
	v, _ := crand.Int(crand.Reader, big.NewInt(1<<62))
	rand.Seed(v.Int64())

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	cycleTimes := map[string][]time.Duration{}
	Check(json.NewDecoder(os.Stdin).Decode(&cycleTimes))

	B := 9999
	confidenceIntervals := map[string]ConfidenceInterval{}
	for category, durations := range cycleTimes {
		sortedDurations := DurationsToHours(durations)
		sort.Float64s(sortedDurations)
		sampleMeans := make([]float64, 0, B)
		l := len(sortedDurations)
		maxIndex := int64(l - 1)
		// rnd := func() int {
		// 	if maxIndex == 0 {
		// 		return 0
		// 	}
		// 	v, _ := crand.Int(crand.Reader, big.NewInt(maxIndex))
		// 	return int(v.Int64())
		// }
		β := dst.Beta(2, 5)
		rnd := func() int {
			v := β()
			// log.Println(v)
			return int(xmath.FloatRound(v*float64(maxIndex), 0))
		}
		// rnd := func() int {
		// 	v := rand.ExpFloat64()
		// 	// log.Println(v)
		// 	return int(xmath.FloatRound(float64(maxIndex)*v/20, 0))
		// }
		for i := 0; i < B; i++ {
			sample := make([]float64, 0, l)
			for j := 0; j < len(sortedDurations); j++ {
				r := rnd()
				sample = append(sample, sortedDurations[r])
			}
			sampleMeans = append(sampleMeans, stats.StatsMean(sample))
		}

		sort.Float64s(sampleMeans)

		optimistic := sampleMeans[499] / 24.0
		pessimistic := sampleMeans[9499] / 24.0
		mean := (sampleMeans[4998] + sampleMeans[4999]) / 48.0

		log.Printf("%s is from %.2f to %.2f days (from %d samples)", category, optimistic, pessimistic, l)
		log.Printf("\tmean is %.2f", mean)
		log.Printf("\tdist from optimistic %.2f", mean-optimistic)
		log.Printf("\tdist from pessimistic %.2f", pessimistic-mean)
		confidenceIntervals[category] = ConfidenceInterval{
			Optimistic:  optimistic,
			Pessimistic: pessimistic,
		}
	}

	Check(json.NewEncoder(os.Stdout).Encode(&confidenceIntervals))
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

func DurationsToHours(ds []time.Duration) []float64 {
	hs := make([]float64, len(ds))
	for i, d := range ds {
		hs[i] = d.Hours()
	}
	return hs
}
