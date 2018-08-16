package cfg

import (
	"strings"
	"testing"
	"time"
)

func TestParseCommon(t *testing.T) {
	var input = `
listen: ":8000"
maxProcs: 32
concurrencyLimit: 2048
maxIdleConnsPerHost: 1024
carbonsearch:
    backend: "http://127.0.0.1:8070"
    prefix: "virt.v1.*."
timeouts:
    global: "20s"
    afterStarted: "15s"
graphite09compat: true
backends:
        - "http://10.190.202.30:8080"
        - "http://10.190.197.9:8080"
logger:
    -
       logger: ""
       file: "/var/log/carbonzipper/carbonzipper.log"
       level: "info"
       encoding: "json"
`

	r := strings.NewReader(input)
	got, err := ParseCommon(r)
	if err != nil {
		t.Fatal(err)
	}

	expected := Common{
		Listen: ":8000",
		Backends: []string{
			"http://10.190.202.30:8080",
			"http://10.190.197.9:8080",
		},

		MaxProcs: 32,
		Timeouts: Timeouts{
			Global:       20 * time.Second,
			AfterStarted: 15 * time.Second,
			Connect:      200 * time.Millisecond,
		},
		ConcurrencyLimitPerServer: 2048,
		KeepAliveInterval:         30 * time.Second,
		MaxIdleConnsPerHost:       1024,

		ExpireDelaySec: 600,
		CarbonSearch: CarbonSearch{
			Backend: "http://127.0.0.1:8070",
			Prefix:  "virt.v1.*.",
		},
		GraphiteWeb09Compatibility: true,

		Buckets: 10,
		Graphite: GraphiteConfig{
			Pattern:  "{prefix}.{fqdn}",
			Host:     "127.0.0.1:3002",
			Interval: 60 * time.Second,
			Prefix:   "carbon.zipper",
		},
	}

	if !eqCommon(got, expected) {
		t.Fatalf("Didn't parse expected struct from config\nGot: %v\nExp: %v", got, expected)
	}
}

type comparable struct {
	Listen                     string
	MaxProcs                   int
	Timeouts                   Timeouts
	ConcurrencyLimitPerServer  int
	KeepAliveInterval          time.Duration
	MaxIdleConnsPerHost        int
	CarbonSearch               CarbonSearch
	ExpireDelaySec             int32
	GraphiteWeb09Compatibility bool
	Buckets                    int
	Graphite                   GraphiteConfig
}

func project(a Common) comparable {
	return comparable{
		Listen:                     a.Listen,
		MaxProcs:                   a.MaxProcs,
		Timeouts:                   a.Timeouts,
		ConcurrencyLimitPerServer:  a.ConcurrencyLimitPerServer,
		KeepAliveInterval:          a.KeepAliveInterval,
		MaxIdleConnsPerHost:        a.MaxIdleConnsPerHost,
		CarbonSearch:               a.CarbonSearch,
		ExpireDelaySec:             a.ExpireDelaySec,
		GraphiteWeb09Compatibility: a.GraphiteWeb09Compatibility,
		Buckets:                    a.Buckets,
		Graphite:                   a.Graphite,
	}
}

func eqCommon(a, b Common) bool {
	aa := project(a)
	bb := project(b)

	if aa != bb {
		return false
	}

	if len(a.Backends) != len(b.Backends) {
		return false
	}

	for i := range a.Backends {
		if a.Backends[i] != b.Backends[i] {
			return false
		}
	}

	return true
}
