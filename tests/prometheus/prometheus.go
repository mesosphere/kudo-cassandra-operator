package prometheus

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mesosphere/kudo-cassandra-operator/tests/curl"
)

type Search struct {
	Status string     `json:"parameters"`
	Data   SearchData `json:"data"`
}

type SearchData struct {
	Type   string         `json:"resultType"`
	Result []SearchResult `json:"result"`
}

type SearchResult struct {
	Metric map[string]interface{} `json:"metric"`
	Value  []interface{}          `json:"value"`
}

func QueryForStats(curl curl.Runner, baseURL, query string) (*Search, error) {

	url := fmt.Sprintf("%s/api/v1/query?query=%s", baseURL, query)

	stdout, stderr, err := curl.Run("-s", url)
	if err != nil {
		return nil, fmt.Errorf("failed to run curl command, stdErr is %s:, %v", stderr, err)
	}

	reader := strings.NewReader(stdout)
	promResult := &Search{}
	jsonReader := json.NewDecoder(reader)
	err = jsonReader.Decode(promResult)
	if err != nil {
		return nil, err
	}

	return promResult, nil
}
