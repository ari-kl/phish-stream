package util

import "github.com/weppos/publicsuffix-go/publicsuffix"

// Strips the effective top-level domain from a given domain name
// Examples: "www.example.com" -> "www.example", "example.co.uk" -> "example"
func StripETLD(domain string) string {
	data, err := publicsuffix.Parse(domain)

	if err != nil {
		return domain
	}

	// Subdomain + Second level domain
	return data.TRD + "." + data.SLD
}
