package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/UnstoppableMango/thecluster.lan/api/internal/model"
)

type Sample struct {
	Labels map[string]string
	Value  float64
}

type Series struct {
	Labels map[string]string
	Points []model.Point
}

// Client is a minimal Prometheus HTTP API client covering instant and range queries.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTP:    &http.Client{Timeout: 10 * time.Second},
	}
}

type apiResponse struct {
	Status    string `json:"status"`
	ErrorType string `json:"errorType"`
	Error     string `json:"error"`
	Data      struct {
		ResultType string          `json:"resultType"`
		Result     json.RawMessage `json:"result"`
	} `json:"data"`
}

type rawSample struct {
	Metric map[string]string `json:"metric"`
	Value  [2]any            `json:"value"`
}

type rawSeries struct {
	Metric map[string]string `json:"metric"`
	Values [][2]any          `json:"values"`
}

// Query runs an instant query. NaN and infinite samples are dropped.
func (c *Client) Query(ctx context.Context, query string, at time.Time) ([]Sample, error) {
	params := url.Values{
		"query": {query},
		"time":  {formatTime(at)},
	}

	var raw []rawSample
	if err := c.get(ctx, "/api/v1/query", params, "vector", &raw); err != nil {
		return nil, err
	}

	samples := make([]Sample, 0, len(raw))
	for _, r := range raw {
		p, ok := parsePoint(r.Value)
		if !ok {
			continue
		}
		samples = append(samples, Sample{Labels: r.Metric, Value: p[1]})
	}

	return samples, nil
}

// QueryRange runs a range query. NaN and infinite points are dropped.
func (c *Client) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]Series, error) {
	params := url.Values{
		"query": {query},
		"start": {formatTime(start)},
		"end":   {formatTime(end)},
		"step":  {strconv.FormatFloat(step.Seconds(), 'f', -1, 64)},
	}

	var raw []rawSeries
	if err := c.get(ctx, "/api/v1/query_range", params, "matrix", &raw); err != nil {
		return nil, err
	}

	series := make([]Series, 0, len(raw))
	for _, r := range raw {
		points := make([]model.Point, 0, len(r.Values))
		for _, v := range r.Values {
			if p, ok := parsePoint(v); ok {
				points = append(points, p)
			}
		}
		series = append(series, Series{Labels: r.Metric, Points: points})
	}

	return series, nil
}

func (c *Client) get(ctx context.Context, path string, params url.Values, resultType string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path+"?"+params.Encode(), nil)
	if err != nil {
		return err
	}

	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var body apiResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return fmt.Errorf("prometheus %s: status %d: decode: %w", path, res.StatusCode, err)
	}
	if body.Status != "success" {
		return fmt.Errorf("prometheus %s: %s: %s", path, body.ErrorType, body.Error)
	}
	if body.Data.ResultType != resultType {
		return fmt.Errorf("prometheus %s: expected %s result, got %s", path, resultType, body.Data.ResultType)
	}

	return json.Unmarshal(body.Data.Result, out)
}

func parsePoint(v [2]any) (model.Point, bool) {
	ts, ok := v[0].(float64)
	if !ok {
		return nil, false
	}
	s, ok := v[1].(string)
	if !ok {
		return nil, false
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return nil, false
	}

	return model.Point{ts, f}, true
}

func formatTime(t time.Time) string {
	return strconv.FormatFloat(float64(t.UnixMilli())/1000, 'f', -1, 64)
}
