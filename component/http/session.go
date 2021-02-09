package http

import (
	"auto-portal-auth/component/base"
	"auto-portal-auth/component/device"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
)

type Session struct {
	DeviceType        string `json:"deviceType"`
	ExperienceEndTime int64  `json:"experienceEndTime"`
	Username          string `json:"user_name"`
	SessionId         string `json:"acct_session_id"`
	NasIpAddr         string `json:"nas_ip_address"`
	UserIpAddr        string `json:"framed_ip_address"`
	UserMacAddr       string `json:"calling_station_id"`
	StartTime         string `json:"acct_start_time"`
	UniqueId          string `json:"acct_unique_id"`
	IsCurrentSession  bool   `json:"-"`
}

type SessionList struct {
	Concurrency int       `json:"concurrency"`
	Sessions    []Session `json:"sessions"`
}

type SessionListHelper struct {
	SessionList    *SessionList
	MacSessionMap  map[string]Session
	SessionMacList []string

	SessionListUrl string
	LogoutUrl      string
	GetIpUrl       string

	onlineHelper *OnlineHelper
	loggerHelper *base.LoggerHelper
}

func InitSessionListHelper(configHelper *base.ConfigHelper, loggerHelper *base.LoggerHelper) (*SessionListHelper, error) {

	if loggerHelper == nil {
		err := errors.New("logger is invalid")
		return nil, err
	}

	if configHelper == nil {
		err := errors.New("ConfigHelper is invalid")
		return nil, err
	}

	onlineHelper, err := InitOnlineHelper(configHelper, loggerHelper)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error creating OnlineHelper [%v]", err))
		return nil, err
	}

	sessionListHelper := &SessionListHelper{

		SessionListUrl: getUrl("http",
			configHelper.ProgramSettings.Api.PortalServer.Hostname,
			configHelper.ProgramSettings.Api.PortalServer.SessionListPath,
		),
		LogoutUrl: getUrl("http",
			configHelper.ProgramSettings.Api.PortalServer.Hostname,
			configHelper.ProgramSettings.Api.PortalServer.LogoutPath,
		),
		GetIpUrl: getUrl("http",
			configHelper.ProgramSettings.Api.SpeedCheckServer.Hostname,
			configHelper.ProgramSettings.Api.SpeedCheckServer.GetIpPath,
		),

		onlineHelper: onlineHelper,
		loggerHelper: loggerHelper,
	}

	return sessionListHelper, nil

}

func (sessionListHelper *SessionListHelper) SessionListGet() (statusCode int, err error) {
	statusCode, err = sessionListHelper.onlineHelper.GetAuthToken()

	if err != nil {
		err = errors.New(fmt.Sprintf("Cannot get token [%v]", err))
		return statusCode, err
	}

	request, err := http.NewRequest("GET", sessionListHelper.SessionListUrl, nil)
	if err != nil {
		err = errors.New(fmt.Sprintf("Cannot create online request [%v]", err))
		return -1, err
	}
	request.AddCookie(&http.Cookie{
		Name: "token", Value: sessionListHelper.onlineHelper.onlineResponse.Token,
	})
	request.Header.Set(
		"Authorization", sessionListHelper.onlineHelper.onlineResponse.Token,
	)

	sessionListHelper.loggerHelper.AddLog(base.DEBUG,
		fmt.Sprintf("Send SessionList Get to [%s]", sessionListHelper.SessionListUrl))

	response, err := sessionListHelper.onlineHelper.httpHelper.SendRequest(request)
	if err != nil {
		return -1, err
	}

	sessionListHelper.loggerHelper.AddLog(base.DEBUG,
		fmt.Sprintf("SessionList response:\n%+v", response))

	// Response code error
	if response.StatusCode != 200 {
		err = errors.New("session list response return error code")
		return response.StatusCode, err
	}

	// Read response body error
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return -1, err
	}

	sessionList := &SessionList{}
	err = json.Unmarshal(content, sessionListHelper)
	if err != nil {
		return -1, err
	}
	sessionListHelper.SessionList = sessionList

	return 200, nil

}

