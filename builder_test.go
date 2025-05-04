package goty_test

import (
	"testing"
	"time"

	"golift.io/goty"
)

type TestWrapper struct {
	Profile TestLevel1
	Level1  TestLevel1
	TestEndpoint
	EP   *TestEndpoint
	Auth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
}

var Weekdays = []time.Weekday{
	time.Sunday,
	time.Monday,
	time.Tuesday,
	time.Wednesday,
	time.Thursday,
	time.Friday,
	time.Saturday,
}

type TestLevel1 struct {
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

type TestEndpoint struct {
	URL    string `json:"url"`
	APIKey string `json:"apiKey"`
}

// This test is incomplete.
func TestBuilder(t *testing.T) {
	t.Parallel()

	goty := goty.NewGoty(nil)
	goty.Parse(TestWrapper{})
	goty.Print()
}
