package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"xjtuportal/component/basic"
	"xjtuportal/component/device"
)

var (
	ipv4Regex = regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
)

type SessionListPortal struct {
	Concurrency string `json:"concurrency"`
	Sessions    []struct {
		DeviceType        string `json:"deviceType"`
		ExperienceEndTime int64  `json:"experienceEndTime"`
		Username          string `json:"user_name"`
		SessionId         string `json:"acct_session_id"`
		NasIpAddr         string `json:"nas_ip_address"`
		UserIpAddr        string `json:"framed_ip_address"`
		UserMacAddr       string `json:"calling_station_id"`
		StartTime         string `json:"acct_start_time"`
		UniqueId          string `json:"acct_unique_id"`
	} `json:"sessions"`
}

type Session struct {
	SessionId        string
	NasIpAddr        string
	UserIpAddr       string
	UserMacAddr      string
	StartTime        string
	UniqueId         string
	IsCurrentSession bool
}

type SessionListHelper struct {
	OnlineHelper    *OnlineHelper
	loggerHelper    *basic.LoggerHelper
	sessionSettings *basic.ProgramSessionSettings

	sessionListUrl string
	logoutUrl      string
	getIpUrl       string

	MacSessionMap  map[string]*Session
	SessionMacList []string
}

func InitSessionListHelper(
	configHelper *basic.ConfigHelper,
	loggerHelper *basic.LoggerHelper,
	requestHelper *RequestHelper,
) (*SessionListHelper, error) {

	onlineHelper, err := InitOnlineHelper(configHelper, loggerHelper, requestHelper)
	if err != nil {
		err = errors.New(fmt.Sprintf("http/session: Error creating OnlineHelper [%v]", err))
		return nil, err
	}

	sessionListHelper := &SessionListHelper{
		OnlineHelper:    onlineHelper,
		loggerHelper:    loggerHelper,
		sessionSettings: &configHelper.ProgramSettings.ProgramSessionSettings,
		sessionListUrl: configHelper.ProgramSettings.ProgramSessionSettings.PortalServer.Hostname +
			configHelper.ProgramSettings.ProgramSessionSettings.PortalServer.SessionListPath,
		logoutUrl: configHelper.ProgramSettings.ProgramSessionSettings.PortalServer.Hostname +
			configHelper.ProgramSettings.ProgramSessionSettings.PortalServer.LogoutPath,
		getIpUrl: configHelper.ProgramSettings.ProgramSessionSettings.SpeedCheckServer.Hostname +
			configHelper.ProgramSettings.ProgramSessionSettings.SpeedCheckServer.GetIpPath,
		MacSessionMap:  make(map[string]*Session),
		SessionMacList: make([]string, 0),
	}

	return sessionListHelper, nil

}

func (sessionListHelper *SessionListHelper) sessionListPortalGet() (sessionListPortal *SessionListPortal, err error) {

	_, err = sessionListHelper.OnlineHelper.GetAuthToken()
	if err != nil {
		err = errors.New(fmt.Sprintf("http/session: Cannot get token [%v]", err))
		return nil, err
	}

	header := &http.Header{}
	header.Set("Authorization", sessionListHelper.OnlineHelper.OnlineResponse.Token)
	cookies := make([]*http.Cookie, 0, 2)
	cookies = append(cookies, &http.Cookie{
		Name:  "token",
		Value: sessionListHelper.OnlineHelper.OnlineResponse.Token,
	})

	_, body, _, err := sessionListHelper.OnlineHelper.requestHelper.SendRequest(
		sessionListHelper.sessionListUrl,
		"GET",
		nil,
		header,
		cookies,
	)

	if err != nil {
		return nil, err
	}

	sessionListPortal = &SessionListPortal{}
	err = json.Unmarshal(body, sessionListPortal)
	if err != nil {
		return nil, err
	}

	return sessionListPortal, nil

}