func (sessionListHelper *SessionListHelper) InitSessionList() (statusCode int, err error) {
	statusCode, err = sessionListHelper.SessionListGet()

	if err != nil {
		return statusCode, err
	}

	if sessionListHelper.SessionList.Concurrency == 0 {
		err = errors.New("error getting concurrency")
		return -1, err
	}

	if len(sessionListHelper.SessionList.Sessions) == 0 {
		sessionListHelper.loggerHelper.AddLog(base.INFO, "No session")
		return 200, nil
	}

	for _, session := range sessionListHelper.SessionList.Sessions {

		// Invalid session check
		if session.UserMacAddr == "" ||
			session.UserIpAddr == "" ||
			session.UniqueId == "" {
			sessionListHelper.loggerHelper.AddLog(base.WARNING, fmt.Sprintf("Invalid session:\n%+v", session))
			continue
		}

		// Session with invalid MAC address check
		standardMac, err := device.MacStandardize(session.UserMacAddr)
		if err != nil {
			sessionListHelper.loggerHelper.AddLog(base.WARNING,
				fmt.Sprintf("Session with invalid MAC address [%s]", session.UserMacAddr))
			continue
		}
		session.UserMacAddr = standardMac

		// Session with invalid user IP address check
		standardIp := net.ParseIP(session.UserIpAddr)
		if standardIp == nil {
			sessionListHelper.loggerHelper.AddLog(base.WARNING,
				fmt.Sprintf("Session with invalid user IP address [%s]", session.UserIpAddr))
			continue
		}
		session.UserIpAddr = standardIp.String()

		// Duplicated session check
		if _, ok := sessionListHelper.MacSessionMap[session.UserMacAddr]; ok {
			sessionListHelper.loggerHelper.AddLog(base.WARNING,
				fmt.Sprintf("Duplicated session with MAC address [%s]", session.UserMacAddr))
			continue
		}

		// After-handle the session
		session.IsCurrentSession = false

		// Add to list or map
		sessionListHelper.SessionMacList = append(sessionListHelper.SessionMacList, session.UserMacAddr)
		sessionListHelper.MacSessionMap[session.UserMacAddr] = session
	}

	return 200, nil

}

func (sessionListHelper *SessionListHelper) findCurrentSessionBySpeedTestApp() (err error) {
	request, err := http.NewRequest("GET", sessionListHelper.GetIpUrl, nil)
	if err != nil {
		err = errors.New(fmt.Sprintf("Cannot create GetIp request [%v]", err))
		return err
	}

	sessionListHelper.loggerHelper.AddLog(base.DEBUG,
		fmt.Sprintf("Send GetIP Get [%s]", sessionListHelper.GetIpUrl))

	response, err := sessionListHelper.onlineHelper.httpHelper.SendRequest(request)
	if err != nil {
		return err
	}

	sessionListHelper.loggerHelper.AddLog(base.DEBUG,
		fmt.Sprintf("GetIP response:\n%+v", response))

	// Response code error
	if response.StatusCode != 200 {
		err = errors.New("get ip response return error code")
		return err
	}

	// Read response body error
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	currentIp := net.ParseIP(string(content))
	if currentIp == nil {
		err = errors.New("cannot get a valid IP")
		return err
	}

	for key, session := range sessionListHelper.MacSessionMap {
		if session.UserIpAddr == currentIp.String() {
			session.IsCurrentSession = true
			sessionListHelper.MacSessionMap[key] = session
			return nil
		}
	}

	err = errors.New(fmt.Sprintf("there's no session with IP [%s]", currentIp.String()))
	return err

}

func (sessionListHelper *SessionListHelper) findCurrentSessionByKnownMacList(macList []string) (err error) {

	for _, mac := range macList {
		if session, ok := sessionListHelper.MacSessionMap[mac]; ok {
			session.IsCurrentSession = true
			sessionListHelper.MacSessionMap[mac] = session
			return nil
		}
	}

	err = errors.New("there's no session that has MAC address in known MAC list")
	return err

}

func (sessionListHelper *SessionListHelper) logoutDelete(uniqueId string) (statusCode int, err error) {

	statusCode, err = sessionListHelper.onlineHelper.GetAuthToken()

	if err != nil {
		err = errors.New(fmt.Sprintf("Cannot get token [%v]", err))
		return statusCode, err
	}

	request, err := http.NewRequest("DELETE",
		fmt.Sprintf("%s/%s", sessionListHelper.LogoutUrl, uniqueId),
		nil,
	)
	if err != nil {
		err = errors.New(fmt.Sprintf("Cannot create Logout request [%v]", err))
		return -1, err
	}
	request.AddCookie(&http.Cookie{
		Name: "token", Value: sessionListHelper.onlineHelper.onlineResponse.Token,
	})
	request.Header.Set(
		"Authorization", sessionListHelper.onlineHelper.onlineResponse.Token,
	)

	sessionListHelper.loggerHelper.AddLog(base.DEBUG,
		fmt.Sprintf("Send Logout Get to [%s]", sessionListHelper.LogoutUrl))

	response, err := sessionListHelper.onlineHelper.httpHelper.SendRequest(request)
	if err != nil {
		return -1, err
	}

	sessionListHelper.loggerHelper.AddLog(base.DEBUG,
		fmt.Sprintf("SessionList response:\n%+v", response))

	// Response code error
	if response.StatusCode != 200 {
		err = errors.New("logout response return error code")
		return response.StatusCode, err
	}
	return 200, nil

}
