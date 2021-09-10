package http

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
	"xjtuportal/component/basic"
)

func getUrl(protocol string, hostname string, path string) (url string) {
	url = fmt.Sprintf("%s://%s%s", protocol, hostname, path)
	return
}

type RequestHelper struct {
	loggerHelper    *basic.LoggerHelper
	requestSettings *basic.ProgramRequestSettings
}

func InitRequestHelper(configHelper *basic.ConfigHelper, loggerHelper *basic.LoggerHelper) (httpHelper *RequestHelper, err error) {

	if configHelper == nil {
		err = errors.New("http/request: ConfigHelper is invalid")
		return nil, err
	}

	if loggerHelper == nil {
		err := errors.New("http/request: logger is invalid")
		return nil, err
	}

	httpHelper = &RequestHelper{
		loggerHelper:    loggerHelper,
		requestSettings: &configHelper.ProgramSettings.ProgramRequestSettings,
	}

	return httpHelper, nil
}

func (requestHelper *RequestHelper) redirectPolicy(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

func (requestHelper *RequestHelper) SendRequest(
	url string, method string, data io.Reader, header *http.Header, cookies []*http.Cookie,
) (
	response *http.Response, body []byte, statusCode int, error error,
) {

	// Create request
	request, err := http.NewRequest(method, url, data)
	if err != nil {
		err = errors.New(fmt.Sprintf("http/request: Cannot create request [%v]", err))
		return nil, nil, -1, err
	}

	// Set header
	if header != nil {
		request.Header = *header
	}
	for key, value := range requestHelper.requestSettings.Header {
		request.Header.Set(key, value)
	}

	// Set cookies
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	// Create http client
	defaultTransport := &http.Transport{
		Proxy: nil, // No proxy
		DialContext: (&net.Dialer{
			Timeout: time.Duration(requestHelper.requestSettings.Connect.Timeout) * time.Second,
		}).DialContext,
	}

	client := &http.Client{
		Transport:     defaultTransport,
		CheckRedirect: requestHelper.redirectPolicy,
	}

	requestHelper.loggerHelper.AddLog(basic.DEBUG, fmt.Sprintf("http/request: Send [%s] request to [%s]", method, url))

	// Do request
	response, err = client.Do(request)

	// Request error
	if err != nil { // Request error
		return nil, nil, -1, err
	}

	// Response empty error
	if response == nil {
		err = errors.New("http/request: empty response")
		return nil, nil, -1, err
	}
	// Response not empty
	defer func() {
		err = response.Body.Close()
		if err != nil {
			response = nil
			body = nil
			statusCode = -1
			error = err
		} else {
			requestHelper.loggerHelper.AddLog(basic.DEBUG, "http/request: response body successfully closed")
		}
	}()

	requestHelper.loggerHelper.AddLog(basic.DEBUG, fmt.Sprintf("http/request: response\n%+v", *response))

	// Read respond body error
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, nil, -1, err
	}

	requestHelper.loggerHelper.AddLog(basic.DEBUG, fmt.Sprintf("http/request: response body\n%s", string(content)))

	// Response code error
	if response.StatusCode >= 400 {
		err = errors.New(fmt.Sprintf("http/request: response return error code [%d]", response.StatusCode))
		return nil, nil, response.StatusCode, err
	}

	return response, content, response.StatusCode, nil
}
