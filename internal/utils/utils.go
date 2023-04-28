package utils

import (
	"os"
	"sort"
	"time"

	"github.com/hibare/GoPG2S3Dump/internal/constants"
)

func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

func SortDateTimes(datetimes []string) []string {
	// Convert the strings to time.Time objects
	var times []time.Time
	for _, dt := range datetimes {
		t, _ := time.Parse(constants.DefaultDateTimeLayout, dt)
		times = append(times, t)
	}

	// Define a sorting function
	sortFn := func(i, j int) bool {
		return times[i].After(times[j])
	}

	// Sort the slice of time.Time objects
	sort.Slice(times, sortFn)

	// Convert the sorted time.Time objects back to strings
	var sorted []string
	for _, t := range times {
		sorted = append(sorted, t.Format(constants.DefaultDateTimeLayout))
	}

	return sorted
}
