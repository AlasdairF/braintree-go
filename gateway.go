package braintree

import (
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
)

type Gateway interface {
	Execute(method, urlExtension string, body io.Reader) (*Response, error)
}

func NewGateway(config Configuration) BraintreeGateway {
	return BraintreeGateway{
		config: config,
		client: &http.Client{},
	}
}

type BraintreeGateway struct {
	config Configuration
	client *http.Client
}

func (this BraintreeGateway) Execute(method, urlExtension string, body io.Reader) (*Response, error) {
	request, err := http.NewRequest(method, this.config.BaseURL()+urlExtension, body)
	if err != nil {
		return nil, errors.New("Error creating HTTP request: " + err.Error())
	}

	request.Header.Set("Content-Type", "application/xml")
	request.Header.Set("Accept", "application/xml")
	request.Header.Set("Accept-Encoding", "gzip")
	request.Header.Set("User-Agent", "Braintree-Go")
	request.Header.Set("X-ApiVersion", "3")
	request.SetBasicAuth(this.config.publicKey, this.config.privateKey)

	response, err := this.client.Do(request)
	defer response.Body.Close()
	if err != nil {
		return nil, errors.New("Error sending request to Braintree: " + err.Error())
	}

	gzipBody, err := gzip.NewReader(response.Body)
	defer gzipBody.Close()
	if err != nil {
		return nil, errors.New("Error reading gzipped response from Braintree: " + err.Error())
	}

	contents, err := ioutil.ReadAll(gzipBody)
	if err != nil {
		return nil, errors.New("Error reading response from Braintree: " + err.Error())
	}

	return &Response{StatusCode: response.StatusCode, Status: response.Status, Body: contents}, nil
}

// Stub gateways, included for testing
type blowUpGateway struct{}

func (this blowUpGateway) Execute(method, url string, body io.Reader) (*Response, error) {
	return &Response{StatusCode: 500, Status: "500 Internal Server Error"}, nil
}

type badInputGateway struct{}

func (this badInputGateway) Execute(method, url string, body io.Reader) (*Response, error) {
	xml := "<?xml version=\"1.0\" encoding=\"UTF-8\"?><api-error-response><errors><errors type=\"array\"/></errors><message>Card Issuer Declined CVV</message></api-error-response>"
	return &Response{StatusCode: 422, Body: []byte(xml)}, nil
}

type notFoundGateway struct{}

func (this notFoundGateway) Execute(method, url string, body io.Reader) (*Response, error) {
	return &Response{StatusCode: 404}, nil
}
