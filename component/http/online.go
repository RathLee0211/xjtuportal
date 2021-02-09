package http

import (
	"auto-portal-auth/component/base"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
	loggerHelper    *base.LoggerHelper
	httpHelper      *RequestHelper
	Url             string
	RedirectUrl     string
	FakeRedirectUrl string
	AuthData        *AuthData
	onlineResponse  *OnlineResponse
}

func InitOnlineHelper(configHelper *base.ConfigHelper, loggerHelper *base.LoggerHelper) (*OnlineHelper, error) {

	if loggerHelper == nil {
		err := errors.New("logger is invalid")
		return nil, err
	}

	if configHelper == nil {
		err := errors.New("ConfigHelper is invalid")
		return nil, err
	}

	httpHelper, err := InitRequestHeader(configHelper)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error when creating RequestHelper [%v]", err))
		return nil, err
	}

	onlineHelper := &OnlineHelper{
		loggerHelper: loggerHelper,
		httpHelper:   httpHelper,
		Url: getUrl("http",
			configHelper.ProgramSettings.Api.PortalServer.Hostname,
			configHelper.ProgramSettings.Api.PortalServer.LoginPath,
		),
		RedirectUrl: "",
		FakeRedirectUrl: getUrl("http",
			configHelper.ProgramSettings.Api.PortalServer.Hostname,
			configHelper.ProgramSettings.Api.PortalServer.FakeRedirectPath,
		),
		AuthData: &AuthData{
			DeviceType:  "PC",
			RedirectUrl: "",
			DataType:    "login",
			Username: fmt.Sprintf("%s@%s",
				configHelper.UserSettings.AuthData.Username,
				configHelper.UserSettings.AuthData.Domain),
			Password: configHelper.UserSettings.AuthData.Password,
		},
	}

	return onlineHelper, nil
}

func (onlineHelper *OnlineHelper) OnlinePost(redirectUrl string) (int, error) {
	data, err := json.Marshal(onlineHelper.AuthData)
	if err != nil {
		err = errors.New(fmt.Sprintf("Cannot create online request json data [%v]", err))
		return -1, err
	}
	request, err := http.NewRequest("POST", onlineHelper.Url, bytes.NewBuffer(data))
	if err != nil {
		err = errors.New(fmt.Sprintf("Cannot create online request [%v]", err))
		return -1, err
	}
	request.AddCookie(&http.Cookie{Name: "redirectUrl", Value: redirectUrl})
	request.Header.Set("Content-Type", "application/json")

	onlineHelper.loggerHelper.AddLog(base.DEBUG, fmt.Sprintf("Send Online Post to [%s]", onlineHelper.Url))

	response, err := onlineHelper.httpHelper.SendRequest(request)
	if err != nil {
		return -1, err
	}

	onlineHelper.loggerHelper.AddLog(base.DEBUG, fmt.Sprintf("Online response:\n%+v", response))
	// Response code error
	if response.StatusCode != 200 {
		err = errors.New("online response return error code")
		return response.StatusCode, err
	}

	// Read response body error
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return -1, err
	}

	// Parse json error
	onlineResponse := &OnlineResponse{}
	err = json.Unmarshal(content, onlineResponse)
	if err != nil {
		return -1, err
	}
	onlineHelper.onlineResponse = onlineResponse

	return 200, nil

}

func (onlineHelper *OnlineHelper) GetAuthToken() (statusCode int, err error) {
	// Request encountered error
	statusCode, err = onlineHelper.OnlinePost(onlineHelper.FakeRedirectUrl)
	if err != nil {
		return statusCode, err
	}

	// Missing token error
	if onlineHelper.onlineResponse.Token == "" {
		err = errors.New("empty token")
		return -1, err
	}

	onlineHelper.loggerHelper.AddLog(base.INFO, "Successfully get token via online post")

	return 200, nil

}
