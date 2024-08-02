package filter

import (
	"os"
	"path/filepath"

	"github.com/ari-kl/phish-stream/config"
	"github.com/ari-kl/phish-stream/review"
	"github.com/ari-kl/phish-stream/util"
	"gopkg.in/yaml.v3"
)

var filters []Filter

func RunFilters(domain string) {
	stripped_domain := util.StripETLD(domain)

	for _, filter := range filters {
		result := filter.FilterDomain(stripped_domain)
		if result.Matched {
			// Just log the match for now
			// TODO: further processing & review
			util.Logger.Info("Match", "domain", domain, "filter", filter.Name, "matchType", result.MatchType, "matchedBy", result.MatchedBy, "similarityScore", result.SimilarityScore)
			review.SendMessage(domain, result)
		}
	}
}

func InitFilters() {
	filtersPath := config.FiltersDir
	filters = LoadFilters(filtersPath)
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
