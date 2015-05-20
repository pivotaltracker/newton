package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"runtime"
	"sync"
	"time"

	dst "code.google.com/p/probab/dst"
	gstats "github.com/GaryBoone/GoStats/stats"
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

	stories, err := getStoriesFromPastIterations(*token, *projectID, *pastIters)
	Check(err)

	cycleTimes, err := getCycleTimes(*token, *projectID, stories)
	Check(err)

	cycleStats := computeCycleTimeStats(cycleTimes, *confidence)

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

type CycleTimeStat struct {
	Mean       float64 `json:"mean"`    // hours
	StdDev     float64 `json:"std_dev"` // hours
	SampleSize int     `json:"sample_size"`
}

func (s CycleTimeStat) Optimistic(α float64) float64 {
	z := s.z(α)
	return s.Mean - z*minFloat64(s.σ()/math.Sqrt(float64(s.SampleSize)), s.Mean/(4*z))
}

func (s CycleTimeStat) Pessimistic(α float64) float64 {
	return s.Mean + s.z(α)*s.σ()/math.Sqrt(float64(s.SampleSize))
}

func (s CycleTimeStat) z(α float64) float64 {
	return dst.BetaQtlFor(5.0, 1.0, 1.0-α/2.0)
}

func (s CycleTimeStat) σ() float64 {
	if math.IsNaN(s.StdDev) {
		return s.Mean / 2.0
	}
	return s.StdDev
}

func (s CycleTimeStat) AdjustedStdDev() float64 {
	return s.σ()
}

func minFloat64(xs ...float64) float64 {
	m := xs[0]
	for _, x := range xs {
		if x < m {
			m = x
		}
	}
	return m
}

func maxFloat64(xs ...float64) float64 {
	m := xs[0]
	for _, x := range xs {
		if x > m {
			m = x
		}
	}
	return m
}

func maxUnder(under float64, xs ...float64) float64 {
	m := xs[0]
	if m >= under {
		for _, x := range xs {
			if x < under {
				m = x
				break
			}
		}
	}
	for _, x := range xs {
		if x > m && x < under {
			m = x
		}
	}
	return m
}

func computeCycleTimeStats(cycleTimes map[string][]time.Duration, confidence float64) map[string]CycleTimeStat {
	stats := make(map[string]CycleTimeStat, len(cycleTimes))
	for key, durations := range cycleTimes {
		data := qTest(durationsToHours(durations), confidence)
		set := gstats.Stats{}
		set.UpdateArray(data)
		stat := CycleTimeStat{
			Mean:       set.Mean(),
			StdDev:     set.SampleStandardDeviation(),
			SampleSize: len(durations),
		}
		stats[key] = stat
	}
	return stats
}

func durationsToHours(ds []time.Duration) []float64 {
	hs := make([]float64, 0, len(ds))
	for _, d := range ds {
		hs = append(hs, float64(d)/float64(time.Hour))
	}
	return hs
}

func qTest(xs []float64, confidence float64) []float64 {
	if len(xs) < 3 {
		return xs
	}

	min := minFloat64(xs...)
	max := maxFloat64(xs...)
	maxUnder := maxUnder(max, xs...)
	gap := max - maxUnder
	r := max - min

	if gap/r <= qCrit(len(xs), int(confidence*100.0)) {
		return xs
	}

	ns := make([]float64, 0, len(xs)-1)
	found := false
	for _, x := range xs {
		if !found && x == max {
			found = true
			continue
		}
		ns = append(ns, x)
	}
	return ns
}

func qCrit(n, confidence int) float64 {
	if n > 10 {
		n = 10
	}

	switch {
	case confidence < 95:
		confidence = 90
	case confidence < 99:
		confidence = 95
	default:
		confidence = 99
	}

	return qTable[confidence][n]
}

// Assumes that the first transition is creation (uncreated -> *)
// Assumes that the story has been accepted
func getCycleTime(transitions []StoryTransition) time.Duration {
	d := time.Duration(0)
	f := dontSumCycleTime
	for i := 1; i < len(transitions); i++ {
		f, d = f(transitions[i], transitions[i-1], d)
	}
	return d
}

type CycleTimeFn func(this StoryTransition, last StoryTransition, d time.Duration) (CycleTimeFn, time.Duration)

func dontSumCycleTime(transition StoryTransition, last StoryTransition, d time.Duration) (CycleTimeFn, time.Duration) {
	// log.Printf("outside %s -> %s", transition.FromState, transition.ToState)
	if transition.ToState != "unstarted" && transition.ToState != "unscheduled" {
		return sumCycleTime, d
	}
	return dontSumCycleTime, d
}

