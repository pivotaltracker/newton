package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func GetCycleTimesByCategory(token string, projectID int, stories []Story) (map[string][]time.Duration, error) {
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
				log.Println(err)
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
func getTransitions(token string, projectID int, storyID int) ([]storyTransition, error) {
	uri := fmt.Sprintf(
		"https://www.pivotaltracker.com/services/v5/projects/%d/stories/%d/activity?limit=500&fields=occurred_at,changes",
		projectID,
		storyID,
	)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return []storyTransition{}, err
	}
	req.Header.Set("X-TrackerToken", token)

	res, err := client.Do(req)
	if err != nil {
		return []storyTransition{}, err
	}
	defer res.Body.Close()

	items := []activityItem{}

	err = json.NewDecoder(res.Body).Decode(&items)
	if err != nil {
		return []storyTransition{}, err
	}

	transitions := make([]storyTransition, 0, len(items))
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		for _, change := range item.Changes {
			if change.Kind == "story" {
				if toState, ok := change.NewValues["current_state"]; ok {
					fromState := "uncreated"
					if change.OriginalValues["current_state"] != nil {
						fromState = change.OriginalValues["current_state"].(string)
					}
					transitions = append(transitions, storyTransition{
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

type storyTransition struct {
	FromState  string
	ToState    string
	OccurredAt time.Time
}

type activityItem struct {
	OccurredAt time.Time        `json:"occurred_at"`
	Changes    []activityChange `json:"changes"`
}

type activityChange struct {
	Kind           string                 `json:"kind"`
	ID             int                    `json:"id"`
	OriginalValues map[string]interface{} `json:"original_values"`
	NewValues      map[string]interface{} `json:"new_values"`
}

// Assumes that the first transition is creation (uncreated -> *)
// Assumes that the story has been accepted
func getCycleTime(transitions []storyTransition) time.Duration {
	d := time.Duration(0)
	f := dontSumCycleTime
	for i := 1; i < len(transitions); i++ {
		f, d = f(transitions[i], transitions[i-1], d)
	}
	return d
}

type CycleTimeFn func(this storyTransition, last storyTransition, d time.Duration) (CycleTimeFn, time.Duration)

func dontSumCycleTime(transition storyTransition, last storyTransition, d time.Duration) (CycleTimeFn, time.Duration) {
	// log.Printf("outside %s -> %s", transition.FromState, transition.ToState)
	if transition.ToState != "unstarted" && transition.ToState != "unscheduled" {
		return sumCycleTime, d
	}
	return dontSumCycleTime, d
}

func sumCycleTime(transition storyTransition, last storyTransition, d time.Duration) (CycleTimeFn, time.Duration) {
	// log.Printf("inside %s -> %s", transition.FromState, transition.ToState)
	d = d + transition.OccurredAt.Sub(last.OccurredAt)

	if transition.ToState == "unstarted" || transition.ToState == "unscheduled" || transition.ToState == "accepted" {
		return dontSumCycleTime, d
	}

	return sumCycleTime, d
}
