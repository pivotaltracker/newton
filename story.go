package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func GetStoriesFromPastIterations(token string, projectID int, pastIters int) ([]Story, error) {
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
