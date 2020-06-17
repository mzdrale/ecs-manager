package common

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"time"
)

// ElementInSlice - returns true if slice contains element
func ElementInSlice(element string, s []string) bool {
	for _, e := range s {
		if e == element {
			return true
		}
	}
	return false
}

// FileExists - returns true if file exists, otherwise returns false
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// FormatDuration - format duration into human readable time format
func FormatDuration(d time.Duration) string {
	durationString := ""

	// Hours
	h := d / time.Hour
	d -= h * time.Hour

	if h > 0 {
		durationString = fmt.Sprintf("%d hour", h)
	}

	// Minutes
	m := d / time.Minute
	d -= m * time.Minute

	// Seconds
	s := d / time.Second

	durationString = fmt.Sprintf("%ds", s)

	if m > 0 {
		durationString = fmt.Sprintf("%dm %s", m, durationString)
	}

	if h > 0 {
		durationString = fmt.Sprintf("%dh %s", h, durationString)
	}

	// return fmt.Sprintf("%d hour(s) %d minute(s) %d second(s)", h, m, s)
	// fmt.Printf("Duration debug: %s\n", durationString)
	return durationString
}

// ReadExcludedInstancesList - reads excluded instances list from file
func ReadExcludedInstancesList(path string) ([]string, error) {
	var lines []string
	if FileExists(path) {
		file, err := os.Open(path)
		if err != nil {
			return lines, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			re := regexp.MustCompile(`^(\S+)\s+\(.*\)$`)
			m := re.FindStringSubmatch(line)
			if len(m) > 0 {
				lines = append(lines, m[1])
			}
		}
	}
	return lines, nil
}
