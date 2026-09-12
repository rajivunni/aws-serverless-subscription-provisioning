package provider

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	client              *http.Client
	baseUrl             *url.URL
	accountId           int
	companyId           int
	provisioningOrderId int
}

type Options struct {
	ClientId            string
	ClientSecret        string
	ApiUrl              string
	AuthUrl             string
	AccountId           int
	CompanyId           int
	ProvisioningOrderId int
}

type loggingRoundTripper struct {
	rt http.RoundTripper
}

type listResponse[T any] struct {
	Results []T `json:"result"`
	Count   int `json:"count"`
}

func (l *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := l.rt.RoundTrip(req)
	if err != nil {
		log.Print("Provider HTTP request failed; details omitted")
		return nil, err
	}
	log.Printf("Provider HTTP request completed: method=%s status=%d", req.Method, resp.StatusCode)
	return resp, nil
}

func NewClient(options Options) (*Client, error) {
	baseUrl, err := url.Parse(options.ApiUrl)
	if err != nil {
		return nil, err
	}
	lg := &loggingRoundTripper{rt: &retryRoundTripper{rt: http.DefaultTransport, maxRetries: 2, baseDelay: time.Second * 2}}
	auth, err := newAuthenticator(options.AuthUrl, options.ClientId, options.ClientSecret, lg)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{
		Transport: auth,
		Timeout:   time.Second * 8,
	}
	return &Client{
		client:              httpClient,
		accountId:           options.AccountId,
		companyId:           options.CompanyId,
		provisioningOrderId: options.ProvisioningOrderId,
		baseUrl:             baseUrl,
	}, nil
}
