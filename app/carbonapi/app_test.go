package carbonapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bookingcom/carbonapi/cache"
	"github.com/bookingcom/carbonapi/cfg"
	"github.com/bookingcom/carbonapi/pkg/backend"
	"github.com/bookingcom/carbonapi/pkg/backend/mock"
	dataTypes "github.com/bookingcom/carbonapi/pkg/types"
	"github.com/bookingcom/carbonapi/pkg/types/encoding/json"

	"github.com/lomik/zapwriter"
	"github.com/stretchr/testify/assert"
)

var testApp *App

func find(ctx context.Context, request dataTypes.FindRequest) (dataTypes.Matches, error) {
	return getMetricGlobResponse(request.Query), nil
}

func info(ctx context.Context, request dataTypes.InfoRequest) ([]dataTypes.Info, error) {
	return getMockInfoResponse(), nil
}

func getMockInfoResponse() []dataTypes.Info {
	return []dataTypes.Info{
		dataTypes.Info{
			Host:              "http://127.0.0.1:8080",
			Name:              "foo.bar",
			AggregationMethod: "Average",
			MaxRetention:      157680000,
			XFilesFactor:      0.5,
			Retentions: []dataTypes.Retention{
				dataTypes.Retention{
					SecondsPerPoint: 60,
					NumberOfPoints:  43200,
				},
			},
		},
	}
}

func render(ctx context.Context, request dataTypes.RenderRequest) ([]dataTypes.Metric, error) {
	return []dataTypes.Metric{
		dataTypes.Metric{
			Name:      "foo.bar",
			StartTime: 1510913280,
			StopTime:  1510913880,
			StepTime:  60,
			Values:    []float64{0, 1510913759, 1510913818},
			IsAbsent:  []bool{true, false, false},
		},
	}, nil
}

func getMetricGlobResponse(metric string) dataTypes.Matches {
	match := dataTypes.Match{
		Path:   metric,
		IsLeaf: true,
	}

	switch metric {
	case "foo.bar*":
		return dataTypes.Matches{
			Name:    "foo.bar",
			Matches: []dataTypes.Match{match},
		}

	case "foo.bar":
		return dataTypes.Matches{
			Name:    "foo.bar",
			Matches: []dataTypes.Match{match},
		}

	case "foo.b*":
		return dataTypes.Matches{
			Name: "foo.b",
			Matches: []dataTypes.Match{
				match,
				dataTypes.Match{
					Path:   "foo.bat",
					IsLeaf: true,
				},
			},
		}
	}

	return dataTypes.Matches{}
}

func init() {
	testApp = setUpTestConfig()
}

func setUpTestConfig() *App {
	c := cfg.GetDefaultLoggerConfig()
	c.Level = "none"
	zapwriter.ApplyConfig([]zapwriter.Config{c})
	logger := zapwriter.Logger("main")

	config := cfg.DefaultAPIConfig()

	// TODO (grzkv): Should use New
	app := &App{
		config:            config,
		queryCache:        cache.NewMemcached("capi", ``),
		findCache:         cache.NewExpireCache(1000),
		prometheusMetrics: newPrometheusMetrics(config),
	}
	app.backends = []backend.Backend{
		mock.New(mock.Config{
			Find:   find,
			Info:   info,
			Render: render,
		}),
	}

	app.config.ConcurrencyLimitPerServer = 1024

	setUpConfig(app, logger)
	initHandlers(app)

	return app
}

func setUpRequest(t *testing.T, url string) (*http.Request, *httptest.ResponseRecorder) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	return req, rr
}

func TestRenderHandler(t *testing.T) {
	req, rr := setUpRequest(t, "/render/?target=fallbackSeries(foo.bar,foo.baz)&from=-10minutes&format=json&noCache=1")
	testApp.renderHandler(rr, req)

	expected := `[{"target":"foo.bar","datapoints":[[null,1510913280],[1510913759,1510913340],[1510913818,1510913400]]}]`

	// Check the status code is what we expect.
	r := assert.Equal(t, rr.Code, http.StatusOK, "HttpStatusCode should be 200 OK.")
	if !r {
		t.Error("HttpStatusCode should be 200 OK.")
	}
	r = assert.Equal(t, expected, rr.Body.String(), "Http response should be same.")
	if !r {
		t.Error("Http response should be same.")
	}
}

func TestFindHandler(t *testing.T) {
	req, rr := setUpRequest(t, "/metrics/find/?query=foo.bar&format=json")
	testApp.findHandler(rr, req)

	body := rr.Body.String()
	// LOL this test is so fragile
	expected := "[{\"allowChildren\":0,\"context\":{},\"expandable\":0,\"id\":\"foo.bar\",\"leaf\":1,\"text\":\"bar\"}]"
	r := assert.Equal(t, rr.Code, http.StatusOK, "HttpStatusCode should be 200 OK.")
	if !r {
		t.Error("HttpStatusCode should be 200 OK.")
	}
	r = assert.Equal(t, string(expected), body, "Http response should be same.")
	if !r {
		t.Error("Http response should be same.")
	}
}

func TestFindHandlerCompleter(t *testing.T) {
	testMetrics := []string{"foo.b/", "foo.bar"}
	for _, testMetric := range testMetrics {
		req, rr := setUpRequest(t, "/metrics/find/?query="+testMetric+"&format=completer")
		testApp.findHandler(rr, req)
		body := rr.Body.String()
		expectedValue, _ := findCompleter(getMetricGlobResponse(getCompleterQuery(testMetric)))
		r := assert.Equal(t, rr.Code, http.StatusOK, "HttpStatusCode should be 200 OK.")
		if !r {
			t.Error("HttpStatusCode should be 200 OK.")
		}
		r = assert.Equal(t, string(expectedValue), body, "Http response should be same.")
		if !r {
			t.Error("Http response should be same.")
		}
	}
}

func TestInfoHandler(t *testing.T) {
	req, rr := setUpRequest(t, "/info/?target=foo.bar&format=json")
	testApp.infoHandler(rr, req)

	body := rr.Body.String()
	expected := getMockInfoResponse()
	expectedJson, err := json.InfoEncoder(expected)
	r := assert.Nil(t, err)
	if !r {
		t.Errorf("err should be nil, %v instead", err)
	}

	r = assert.Equal(t, rr.Code, http.StatusOK, "HttpStatusCode should be 200 OK.")
	if !r {
		t.Error("Http response should be same.")
	}
	r = assert.Equal(t, string(expectedJson), body, "Http response should be same.")
	if !r {
		t.Error("Http response should be same.")
	}
}
