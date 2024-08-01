package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/antzucaro/matchr"
	"github.com/ari-kl/phish-stream/util"
	"gopkg.in/yaml.v3"
)

// YAML configuration file
type Filter struct {
	Name       string             `yaml:"name"`
	Enabled    bool               `yaml:"enabled"`
	Keywords   []string           `yaml:"keywords"`
	Similarity []FilterSimilarity `yaml:"similarity"`
	Regex      []string           `yaml:"regex"`
	Exclusions []string           `yaml:"exclusions"`
}

type FilterSimilarity struct {
	Threshold float64  `yaml:"threshold"`
	Terms     []string `yaml:"terms"`
}

// Domain filtering result data
type FilterMatchType int

const (
	FilterMatchTypeKeyword FilterMatchType = iota
	FilterMatchTypeSimilarity
	FilterMatchTypeRegex
	FilterMatchTypeNone
)

type FilterResult struct {
	name      string
	matched   bool
	matchType FilterMatchType
	// The keyword, regex, or similarity term that matched the domain
	matchedBy string
	// similarityScore is only used when matchType is FilterMatchTypeSimilarity
	similarityScore float64
}

// Empty non-matching result, here for reusability
var noMatch = FilterResult{matched: false, matchType: FilterMatchTypeNone, matchedBy: "", similarityScore: 0}

func (f Filter) FilterDomain(domain string) FilterResult {
	if !f.Enabled {
		return noMatch
	}

	for _, exclusion := range f.Exclusions {
		if strings.Contains(domain, exclusion) {
			return noMatch
		}
	}

	for _, keyword := range f.Keywords {
		if strings.Contains(domain, keyword) {
			return FilterResult{name: f.Name, matched: true, matchType: FilterMatchTypeKeyword, matchedBy: keyword}
		}
	}

	for _, similarity := range f.Similarity {
		for _, term := range similarity.Terms {
			distance := matchr.JaroWinkler(domain, term, true)

			if distance >= similarity.Threshold {
				return FilterResult{name: f.Name, matched: true, matchType: FilterMatchTypeSimilarity, matchedBy: term, similarityScore: distance}
			}
		}
	}

	for _, pattern := range f.Regex {
		match, err := regexp.MatchString(pattern, domain)

		if err != nil {
			util.Logger.Error(err.Error())
			continue
		}

		if match {
			return FilterResult{name: f.Name, matched: true, matchType: FilterMatchTypeRegex, matchedBy: pattern}
		}
	}

	return noMatch
}

func LoadFilters(filtersPath string) []Filter {
	// Iterate over all files in the filters directory
	files, err := os.ReadDir(filtersPath)

	if err != nil {
		util.Logger.Error(err.Error())
		return []Filter{}
	}

	var filters []Filter

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		extension := filepath.Ext(file.Name())

		if extension != ".yaml" && extension != ".yml" {
			continue
		}

		filter, err := readFilterFile(filepath.Join(filtersPath, file.Name()))

		if err != nil {
			util.Logger.Error(err.Error())
			continue
		}

		filters = append(filters, filter)
	}

	return filterEnabledFilters(filters)
}

func readFilterFile(filePath string) (Filter, error) {
	data, err := os.ReadFile(filePath)

	if err != nil {
		return Filter{}, err
	}

	filter := Filter{}
	err = yaml.Unmarshal(data, &filter)

	if err != nil {
		return Filter{}, err
	}

	return filter, nil
}

func filterEnabledFilters(filters []Filter) []Filter {
	var enabledFilters []Filter

	for _, filter := range filters {
		if filter.Enabled {
			enabledFilters = append(enabledFilters, filter)
		}
	}

	return enabledFilters
}
