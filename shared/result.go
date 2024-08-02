package shared

// Domain filtering result data
type FilterMatchType int

const (
	FilterMatchTypeKeyword FilterMatchType = iota
	FilterMatchTypeSimilarity
	FilterMatchTypeRegex
	FilterMatchTypeNone
)

type FilterResult struct {
	Name      string
	Matched   bool
	MatchType FilterMatchType
	// The keyword, regex, or similarity term that matched the domain
	MatchedBy string
	// SimilarityScore is only used when matchType is FilterMatchTypeSimilarity
	SimilarityScore float64
}
