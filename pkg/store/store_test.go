package store

import (
	"errors"
	"os"
	"regexp"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/thanos-io/thanos/pkg/store/storepb"

	opentsdb "github.com/G-Research/opentsdb-goclient/client"
)

func TestComposeOpenTSDBQuery(t *testing.T) {
	testCases := []struct {
		req                            storepb.SeriesRequest
		tsdbQ                          *opentsdb.QueryParam
		knownMetrics                   []string
		err                            error
		allowedMetrics, blockedMetrics *regexp.Regexp
	}{
		{
			knownMetrics: []string{"test.metric"},
			req: storepb.SeriesRequest{
				MinTime: 0,
				MaxTime: 100,
				Matchers: []storepb.LabelMatcher{
					{
						Type:  storepb.LabelMatcher_EQ,
						Name:  "key",
						Value: "value",
					},
					{
						Type:  storepb.LabelMatcher_EQ,
						Name:  "__name__",
						Value: "test.metric",
					},
				},
				MaxResolutionWindow:     5,
				Aggregates:              []storepb.Aggr{storepb.Aggr_SUM},
				PartialResponseDisabled: false,
			},
			tsdbQ: &opentsdb.QueryParam{
				Start: 0,
				End:   100,
				Queries: []opentsdb.SubQuery{
					{
						Aggregator: "none",
						Metric:     "test.metric",
						Fiters: []opentsdb.Filter{
							{
								Type:      "wildcard",
								Tagk:      "key",
								FilterExp: "value",
								GroupBy:   true,
							},
						},
					},
				},
			},
		},
		{
			req: storepb.SeriesRequest{
				MinTime: 0,
				MaxTime: 100,
				Matchers: []storepb.LabelMatcher{
					{
						Type:  storepb.LabelMatcher_EQ,
						Name:  "__name__",
						Value: "test.metric",
					},
				},
				MaxResolutionWindow:     5,
				Aggregates:              []storepb.Aggr{storepb.Aggr_MAX},
				PartialResponseDisabled: false,
			},
			tsdbQ: &opentsdb.QueryParam{
				Start: 0,
				End:   100,
				Queries: []opentsdb.SubQuery{
					{
						Aggregator: "none",
						Metric:     "test.metric",
						Fiters:     []opentsdb.Filter{},
					},
				},
			},
		},
		{
			knownMetrics: []string{"test.metric2"},
			req: storepb.SeriesRequest{
				MinTime: 0,
				MaxTime: 100,
				Matchers: []storepb.LabelMatcher{
					{
						Type:  storepb.LabelMatcher_EQ,
						Name:  "host",
						Value: "*",
					},
					{
						Type:  storepb.LabelMatcher_EQ,
						Name:  "__name__",
						Value: "test.metric2",
					},
				},
				MaxResolutionWindow:     5,
				Aggregates:              []storepb.Aggr{storepb.Aggr_MIN},
				PartialResponseDisabled: false,
			},
			tsdbQ: &opentsdb.QueryParam{
				Start: 0,
				End:   100,
				Queries: []opentsdb.SubQuery{
					{
						Aggregator: "none",
						Metric:     "test.metric2",
						Fiters: []opentsdb.Filter{
							{
								Type:      "wildcard",
								Tagk:      "host",
								FilterExp: "*",
								GroupBy:   true,
							},
						},
					},
				},
			},
		},
		{
			req: storepb.SeriesRequest{
				MinTime: 0,
				MaxTime: 100,
				Matchers: []storepb.LabelMatcher{
					{
						Type:  storepb.LabelMatcher_NRE,
						Name:  "host",
						Value: ".*",
					},
					{
						Type:  storepb.LabelMatcher_NRE,
						Name:  "__name__",
						Value: "test.metric2",
					},
				},
				MaxResolutionWindow:     5,
				Aggregates:              []storepb.Aggr{storepb.Aggr_MIN},
				PartialResponseDisabled: false,
			},
			err: errors.New("LabelMatcher_NRE is not supported"),
		},
		{
			req: storepb.SeriesRequest{
				MinTime: 0,
				MaxTime: 100,
				Matchers: []storepb.LabelMatcher{
					{
						Type:  storepb.LabelMatcher_RE,
						Name:  "host",
						Value: ".*",
					},
					{
						Type:  storepb.LabelMatcher_EQ,
						Name:  "__name__",
						Value: "test.metric2",
					},
					{
						Type:  storepb.LabelMatcher_NEQ,
						Name:  "key",
						Value: "v",
					},
				},
				MaxResolutionWindow:     5,
				Aggregates:              []storepb.Aggr{storepb.Aggr_MIN},
				PartialResponseDisabled: false,
			},
			tsdbQ: &opentsdb.QueryParam{
				Start: 0,
				End:   100,
				Queries: []opentsdb.SubQuery{
					{
						Aggregator: "none",
						Metric:     "test.metric2",
						Fiters: []opentsdb.Filter{
							{
								Type:      "regexp",
								Tagk:      "host",
								FilterExp: ".*",
								GroupBy:   true,
							},
							{
								Type:      "not_literal_or",
								Tagk:      "key",
								FilterExp: "v",
								GroupBy:   true,
							},
						},
					},
				},
			},
		},
		{
			req: storepb.SeriesRequest{
				MinTime: 0,
				MaxTime: 100,
				Matchers: []storepb.LabelMatcher{
					{
						Type:  storepb.LabelMatcher_RE,
						Name:  "host",
						Value: ".*",
					},
					{
						Type:  storepb.LabelMatcher_EQ,
						Name:  "__name__",
						Value: "test:metric2:sub:subsub",
					},
					{
						Type:  storepb.LabelMatcher_NEQ,
						Name:  "key",
						Value: "v",
					},
				},
				MaxResolutionWindow:     5,
				Aggregates:              []storepb.Aggr{storepb.Aggr_MIN},
				PartialResponseDisabled: false,
			},
			tsdbQ: &opentsdb.QueryParam{
				Start: 0,
				End:   100,
				Queries: []opentsdb.SubQuery{
					{
						Aggregator: "none",
						Metric:     "test.metric2.sub.subsub",
						Fiters: []opentsdb.Filter{
							{
								Type:      "regexp",
								Tagk:      "host",
								FilterExp: ".*",
								GroupBy:   true,
							},
							{
								Type:      "not_literal_or",
								Tagk:      "key",
								FilterExp: "v",
								GroupBy:   true,
							},
						},
					},
				},
			},
		},
		{
			knownMetrics: []string{"a", "b", "c"},
			req: storepb.SeriesRequest{
				MinTime: 0,
				MaxTime: 100,
				Matchers: []storepb.LabelMatcher{
					{
						Type:  storepb.LabelMatcher_RE,
						Name:  "__name__",
						Value: ".*",
					},
					{
						Type:  storepb.LabelMatcher_NEQ,
						Name:  "key",
						Value: "v",
					},
				},
				MaxResolutionWindow:     5,
				Aggregates:              []storepb.Aggr{storepb.Aggr_MIN},
				PartialResponseDisabled: false,
			},
			tsdbQ: &opentsdb.QueryParam{
				Start: 0,
				End:   100,
				Queries: []opentsdb.SubQuery{
					{
						Aggregator: "none",
						Metric:     "a",
						Fiters: []opentsdb.Filter{
							{
								Type:      "not_literal_or",
								Tagk:      "key",
								FilterExp: "v",
								GroupBy:   true,
							},
						},
					},
					{
						Aggregator: "none",
						Metric:     "b",
						Fiters: []opentsdb.Filter{
							{
								Type:      "not_literal_or",
								Tagk:      "key",
								FilterExp: "v",
								GroupBy:   true,
							},
						},
					},
					{
						Aggregator: "none",
						Metric:     "c",
						Fiters: []opentsdb.Filter{
							{
								Type:      "not_literal_or",
								Tagk:      "key",
								FilterExp: "v",
								GroupBy:   true,
							},
						},
					},
				},
			},
		},
		{
			knownMetrics:   []string{"test.metric", "other.metric"},
			allowedMetrics: regexp.MustCompile(`test\..*`),
			req: storepb.SeriesRequest{
				MinTime: 0,
				MaxTime: 100,
				Matchers: []storepb.LabelMatcher{
					{
						Type:  storepb.LabelMatcher_RE,
						Name:  "__name__",
						Value: `(other|test)\.metric`,
					},
				},
				MaxResolutionWindow:     5,
				Aggregates:              []storepb.Aggr{storepb.Aggr_SUM},
				PartialResponseDisabled: false,
			},
			tsdbQ: &opentsdb.QueryParam{
				Start: 0,
				End:   100,
				Queries: []opentsdb.SubQuery{
					{
						Aggregator: "none",
						Metric:     "test.metric",
					},
				},
			},
		},
		{
			knownMetrics:   []string{"test.metric"},
			allowedMetrics: regexp.MustCompile(`^\w+\.`),
			req: storepb.SeriesRequest{
				MinTime: 0,
				MaxTime: 100,
				Matchers: []storepb.LabelMatcher{
					{
						Type:  storepb.LabelMatcher_EQ,
						Name:  "__name__",
						Value: "up",
					},
				},
				MaxResolutionWindow:     5,
				Aggregates:              []storepb.Aggr{storepb.Aggr_SUM},
				PartialResponseDisabled: false,
			},
			// All metric names filtered out
			tsdbQ: &opentsdb.QueryParam{},
		},
		{
			knownMetrics:   []string{"bad.metric"},
			blockedMetrics: regexp.MustCompile(`bad`),
			req: storepb.SeriesRequest{
				MinTime: 0,
				MaxTime: 100,
				Matchers: []storepb.LabelMatcher{
					{
						Type:  storepb.LabelMatcher_EQ,
						Name:  "__name__",
						Value: "bad.metric",
					},
				},
				MaxResolutionWindow:     5,
				Aggregates:              []storepb.Aggr{storepb.Aggr_SUM},
				PartialResponseDisabled: false,
			},
			err: errors.New(`Metric "bad.metric" is blocked on Geras`),
		},
		{
			knownMetrics:   []string{"bad.metric"},
			blockedMetrics: regexp.MustCompile(`bad\.`),
			req: storepb.SeriesRequest{
				MinTime: 0,
				MaxTime: 100,
				Matchers: []storepb.LabelMatcher{
					{
						Type:  storepb.LabelMatcher_EQ,
						Name:  "__name__",
						Value: "bad:metric",
					},
				},
				MaxResolutionWindow:     5,
				Aggregates:              []storepb.Aggr{storepb.Aggr_SUM},
				PartialResponseDisabled: false,
			},
			err: errors.New(`Metric "bad.metric" is blocked on Geras`),
		},
	}

	for _, test := range testCases {
		allowedMetrics := regexp.MustCompile(".*")
		if test.allowedMetrics != nil {
			allowedMetrics = test.allowedMetrics
		}
		store := OpenTSDBStore{
			metricNames:        test.knownMetrics,
			logger:             log.NewJSONLogger(os.Stdout),
			allowedMetricNames: allowedMetrics,
			blockedMetricNames: test.blockedMetrics,
		}

		p, _, err := store.composeOpenTSDBQuery(&test.req)
		if test.err != nil {
			if test.err.Error() != err.Error() {
				t.Error("not expected error")
			}
			continue
		}
		if err != nil {
			t.Error(err)
		}
		if len(p.Queries) != len(test.tsdbQ.Queries) {
			t.Errorf("expected %d queries, got %d", len(test.tsdbQ.Queries), len(p.Queries))
		}
		if len(test.tsdbQ.Queries) == 0 {
			continue
		}
		// test the requested ranges
		if test.tsdbQ.Start.(int) != int(p.Start.(int64)) ||
			test.tsdbQ.End.(int) != int(p.End.(int64)) {
			t.Errorf("requested range is not equal to sent range (%d - %d) != (%d - %d)",
				p.Start, p.End, test.tsdbQ.Start, test.tsdbQ.End)
		}
		if len(p.Queries) != len(test.tsdbQ.Queries) {
			t.Errorf("number of subqueries does not match")
		}
		for _, referenceQ := range test.tsdbQ.Queries {
			match := false
			for _, subQ := range p.Queries {
				// test aggregator
				if subQ.Aggregator != referenceQ.Aggregator {
					t.Log("\taggregator does not match")
					match = false
					continue
				}
				// test filters
				filters := map[string]opentsdb.Filter{}
				for _, f := range referenceQ.Fiters {
					filters[f.Tagk] = f
				}
				if len(filters) != len(subQ.Fiters) {
					t.Log("\tfilter length does not match")
					match = false
					continue
				}
				for _, f := range referenceQ.Fiters {
					if expFilter, ok := filters[f.Tagk]; !ok || expFilter != f {
						t.Log("\tfilter does not match")
						match = false
						break
					}
				}
				// check metric name
				if referenceQ.Metric != subQ.Metric {
					continue
				}
				match = true
				break
			}
			if !match {
				t.Errorf("there is no matching subquery for %v", referenceQ)
			}
		}
	}
}

