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

	B := 1999
	confidenceIntervals := map[string]ConfidenceInterval{}
	for category, durations := range cycleTimes {
		sortedDurations := DurationsToHours(durations)
		sort.Float64s(sortedDurations)
		sampleMedians := make([]float64, 0, B)
		l := len(sortedDurations)
		maxIndex := int64(l - 1)
		// rnd := func() int {
		// 	if maxIndex == 0 {
		// 		return 0
		// 	}
		// 	v, _ := crand.Int(crand.Reader, big.NewInt(maxIndex))
		// 	return int(v.Int64())
		// }
		β := dst.Beta(1, 14)
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
			sampleMedians = append(sampleMedians, Median(sample))
		}

		sort.Float64s(sampleMedians)
		// log.Println(sampleMedians)

		optimistic := sampleMedians[99] / 24.0
		pessimistic := sampleMedians[1899] / 24.0
		median := Median(sampleMedians) / 24.0

		log.Printf("%s is from %.2f to %.2f days (from %d samples)", category, optimistic, pessimistic, l)
		log.Printf("\tmedian is %.2f", median)
		log.Printf("\tdist from optimistic %.2f", median-optimistic)
		log.Printf("\tdist from pessimistic %.2f", pessimistic-median)
		confidenceIntervals[category] = ConfidenceInterval{
			Optimistic:  optimistic,
			Pessimistic: pessimistic,
			Median:      median,
		}
	}

	Check(json.NewEncoder(os.Stdout).Encode(&confidenceIntervals))
}

type ConfidenceInterval struct {
	Optimistic  float64 `json:"optimistic"`
	Pessimistic float64 `json:"pessimistic"`
	Median      float64 `json:"median"`
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

func Median(xs []float64) float64 {
	switch {
	case len(xs) == 0:
		return 0.0
	case len(xs) == 1:
		return xs[0]
	case len(xs)%2 == 0:
		i := len(xs)/2 - 1
		return (xs[i] + xs[i+1]) / 2.0
	default:
		i := len(xs) / 2
		return xs[i]
	}
}
