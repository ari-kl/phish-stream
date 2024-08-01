package main

import (
	"github.com/CaliDog/certstream-go"
	"github.com/ari-kl/phish-stream/util"
)

func StartStreaming() {
	stream, errStream := certstream.CertStreamEventStream(false)
	for {
		select {
		case jq := <-stream:
			messageType, _ := jq.String("message_type")

			if messageType == "certificate_update" {
				// Extract list of domains from the certificate
				domains, err := jq.ArrayOfStrings("data", "leaf_cert", "all_domains")

				if err != nil {
					util.Logger.Error(err.Error())
					continue
				}

				for _, domain := range domains {
					go RunFilters(domain)
				}
			}

		case err := <-errStream:
			util.Logger.Error(err.Error())
		}
	}
}

// TODO: configurable filters directory
var filters []Filter = LoadFilters("./filters")

func RunFilters(domain string) {
	for _, filter := range filters {
		result := filter.FilterDomain(domain)
		if result.matched {
			// Just log the match for now
			// TODO: further processing & review
			util.Logger.Info("Match", "domain", domain, "filter", filter.Name, "matchType", result.matchType, "matchedBy", result.matchedBy, "similarityScore", result.similarityScore)
		}
	}
}