func (sessionListHelper *SessionListHelper) InitSessionListByPortal() (statusCode int, err error) {

	sessionListHelper.SessionMacList = make([]string, 0)
	sessionListHelper.MacSessionMap = make(map[string]*Session)

	sessionListPortal, err := sessionListHelper.sessionListPortalGet()

	if err != nil {
		return statusCode, err
	}

	if concurrency, err := strconv.Atoi(sessionListPortal.Concurrency); err != nil || concurrency == 0 {
		err = errors.New("http/session: error getting concurrency")
		return -1, err
	}

	if len(sessionListPortal.Sessions) == 0 {
		sessionListHelper.loggerHelper.AddLog(basic.WARNING, "http/session: No session")
		return 200, nil
	}

	for _, sessionPortal := range sessionListPortal.Sessions {

		// Invalid session check
		if sessionPortal.UserMacAddr == "" ||
			sessionPortal.UserIpAddr == "" ||
			sessionPortal.UniqueId == "" {
			sessionListHelper.loggerHelper.AddLog(basic.WARNING, fmt.Sprintf("http/session: Invalid session:\n%+v", sessionPortal))
			continue
		}

		// Session with invalid MAC address check
		standardMac, err := device.MacStandardize(sessionPortal.UserMacAddr)
		if err != nil {
			sessionListHelper.loggerHelper.AddLog(basic.WARNING,
				fmt.Sprintf("http/session: Session with invalid MAC address [%s]", sessionPortal.UserMacAddr))
			continue
		}
		sessionPortal.UserMacAddr = standardMac

		// Session with invalid user IP address check
		standardIp := net.ParseIP(sessionPortal.UserIpAddr)
		if standardIp == nil {
			sessionListHelper.loggerHelper.AddLog(basic.WARNING,
				fmt.Sprintf("http/session: Session with invalid user IP address [%s]", sessionPortal.UserIpAddr))
			continue
		}
		sessionPortal.UserIpAddr = standardIp.String()

		// Duplicated session check
		if _, ok := sessionListHelper.MacSessionMap[sessionPortal.UserMacAddr]; ok {
			sessionListHelper.loggerHelper.AddLog(basic.WARNING,
				fmt.Sprintf("http/session: Duplicated session with MAC address [%s]", sessionPortal.UserMacAddr))
			continue
		}

		// After-handle the session
		session := &Session{
			SessionId:        sessionPortal.SessionId,
			NasIpAddr:        sessionPortal.NasIpAddr,
			UserIpAddr:       sessionPortal.UserIpAddr,
			UserMacAddr:      sessionPortal.UserMacAddr,
			StartTime:        sessionPortal.StartTime,
			UniqueId:         sessionPortal.UniqueId,
			IsCurrentSession: false,
		}

		// Add to list or map
		sessionListHelper.SessionMacList = append(sessionListHelper.SessionMacList, sessionPortal.UserMacAddr)
		sessionListHelper.MacSessionMap[sessionPortal.UserMacAddr] = session
	}

	return 200, nil

}

func (sessionListHelper *SessionListHelper) FindCurrentSessionBySpeedTestApp() (err error) {

	_, body, _, err := sessionListHelper.OnlineHelper.requestHelper.SendRequest(
		sessionListHelper.getIpUrl,
		"GET",
		nil,
		nil,
		make([]*http.Cookie, 0, 0),
	)
	if err != nil {
		return err
	}

	currentIp := net.ParseIP(ipv4Regex.FindAllString(string(body), -1)[0])
	if currentIp == nil {
		err = errors.New("http/session: cannot get a valid IP")
		return err
	}
	sessionListHelper.loggerHelper.AddLog(basic.INFO,
		fmt.Sprintf("http/session: Current session IP: %s", currentIp.String()))

	for key, session := range sessionListHelper.MacSessionMap {
		if session.UserIpAddr == currentIp.String() {
			session.IsCurrentSession = true
			sessionListHelper.MacSessionMap[key] = session
			return nil
		}
	}

	err = errors.New(fmt.Sprintf("http/session: there's no session with IP [%s]", currentIp.String()))
	return err

}

func (sessionListHelper *SessionListHelper) FindCurrentSessionByLocalMacList(macList []string) (err error) {

	for _, mac := range macList {
		if session, ok := sessionListHelper.MacSessionMap[mac]; ok {
			session.IsCurrentSession = true
			sessionListHelper.MacSessionMap[mac] = session
			return nil
		}
	}

	err = errors.New("http/session: there's no session that has MAC address in local MAC list")
	return err

}

func (sessionListHelper *SessionListHelper) LogoutDelete(uniqueId string) (statusCode int, err error) {

	statusCode, err = sessionListHelper.OnlineHelper.GetAuthToken()
	if err != nil {
		err = errors.New(fmt.Sprintf("http/session: Cannot get token [%v]", err))
		return statusCode, err
	}

	header := &http.Header{}
	header.Set("Authorization", sessionListHelper.OnlineHelper.OnlineResponse.Token)
	cookies := make([]*http.Cookie, 0, 2)
	cookies = append(cookies, &http.Cookie{
		Name:  "token",
		Value: sessionListHelper.OnlineHelper.OnlineResponse.Token,
	})

	_, _, statusCode, err = sessionListHelper.OnlineHelper.requestHelper.SendRequest(
		fmt.Sprintf("%s/%s", sessionListHelper.logoutUrl, uniqueId),
		"DELETE",
		nil,
		header,
		cookies,
	)

	return statusCode, err

}
