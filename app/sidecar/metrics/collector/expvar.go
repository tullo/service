package collector

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"time"
)

// Expvar provides the ability to receive metrics
// from internal services using expvar.
type Expvar struct {
	host   string
	tr     *http.Transport
	client http.Client
}

// New creates a Expvar for collection metrics.
func New(host string) (*Expvar, error) {
	tr := http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          2,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	exp := Expvar{
		host: host,
		tr:   &tr,
		client: http.Client{
			Transport: &tr,
			Timeout:   1 * time.Second,
		},
	}

	return &exp, nil
}

// Collect fetches metrics from internal services using expvar
func (exp *Expvar) Collect() (map[string]interface{}, error) {
	r, err := http.NewRequest(http.MethodGet, exp.host, nil)
	if err != nil {
		return nil, err
	}

	// prevent re-use of TCP connections between requests
	r.Close = true

	resp, err := exp.client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(msg))
	}

	data := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}
