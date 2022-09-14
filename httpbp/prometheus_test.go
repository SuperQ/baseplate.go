package httpbp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/reddit/baseplate.go"
	"github.com/reddit/baseplate.go/ecinterface"
	"github.com/reddit/baseplate.go/httpbp/metrics"
	"github.com/reddit/baseplate.go/internal/prometheusbp/spectest"
	"github.com/reddit/baseplate.go/prometheusbp/promtest"
)

type exampleRequest struct {
	Input string `json:"input"`
}

type exampleResponse struct {
	Message string `json:"message"`
}

func TestPrometheusClientServerMetrics(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		success  string
		method   string
		endpoint string
		route    string
		reqSize  int
		respSize int
	}{
		{
			name:     "success get",
			code:     "200",
			success:  "true",
			method:   http.MethodGet,
			endpoint: "test",
			route:    "/test",
		},
		{
			name:     "err post",
			code:     "401",
			success:  "false",
			method:   http.MethodPost,
			endpoint: "error2",
			route:    "/error2",
			reqSize:  16,
			respSize: 29,
		},
		{
			name:     "internal err get",
			code:     "500",
			success:  "false",
			method:   http.MethodGet,
			endpoint: "error",
			route:    "/error",
		},
	}

	const serverSlug = "testServer"

	args := ServerArgs{
		Baseplate: baseplate.NewTestBaseplate(baseplate.NewTestBaseplateArgs{
			Config:          baseplate.Config{Addr: ":8080"},
			EdgeContextImpl: ecinterface.Mock(),
		}),
		Endpoints: map[Pattern]Endpoint{
			"/test": {
				Name:    "test",
				Methods: []string{http.MethodGet},
				Handle:  func(ctx context.Context, w http.ResponseWriter, r *http.Request) error { return nil },
			},
			"/error2": {
				Name:    "error2",
				Methods: []string{http.MethodPost},
				Handle: func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					var req exampleRequest
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						return fmt.Errorf("decoding %T: %w", req, err)
					}
					body := exampleResponse{
						Message: fmt.Sprintf("Input: %q", req.Input),
					}
					return WriteJSON(w, Response{Body: body, Code: Unauthorized().code})
				},
			},
			"/error": {
				Name:    "error",
				Methods: []string{http.MethodGet},
				Handle: func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					return errors.New("test")
				},
			},
		},
	}

	server, ts, err := NewTestBaseplateServer(args)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	middleware := PrometheusClientMetrics(serverSlug)
	client := &http.Client{
		Transport: middleware(http.DefaultTransport),
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			metrics.ServerLatency.Reset()
			metrics.ServerTotalRequests.Reset()
			metrics.ServerActiveRequests.Reset()
			metrics.ClientLatencyDistribution.Reset()
			metrics.ClientTotalRequests.Reset()
			metrics.ClientActiveRequests.Reset()

			serverSizeLabels := prometheus.Labels{
				metrics.MethodLabel:   tt.method,
				metrics.SuccessLabel:  tt.success,
				metrics.EndpointLabel: tt.endpoint,
			}

			serverTotalRequestLabels := prometheus.Labels{
				metrics.MethodLabel:   tt.method,
				metrics.SuccessLabel:  tt.success,
				metrics.CodeLabel:     tt.code,
				metrics.EndpointLabel: tt.endpoint,
			}

			serverActiveRequestLabels := prometheus.Labels{
				metrics.MethodLabel:   tt.method,
				metrics.EndpointLabel: tt.endpoint,
			}

			clientLatencyLabels := prometheus.Labels{
				metrics.MethodLabel:     tt.method,
				metrics.SuccessLabel:    tt.success,
				metrics.ServerSlugLabel: serverSlug,
				metrics.ClientNameLabel: serverSlug,
			}

			clientTotalRequestLabels := prometheus.Labels{
				metrics.MethodLabel:     tt.method,
				metrics.SuccessLabel:    tt.success,
				metrics.CodeLabel:       tt.code,
				metrics.ServerSlugLabel: serverSlug,
				metrics.ClientNameLabel: serverSlug,
			}

			clientActiveRequestLabels := prometheus.Labels{
				metrics.MethodLabel:     tt.method,
				metrics.ServerSlugLabel: serverSlug,
				metrics.ClientNameLabel: serverSlug,
			}

			defer promtest.NewPrometheusMetricTest(t, "server latency", metrics.ServerLatency, serverSizeLabels).CheckSampleCountDelta(1)
			defer promtest.NewPrometheusMetricTest(t, "server total requests", metrics.ServerTotalRequests, serverTotalRequestLabels).CheckDelta(1)
			defer promtest.NewPrometheusMetricTest(t, "server active requests", metrics.ServerActiveRequests, serverActiveRequestLabels).CheckDelta(0)
			defer promtest.NewPrometheusMetricTest(t, "server request size", metrics.ServerRequestSize, serverSizeLabels).CheckDelta(float64(tt.reqSize))
			defer promtest.NewPrometheusMetricTest(t, "server response size", metrics.ServerResponseSize, serverSizeLabels).CheckDelta(float64(tt.respSize))
			defer promtest.NewPrometheusMetricTest(t, "client latency", metrics.ClientLatencyDistribution, clientLatencyLabels).CheckSampleCountDelta(1)
			defer promtest.NewPrometheusMetricTest(t, "client total requests", metrics.ClientTotalRequests, clientTotalRequestLabels).CheckDelta(1)
			defer promtest.NewPrometheusMetricTest(t, "client active requests", metrics.ClientActiveRequests, clientActiveRequestLabels).CheckDelta(0)
			defer spectest.ValidateSpec(t, "http", "server")
			defer spectest.ValidateSpec(t, "http", "client")

			if tt.method == http.MethodGet {
				_, err = client.Get(ts.URL + tt.route)
				if err != nil {
					t.Fatal("client.Get", err)
				}
			}

			if tt.method == http.MethodPost {
				input := exampleRequest{Input: "foo"}
				var body bytes.Buffer
				if err := json.NewEncoder(&body).Encode(input); err != nil {
					t.Fatal(err)
				}

				if _, err := client.Post(ts.URL+tt.route, "", &body); err != nil {
					t.Fatal("client.Post", err)
				}
			}
		})
	}
}
