package filter

import (
	"regexp"
	"strings"

	"github.com/antzucaro/matchr"
	"github.com/ari-kl/phish-stream/shared"
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

// Empty non-matching result, here for reusability
var noMatch = shared.FilterResult{Matched: false, MatchType: shared.FilterMatchTypeNone, MatchedBy: "", SimilarityScore: 0}

func (f Filter) FilterDomain(domain string) shared.FilterResult {
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
			return shared.FilterResult{Name: f.Name, Matched: true, MatchType: shared.FilterMatchTypeKeyword, MatchedBy: keyword}
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
				distance := matchr.JaroWinkler(part, term, false)

				if distance >= similarity.Threshold {
					return shared.FilterResult{Name: f.Name, Matched: true, MatchType: shared.FilterMatchTypeSimilarity, MatchedBy: term, SimilarityScore: distance}
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
			return shared.FilterResult{Name: f.Name, Matched: true, MatchType: shared.FilterMatchTypeRegex, MatchedBy: pattern}
		}
	}

	return noMatch
}
