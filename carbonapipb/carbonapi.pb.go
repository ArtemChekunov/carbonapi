package carbonapipb

import (
	"net/http"
	"net/url"

	"github.com/go-graphite/carbonapi/util"
)

type AccessLogDetails struct {
	Handler                       string            `json:"handler,omitempty"`
	CarbonapiUuid                 string            `json:"carbonapi_uuid,omitempty"`
	Username                      string            `json:"username,omitempty"`
	Url                           string            `json:"url,omitempty"`
	PeerIp                        string            `json:"peer_ip,omitempty"`
	PeerPort                      string            `json:"peer_port,omitempty"`
	Host                          string            `json:"host,omitempty"`
	Referer                       string            `json:"referer,omitempty"`
	Format                        string            `json:"format,omitempty"`
	UseCache                      bool              `json:"use_cache,omitempty"`
	HeadersData                   map[string]string `json:"headers_data,omitempty"`
	RequestMethod                 string            `json:"request_method,omitempty"`
	Targets                       []string          `json:"targets,omitempty"`
	CacheTimeout                  int32             `json:"cache_timeout,omitempty"`
	Metrics                       []string          `json:"metrics,omitempty"`
	HaveNonFatalErrors            bool              `json:"have_non_fatal_errors,omitempty"`
	Runtime                       float64           `json:"runtime,omitempty"`
	HttpCode                      int32             `json:"http_code,omitempty"`
	CarbonzipperResponseSizeBytes int64             `json:"carbonzipper_response_size_bytes,omitempty"`
	CarbonapiResponseSizeBytes    int64             `json:"carbonapi_response_size_bytes,omitempty"`
	Reason                        string            `json:"reason,omitempty"`
	SendGlobs                     bool              `json:"send_globs,omitempty"`
	From                          int32             `json:"from,omitempty"`
	Until                         int32             `json:"until,omitempty"`
	Tz                            string            `json:"tz,omitempty"`
	FromRaw                       string            `json:"from_raw,omitempty"`
	UntilRaw                      string            `json:"until_raw,omitempty"`
	Uri                           string            `json:"uri,omitempty"`
	FromCache                     bool              `json:"from_cache"`
	ZipperRequests                int64             `json:"zipper_requests,omitempty"`
}

func splitAddr(addr string) (string, string) {
	u, err := url.Parse(addr)
	if err != nil {
		return "unknown", "unknown"
	}

	return u.Hostname(), u.Port()
}

func NewAccessLogDetails(r *http.Request, handler string) AccessLogDetails {
	username, _, _ := r.BasicAuth()
	srcIP, srcPort := splitAddr(r.RemoteAddr)

	return AccessLogDetails{
		Handler:       handler,
		Username:      username,
		CarbonapiUuid: util.GetUUID(r.Context()),
		Url:           r.URL.RequestURI(),
		PeerIp:        srcIP,
		PeerPort:      srcPort,
		Host:          r.Host,
		Referer:       r.Referer(),
		Uri:           r.RequestURI,
		HttpCode:      http.StatusOK,
	}
}
