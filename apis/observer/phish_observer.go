package observer

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/ari-kl/phish-stream/util"
)

var client = http.DefaultClient

const PHISH_OBSERVER_URL = "https://phish.observer/api/"

type SubmitBody struct {
	Url  string   `json:"url"`
	Tags []string `json:"tags"`
}

type SubmitResponse struct {
	ID string `json:"id"`
}

func SubmitUrl(url string, tags []string) (SubmitResponse, error) {
	body := SubmitBody{
		Url:  url,
		Tags: tags,
	}

	jsonBody, err := json.Marshal(body)

	if err != nil {
		return SubmitResponse{}, err
	}

	req, err := http.NewRequest("POST", PHISH_OBSERVER_URL+"submit", bytes.NewBuffer(jsonBody))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", util.USER_AGENT)
	req.Header.Set("Authorization", "Bearer "+os.Getenv("PHISH_OBSERVER_API_KEY"))

	resp, err := client.Do(req)

	if err != nil {
		return SubmitResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SubmitResponse{}, errors.New("Failed to submit URL: " + resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return SubmitResponse{}, err
	}

	var submitResponse SubmitResponse
	json.Unmarshal(bodyBytes, &submitResponse)

	return submitResponse, nil
}