func sumCycleTime(transition StoryTransition, last StoryTransition, d time.Duration) (CycleTimeFn, time.Duration) {
	// log.Printf("inside %s -> %s", transition.FromState, transition.ToState)
	d = d + transition.OccurredAt.Sub(last.OccurredAt)

	if transition.ToState == "unstarted" || transition.ToState == "unscheduled" || transition.ToState == "accepted" || transition.ToState == "finished" {
		return dontSumCycleTime, d
	}

	return sumCycleTime, d
}

func getCycleTimes(token string, projectID int, stories []Story) (map[string][]time.Duration, error) {
	cycleTimes := map[string][]time.Duration{}
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for _, story := range stories {
		if story.Type == "release" {
			continue
		}
		wg.Add(1)
		go func(story Story) {
			defer wg.Done()
			transitions, err := getTransitions(token, projectID, story.ID)
			if err != nil {
				return
			}
			if len(transitions) == 0 {
				return
			}

			if transitions[0].FromState != "uncreated" || transitions[len(transitions)-1].ToState != "accepted" {
				return
			}

			key := story.Key()
			ct := getCycleTime(transitions)
			// log.Println(key, story.ID, ct)

			mu.Lock()
			if _, exists := cycleTimes[key]; !exists {
				cycleTimes[key] = []time.Duration{}
			}
			cycleTimes[key] = append(cycleTimes[key], ct)
			mu.Unlock()

		}(story)
	}

	wg.Wait()

	return cycleTimes, nil
}

// returns transitions in chronological order
func getTransitions(token string, projectID int, storyID int) ([]StoryTransition, error) {
	uri := fmt.Sprintf(
		"https://www.pivotaltracker.com/services/v5/projects/%d/stories/%d/activity?limit=500&fields=occurred_at,changes",
		projectID,
		storyID,
	)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return []StoryTransition{}, err
	}
	req.Header.Set("X-TrackerToken", token)

	res, err := client.Do(req)
	if err != nil {
		return []StoryTransition{}, err
	}
	defer res.Body.Close()

	items := []ActivityItem{}

	err = json.NewDecoder(res.Body).Decode(&items)
	if err != nil {
		return []StoryTransition{}, err
	}

	transitions := make([]StoryTransition, 0, len(items))
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		for _, change := range item.Changes {
			if change.Kind == "story" {
				if toState, ok := change.NewValues["current_state"]; ok {
					fromState := "uncreated"
					if change.OriginalValues["current_state"] != nil {
						fromState = change.OriginalValues["current_state"].(string)
					}
					transitions = append(transitions, StoryTransition{
						FromState:  fromState,
						ToState:    toState.(string),
						OccurredAt: item.OccurredAt,
					})
				}
			}
		}
	}

	return transitions, nil
}

func getStoriesFromPastIterations(token string, projectID int, pastIters int) ([]Story, error) {
	uri := fmt.Sprintf(
		"https://www.pivotaltracker.com/services/v5/projects/%d/iterations?scope=done&offset=%d&fields=stories(id,story_type,estimate)",
		projectID,
		-pastIters,
	)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return []Story{}, err
	}

	req.Header.Set("X-TrackerToken", token)

	res, err := client.Do(req)
	if err != nil {
		return []Story{}, err
	}
	defer res.Body.Close()

	iterations := []Iteration{}

	err = json.NewDecoder(res.Body).Decode(&iterations)
	if err != nil {
		return []Story{}, err
	}

	stories := make([]Story, 0, len(iterations))
	for _, iteration := range iterations {
		stories = append(stories, iteration.Stories...)
	}

	return stories, nil
}

type Iteration struct {
	Stories []Story `json:"stories"`
}

type Story struct {
	ID       int    `json:"id"`
	Type     string `json:"story_type"`
	Estimate int    `json:"estimate"`
}

func (s *Story) Key() string {
	return fmt.Sprintf("%s:%d", s.Type, s.Estimate)
}

type StoryTransition struct {
	FromState  string
	ToState    string
	OccurredAt time.Time
}

type ActivityItem struct {
	OccurredAt time.Time        `json:"occurred_at"`
	Changes    []ActivityChange `json:"changes"`
}

type ActivityChange struct {
	Kind           string                 `json:"kind"`
	ID             int                    `json:"id"`
	OriginalValues map[string]interface{} `json:"original_values"`
	NewValues      map[string]interface{} `json:"new_values"`
}

var qTable = map[int]map[int]float64{
	90: map[int]float64{
		3:  0.941,
		4:  0.765,
		5:  0.642,
		6:  0.560,
		7:  0.507,
		8:  0.468,
		9:  0.437,
		10: 0.412,
	},
	95: map[int]float64{
		3:  0.970,
		4:  0.829,
		5:  0.710,
		6:  0.625,
		7:  0.568,
		8:  0.526,
		9:  0.493,
		10: 0.466,
	},
	99: map[int]float64{
		3:  0.994,
		4:  0.926,
		5:  0.821,
		6:  0.740,
		7:  0.680,
		8:  0.634,
		9:  0.598,
		10: 0.568,
	},
}
