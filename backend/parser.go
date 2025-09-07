package main

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type ParsedMovie struct {
	Title string
	Year  int
}

func parseMovieTitle(filename string) ParsedMovie {
	// Remove file extension
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))

	// Common patterns for movie titles
	patterns := []struct {
		regex *regexp.Regexp
		fn    func([]string) ParsedMovie
	}{
		// Pattern: Movie.Name.2023.quality.info.mkv
		{
			regex: regexp.MustCompile(`^(.+?)\.(\d{4})`),
			fn: func(matches []string) ParsedMovie {
				year, _ := strconv.Atoi(matches[2])
				return ParsedMovie{
					Title: strings.ReplaceAll(matches[1], ".", " "),
					Year:  year,
				}
			},
		},
		// Pattern: Movie Name [2023 quality info]
		{
			regex: regexp.MustCompile(`^(.+?)\s*\[(\d{4})`),
			fn: func(matches []string) ParsedMovie {
				year, _ := strconv.Atoi(matches[2])
				return ParsedMovie{
					Title: strings.TrimSpace(matches[1]),
					Year:  year,
				}
			},
		},
		// Pattern: Movie Name (2023)
		{
			regex: regexp.MustCompile(`^(.+?)\s*\((\d{4})\)`),
			fn: func(matches []string) ParsedMovie {
				year, _ := strconv.Atoi(matches[2])
				return ParsedMovie{
					Title: strings.TrimSpace(matches[1]),
					Year:  year,
				}
			},
		},
	}

	// Try each pattern
	for _, pattern := range patterns {
		matches := pattern.regex.FindStringSubmatch(filename)
		if len(matches) > 0 {
			return pattern.fn(matches)
		}
	}

	// If no pattern matches, return the filename as is
	return ParsedMovie{
		Title: strings.ReplaceAll(filename, ".", " "),
		Year:  0,
	}
}
