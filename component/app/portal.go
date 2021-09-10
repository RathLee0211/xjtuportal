package app

import (
	"errors"
	"fmt"
	"strings"
	"xjtuportal/component/basic"
	"xjtuportal/component/device"
	"xjtuportal/component/http"
)

type PortalShellHelper struct {
	loggerHelper        *basic.LoggerHelper
	connectivityChecker *http.ConnectivityChecker
	sessionListHelper   *http.SessionListHelper
	interfaceHelper     *device.InterfaceHelper

	userPortalSettings       *basic.UserPortalSettings
	userUiSettings           *basic.UserUISettings
	programPortalSettings    *basic.ProgramPortalSettings
	programDiagnosisSettings *basic.ProgramDiagnosisSettings
	programShellSettings     *basic.ProgramShellSettings
	printHint                bool
}

func InitPortalShellHelper(
	configHelper *basic.ConfigHelper,
	loggerHelper *basic.LoggerHelper,
	connectivityChecker *http.ConnectivityChecker,
	sessionListHelper *http.SessionListHelper,
	interfaceHelper *device.InterfaceHelper,
) (*PortalShellHelper, error) {

	if configHelper == nil {
		err := errors.New("app/portal: ConfigHelper is invalid")
		return nil, err
	}

	if loggerHelper == nil {
		err := errors.New("app/portal: logger is invalid")
		return nil, err
	}

	if connectivityChecker == nil {
		err := errors.New("app/portal: connectivityChecker is invalid")
		return nil, err
	}

	if sessionListHelper == nil {
		err := errors.New("app/portal: sessionListHelper is invalid")
		return nil, err
	}

	if interfaceHelper == nil {
		err := errors.New("app/portal: macListHelper is invalid")
		return nil, err
	}

	portalHelper := &PortalShellHelper{
		loggerHelper:        loggerHelper,
		connectivityChecker: connectivityChecker,
		sessionListHelper:   sessionListHelper,
		interfaceHelper:     interfaceHelper,

		userPortalSettings:       &configHelper.UserSettings.UserAppSettings.UserPortalSettings,
		userUiSettings:           &configHelper.UserSettings.UserUISettings,
		programPortalSettings:    &configHelper.ProgramSettings.ProgramAppSettings.ProgramPortalSettings,
		programDiagnosisSettings: &configHelper.ProgramSettings.ProgramAppSettings.ProgramDiagnosisSettings,
		programShellSettings:     &configHelper.ProgramSettings.ProgramUiSettings.ProgramShellSettings,
		printHint:                configHelper.UserSettings.UserUISettings.Mode == basic.InteractMode,
	}

	return portalHelper, nil
}

func (portal *PortalShellHelper) errorHandle(
	errorHandleMap map[int]basic.ErrorHandler,
	statusCode int,
) {

	errorHandler, ok := errorHandleMap[statusCode]
	if !ok {
		errorHandler = errorHandleMap[-1]
	}
	errorHandler.LogHandledError(portal.loggerHelper, portal.printHint)
}

func (portal *PortalShellHelper) loginErrorCodeMapping(errorCode int, errorDescription string) int {
	for statusCode, errorHandler := range portal.programPortalSettings.ErrorHandle[basic.LoginErrors] {
		if errorCode == errorHandler.ErrorCode &&
			strings.Contains(errorDescription, errorHandler.ErrorDescription) {
			return statusCode
		}
	}
	return -1
}

func (portal *PortalShellHelper) login() (statusCode int, err error) {

	statusCode, err = portal.connectivityChecker.InternetHttpCheck()
	portal.errorHandle(portal.programDiagnosisSettings.ErrorHandle[basic.InternetErrors], statusCode)
	if err == nil { // Currently Internet is available
		return
	}
	portal.loggerHelper.AddLog(basic.INFO, fmt.Sprintf("%v", err))

	statusCode, err = portal.connectivityChecker.IntranetHttpCheck()
	portal.errorHandle(portal.programDiagnosisSettings.ErrorHandle[basic.IntranetErrors], statusCode)
	if err != nil { // Currently portal server is unavailable
		return
	}

	portal.loggerHelper.AddLog(basic.INFO, "app/portal: Try to login")

	statusCode, err = portal.sessionListHelper.OnlineHelper.GetRedirectUrl()
	if err != nil { // Cannot get redirect URL
		return
	}

	statusCode, err = portal.sessionListHelper.OnlineHelper.OnlinePost(portal.sessionListHelper.OnlineHelper.RedirectUrl)
	if err != nil { // Cannot get online response
		return
	}
	statusCode = portal.loginErrorCodeMapping(portal.sessionListHelper.OnlineHelper.OnlineResponse.ErrorCode,
		portal.sessionListHelper.OnlineHelper.OnlineResponse.Description)
	portal.errorHandle(portal.programPortalSettings.ErrorHandle[basic.LoginErrors], statusCode)
	return
}