func TestConvertOpenTSDBResultsToSeriesResponse(t *testing.T) {
	testCases := []struct {
		input          opentsdb.QueryRespItem
		expectedOutput *storepb.SeriesResponse
	}{
		{
			input: opentsdb.QueryRespItem{
				Metric: "metric",
				Tags:   map[string]string{},
				Dps:    map[string]interface{}{},
			},
			expectedOutput: storepb.NewSeriesResponse(&storepb.Series{
				Labels: []storepb.Label{{Name: "__name__", Value: "metric"}},
				Chunks: []storepb.AggrChunk{{MinTime: 0, MaxTime: 0}},
			}),
		},
		{
			input: opentsdb.QueryRespItem{
				Metric: "metric",
				Tags:   map[string]string{"a": "b"},
				Dps: map[string]interface{}{
					"1": 1.0,
					"2": 1.5,
					"3": 2.0,
				},
			},
			expectedOutput: storepb.NewSeriesResponse(&storepb.Series{
				Labels: []storepb.Label{{Name: "__name__", Value: "metric"}, {Name: "a", Value: "b"}},
				Chunks: []storepb.AggrChunk{{MinTime: 1, MaxTime: 3}},
			}),
		},
		{
			input: opentsdb.QueryRespItem{
				Metric: "metric2",
				Tags:   map[string]string{"a": "b", "host": "test"},
				Dps: map[string]interface{}{
					"10": 1.0,
					"12": 1.5,
					"13": 2.0,
				},
			},
			expectedOutput: storepb.NewSeriesResponse(&storepb.Series{
				Labels: []storepb.Label{
					{Name: "__name__", Value: "metric2"},
					{Name: "host", Value: "test"},
					{Name: "a", Value: "b"}},
				Chunks: []storepb.AggrChunk{{MinTime: 10, MaxTime: 13}},
			}),
		},
	}
	for _, test := range testCases {
		converted, err := convertOpenTSDBResultsToSeriesResponse(test.input)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		expectedTags := map[string]string{}
		for _, v := range test.expectedOutput.GetSeries().Labels {
			expectedTags[v.Name] = v.Value
		}
		if len(converted.GetSeries().Labels) == len(test.expectedOutput.GetSeries().Labels) {
			for _, tag := range converted.GetSeries().Labels {
				if val, ok := expectedTags[tag.Name]; !ok || val != tag.Value {
					t.Errorf("unexpected tag: %s", tag.Name)
				}
			}
		} else {
			t.Errorf("number of tags does not match")
		}
		if len(test.expectedOutput.GetSeries().Chunks) != len(converted.GetSeries().Chunks) {
			t.Error("number of chunks does not match")
		}
		for ci, chunk := range test.expectedOutput.GetSeries().Chunks {
			if chunk.MinTime != converted.GetSeries().Chunks[ci].MinTime {
				t.Errorf("chunk %d min time is not the expected: %d", ci, chunk.MinTime)
			}
			if chunk.MaxTime != converted.GetSeries().Chunks[ci].MaxTime {
				t.Errorf("chunk %d max time is not the expected: %d != %d ", ci, chunk.MaxTime, converted.GetSeries().Chunks[ci].MaxTime)
			}
		}
	}
}
