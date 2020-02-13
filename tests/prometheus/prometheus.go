package prometheus

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/mesosphere/kudo-cassandra-operator/tests/curl"
)

type PrometheusSearch struct {
	Status string               `json:"parameters"`
	Data   PrometheusSearchData `json:"data"`
}

type PrometheusSearchData struct {
	Type   string                   `json:"resultType"`
	Result []PrometheusSearchResult `json:"result"`
}

type PrometheusSearchResult struct {
	Metric map[string]interface{} `json:"metric"`
	Value  []interface{}          `json:"value"`
}

func QueryForStats(client client.Client, namespace string, baseURL, query string) (*PrometheusSearch, error) {

	url := fmt.Sprintf("%s/api/v1/query?query=%s", baseURL, query)

	stdout, stderr, err := curl.RunCommand(client, namespace, "-s", url)
	if err != nil {
		return nil, fmt.Errorf("failed to run curl command, stdErr is %s:, %v", stderr, err)
	}

	reader := strings.NewReader(stdout)
	promResult := &PrometheusSearch{}
	jsonReader := json.NewDecoder(reader)
	err = jsonReader.Decode(promResult)
	if err != nil {
		return nil, err
	}

	return promResult, nil
}