func (portal *PortalShellHelper) getSessionList() (err error) {

	portal.loggerHelper.AddLog(basic.INFO, "app/portal: Try to get session list")

	statusCode, err := portal.sessionListHelper.InitSessionListByPortal()
	portal.errorHandle(portal.programPortalSettings.ErrorHandle[basic.GetSessionErrors], statusCode)
	return err

}

func (portal *PortalShellHelper) logout(macAddr string) (err error) {

	statusCode, err := portal.connectivityChecker.IntranetHttpCheck()
	portal.errorHandle(portal.programPortalSettings.ErrorHandle[basic.IntranetErrors], statusCode)
	if err != nil { // Currently portal server is unavailable
		return
	}

	if session, ok := portal.sessionListHelper.MacSessionMap[macAddr]; ok {
		portal.loggerHelper.AddLog(basic.INFO, fmt.Sprintf("app/portal: Try to logout session with MAC address [%s]", macAddr))
		statusCode, err = portal.sessionListHelper.LogoutDelete(session.UniqueId)
		portal.errorHandle(portal.programPortalSettings.ErrorHandle[basic.LogoutErrors], statusCode)
		return
	} else {
		err = errors.New(fmt.Sprintf("app/portal: There is no session with MAC address [%s]", macAddr))
		return
	}

}

func (portal *PortalShellHelper) DoLogin() {
	statusCode, err := portal.login()
	if err != nil {
		portal.loggerHelper.AddLog(basic.ERROR, fmt.Sprintf("%v", err))
	}

	if portal.userPortalSettings.IsAutoLogout && statusCode == basic.SessionOverload {
		err = portal.getSessionList()
		if err != nil {
			portal.loggerHelper.AddLog(basic.ERROR, fmt.Sprintf("%v", err))
			return
		}
		logoutMacAddr := portal.interfaceHelper.FindLogoutMac(portal.sessionListHelper.SessionMacList)
		err = portal.logout(logoutMacAddr)
		if err != nil {
			portal.loggerHelper.AddLog(basic.ERROR, fmt.Sprintf("%v", err))
		} else {
			_, err = portal.login()
			if err != nil {
				portal.loggerHelper.AddLog(basic.ERROR, fmt.Sprintf("%v", err))
			}
		}
	}

}

func (portal *PortalShellHelper) DoListSession() (err error) {

	err = portal.getSessionList()

	if err != nil {
		portal.loggerHelper.AddLog(basic.ERROR, fmt.Sprintf("%v", err))
		return
	}

	// Get session number
	sessionNumber := len(portal.sessionListHelper.SessionMacList)
	if portal.printHint {
		fmt.Println(fmt.Sprintf(portal.programShellSettings.InteractHint.SessionList.Banner, sessionNumber))
	}
	sessionStrList := make([]string, 0, len(portal.sessionListHelper.SessionMacList))

	// Try to get current session
	err = portal.sessionListHelper.FindCurrentSessionBySpeedTestApp()
	if err != nil {
		portal.loggerHelper.AddLog(basic.WARNING, fmt.Sprintf("%v", err))
		err = portal.sessionListHelper.FindCurrentSessionByLocalMacList(portal.interfaceHelper.LocalMacList)
		if err != nil {
			portal.loggerHelper.AddLog(basic.WARNING, fmt.Sprintf("%v", err))
		}
	}

	for index, mac := range portal.sessionListHelper.SessionMacList {
		if session, ok := portal.sessionListHelper.MacSessionMap[mac]; ok {
			currentSession := ""
			if session.IsCurrentSession {
				currentSession = portal.programPortalSettings.SessionList.CurrentSession
			}
			sessionStr := fmt.Sprintf(portal.programPortalSettings.SessionList.SessionRecord, index,
				fmt.Sprintf(portal.programPortalSettings.SessionList.SessionInfo,
					session.UserMacAddr,
					session.UserIpAddr,
					session.StartTime,
					currentSession,
				))
			sessionStrList = append(sessionStrList, sessionStr)
		}
	}

	if sessionNumber > 0 {
		portal.loggerHelper.AddLog(basic.INFO,
			fmt.Sprintf("app/portal: Current Sessions:\n%s", strings.Join(sessionStrList, "\n")))
		if portal.printHint {
			fmt.Println(strings.Join(sessionStrList, "\n"))
		}
		return nil
	} else {
		err = errors.New("app/portal: no session")
		return err
	}

}

func (portal *PortalShellHelper) DoLogout(sessionIndex int) {

	if sessionIndex >= len(portal.sessionListHelper.SessionMacList) {
		portal.loggerHelper.AddLog(basic.ERROR, fmt.Sprintf("app/portal: No session [%d] exists", sessionIndex))
		if portal.printHint {
			fmt.Println(portal.programShellSettings.InteractHint.BasicHint.SelectError)
		}
		return
	}

	logoutMacAddr := portal.sessionListHelper.SessionMacList[sessionIndex]
	err := portal.logout(logoutMacAddr)
	if err != nil {
		portal.loggerHelper.AddLog(basic.ERROR, fmt.Sprintf("%v", err))
	}
}
