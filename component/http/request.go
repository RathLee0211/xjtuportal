package http

import (
	"auto-portal-auth/component/base"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

func getUrl(protocol string, hostname string, path string) (url string) {
	url = fmt.Sprintf("%s://%s%s", protocol, hostname, path)
	return
}

type RequestHelper struct {
	UserAgent      string
	AcceptLanguage string
	TimeoutSec     int
}

func InitRequestHeader(configHelper *base.ConfigHelper) (httpHelper *RequestHelper, err error) {

	if configHelper == nil {
		err = errors.New("ConfigHelper is invalid")
		return nil, err
	}

	httpHelper = &RequestHelper{
		UserAgent:      configHelper.ProgramSettings.Http.Header.UserAgent,
		AcceptLanguage: configHelper.ProgramSettings.Http.Header.AcceptLanguage,
		TimeoutSec:     configHelper.ProgramSettings.Http.Connect.Timeout,
	}

	return httpHelper, nil
}

func (httpHelper *RequestHelper) setBasicHeader(request *http.Request) error {
	if request == nil {
		return errors.New("invalid request while setting header")
	}
	request.Header.Set("Accept-Language", httpHelper.AcceptLanguage)
	request.Header.Set("User-Agent", httpHelper.UserAgent)
	return nil
}

func (httpHelper *RequestHelper) SendRequest(request *http.Request) (*http.Response, error) {
	defaultTransport := &http.Transport{
		Proxy: nil,
		DialContext: (&net.Dialer{
			Timeout: time.Duration(httpHelper.TimeoutSec) * time.Second,
		}).DialContext,
	}
	client := &http.Client{Transport: defaultTransport}
	err := httpHelper.setBasicHeader(request)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil { // Error happens
		return nil, err
	}

	// Response empty error
	if response == nil {
		err = errors.New("empty response")
		return nil, err
	}

	return response, response.Body.Close()
}
