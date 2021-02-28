package http

import (
	"auto-portal-auth/component/basic"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type AuthData struct {
	DeviceType  string `json:"deviceType"`
	RedirectUrl string `json:"redirectUrl"`
	DataType    string `json:"type"`
	Username    string `json:"webAuthUser"`
	Password    string `json:"webAuthPassword"`
}

type OnlineResponse struct {
	ReturnCode  int    `json:"statusCode"`
	Truncated   bool   `json:"truncated"`
	CreatedTs   int64  `json:"createdAt"`
	ErrorCode   int    `json:"error"`
	Description string `json:"errorDescription"`
	Token       string `json:"token"`
}

type OnlineHelper struct {
	loggerHelper          *basic.LoggerHelper
	requestHelper         *RequestHelper
	userOnlineSettings    *basic.UserOnlineSettings
	programOnlineSettings *basic.ProgramOnlineSettings

	onlineUrl       string
	RedirectUrl     string
	fakeRedirectUrl string
	authData        *AuthData
	OnlineResponse  *OnlineResponse
}

func InitOnlineHelper(configHelper *basic.ConfigHelper,
	loggerHelper *basic.LoggerHelper,
	requestHelper *RequestHelper,
) (*OnlineHelper, error) {

	if configHelper == nil {
		err := errors.New("http/online: ConfigHelper is invalid")
		return nil, err
	}

	if loggerHelper == nil {
		err := errors.New("http/online: logger is invalid")
		return nil, err
	}

	if requestHelper == nil {
		err := errors.New("http/online: RequestHelper is invalid")
		return nil, err
	}

	onlineHelper := &OnlineHelper{
		loggerHelper:          loggerHelper,
		requestHelper:         requestHelper,
		userOnlineSettings:    &configHelper.UserSettings.UserOnlineSettings,
		programOnlineSettings: &configHelper.ProgramSettings.ProgramOnlineSettings,
		onlineUrl: getUrl("http",
			configHelper.ProgramSettings.ProgramOnlineSettings.PortalServer.Hostname,
			configHelper.ProgramSettings.ProgramOnlineSettings.PortalServer.OnlinePath,
		),
		RedirectUrl: "",
		fakeRedirectUrl: getUrl("http",
			configHelper.ProgramSettings.ProgramOnlineSettings.PortalServer.Hostname,
			configHelper.ProgramSettings.ProgramOnlineSettings.PortalServer.FakeRedirectPath,
		),
		authData: &AuthData{
			DeviceType:  "PC",
			RedirectUrl: "",
			DataType:    "login",
			Username: fmt.Sprintf("%s@%s",
				configHelper.UserSettings.UserOnlineSettings.AuthData.Username,
				configHelper.UserSettings.UserOnlineSettings.AuthData.Domain),
			Password: configHelper.UserSettings.UserOnlineSettings.AuthData.Password,
		},
		OnlineResponse: nil,
	}

	return onlineHelper, nil
}

func (onlineHelper *OnlineHelper) OnlinePost(redirectUrl string) (int, error) {

	onlineHelper.authData.RedirectUrl = redirectUrl

	var data bytes.Buffer
	enc := json.NewEncoder(&data)
	enc.SetEscapeHTML(false)
	err := enc.Encode(onlineHelper.authData)
	if err != nil {
		err = errors.New(fmt.Sprintf("http/online: Cannot create online request json data [%v]", err))
		return -1, err
	}

	header := &http.Header{}
	header.Set("Content-Type", "application/json")

	cookies := make([]*http.Cookie, 0, 2)
	cookies = append(cookies, &http.Cookie{Name: "redirectUrl", Value: redirectUrl})

	_, body, statusCode, err := onlineHelper.requestHelper.SendRequest(
		onlineHelper.onlineUrl,
		"POST",
		&data,
		header,
		cookies,
	)

	if err != nil {
		return statusCode, err
	}

	// Parse json error
	onlineResponse := &OnlineResponse{}
	err = json.Unmarshal(body, onlineResponse)
	if err != nil {
		return -1, err
	}
	onlineHelper.OnlineResponse = onlineResponse

	return 200, nil

}

func (onlineHelper *OnlineHelper) GetAuthToken() (statusCode int, err error) {
	// Request encountered error
	statusCode, err = onlineHelper.OnlinePost(onlineHelper.fakeRedirectUrl)
	if err != nil {
		return statusCode, err
	}

	// Missing token error
	if onlineHelper.OnlineResponse.Token == "" {
		err = errors.New("http/online: empty token")
		return -1, err
	}

	onlineHelper.loggerHelper.AddLog(basic.DEBUG,
		fmt.Sprintf("http/online: Successfully get token: [%s]", onlineHelper.OnlineResponse.Token))

	return 200, nil

}

func (onlineHelper *OnlineHelper) GetRedirectUrl() (statusCode int, err error) {
	// TODO: implementation
	response, _, statusCode, err := onlineHelper.requestHelper.SendRequest(
		onlineHelper.programOnlineSettings.BootStrapUrl,
		"GET",
		nil,
		nil,
		make([]*http.Cookie, 0, 0),
	)
	if err != nil {
		return
	}

	url := response.Header.Get("Location")
	if url == "" {
		err = errors.New("http/online: cannot get redirect url")
		return -1, err

	}
	onlineHelper.RedirectUrl = url
	return 200, nil
}
