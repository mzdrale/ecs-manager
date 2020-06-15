package common

import (
	"bufio"
	"os"
	"regexp"
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
