package filter

import (
	"regexp"
	"strings"

	"github.com/antzucaro/matchr"
	"github.com/ari-kl/phish-stream/util"
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

	// Break up the domain into parts before checking for similarity
	// This is to prevent longer domains from bypassing short terms
	parts := strings.FieldsFunc(domain, func(r rune) bool {
		return r == '.' || r == '-'
	})

	for _, similarity := range f.Similarity {
		for _, term := range similarity.Terms {
			for _, part := range parts {
				distance := matchr.JaroWinkler(part, term, true)

				if distance >= similarity.Threshold {
					return FilterResult{name: f.Name, matched: true, matchType: FilterMatchTypeSimilarity, matchedBy: term, similarityScore: distance}
				}
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
