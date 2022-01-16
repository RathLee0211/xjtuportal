package app

import (
	"errors"
	"fmt"
	"strings"
	"xjtuportal/component/basic"
	"xjtuportal/component/device"
	"xjtuportal/component/http"
)

type DiagnosisShellHelper struct {
	loggerHelper             *basic.LoggerHelper
	connectivityChecker      *http.ConnectivityChecker
	proxyChecker             *http.ProxyHelper
	userUiSettings           *basic.UserUISettings
	programDiagnosisSettings *basic.ProgramDiagnosisSettings
	programShellSettings     *basic.ProgramShellSettings
	printHint                bool
}

func InitDiagnosisHelper(
	configHelper *basic.ConfigHelper,
	loggerHelper *basic.LoggerHelper,
	connectivityChecker *http.ConnectivityChecker,
	proxyChecker *http.ProxyHelper,
) (*DiagnosisShellHelper, error) {

	if configHelper == nil {
		err := errors.New("app/diagnosis: ConfigHelper is invalid")
		return nil, err
	}

	if loggerHelper == nil {
		err := errors.New("app/diagnosis: logger is invalid")
		return nil, err
	}

	if connectivityChecker == nil {
		err := errors.New("app/diagnosis: connectivityChecker is invalid")
		return nil, err
	}

	if proxyChecker == nil {
		err := errors.New("app/diagnosis: proxyChecker is invalid")
		return nil, err
	}

	initDiagnosisHelper := &DiagnosisShellHelper{
		loggerHelper:             loggerHelper,
		connectivityChecker:      connectivityChecker,
		proxyChecker:             proxyChecker,
		userUiSettings:           &configHelper.UserSettings.UserUISettings,
		programDiagnosisSettings: &configHelper.ProgramSettings.ProgramAppSettings.ProgramDiagnosisSettings,
		programShellSettings:     &configHelper.ProgramSettings.ProgramUiSettings.ProgramShellSettings,
		printHint:                configHelper.UserSettings.UserUISettings.Mode == basic.InteractMode,
	}
	return initDiagnosisHelper, nil

}

func (diagnosis *DiagnosisShellHelper) errorHandle(
	errorHandleMap map[int]basic.ErrorHandler,
	statusCode int,
) {
	errorHandler, ok := errorHandleMap[statusCode]
	if !ok {
		errorHandler = errorHandleMap[-1]
	}
	errorHandler.LogHandledError(diagnosis.loggerHelper, diagnosis.printHint)

}

func (diagnosis *DiagnosisShellHelper) DoDiagnosis() {

	if diagnosis.printHint {
		fmt.Println(diagnosis.programShellSettings.InteractHint.Diagnosis.Banner)
	}

	_, _, ipList, err := device.GetLocalInterfaceInfo()
	if err != nil {
		diagnosis.loggerHelper.AddLog(basic.INFO, fmt.Sprint("app/diagnosis: Error getting IP from interfaces: ", err))
		if diagnosis.printHint {
			fmt.Println(diagnosis.programShellSettings.InteractHint.BasicHint.Failed)
		}
		return
	}
	diagnosis.loggerHelper.AddLog(basic.INFO, fmt.Sprint("app/diagnosis: All vaild IPv4 address(es):\n", strings.Join(ipList, "\n")))
	if len(ipList) == 0 {
		diagnosis.loggerHelper.AddLog(basic.ERROR, fmt.Sprint("app/diagnosis: Cannot get any interface with valid IP"))
		if diagnosis.printHint {
			fmt.Println(diagnosis.programShellSettings.InteractHint.Diagnosis.NoIp)
		}
		return
	}

	diagnosis.loggerHelper.AddLog(basic.INFO, "app/diagnosis: Start checking network connectivity")

	diagnosis.loggerHelper.AddLog(basic.INFO, "app/diagnosis: Start HTTP check")

	// =============== Internet Check (baidu.com) ================
	diagnosis.loggerHelper.AddLog(basic.INFO, "app/diagnosis: Start internet connectivity check")
	statusCode, err := diagnosis.connectivityChecker.InternetHttpCheck()
	if err != nil {
		diagnosis.loggerHelper.AddLog(basic.WARNING, fmt.Sprintf("%v", err))
	}
	diagnosis.errorHandle(diagnosis.programDiagnosisSettings.ErrorHandle[basic.InternetErrors], statusCode)

	// =============== Intranet Check (p.xjtu.edu.cn) ================
	diagnosis.loggerHelper.AddLog(basic.INFO, "app/diagnosis: Start intranet connectivity check")
	statusCode, err = diagnosis.connectivityChecker.IntranetHttpCheck()
	if err != nil {
		diagnosis.loggerHelper.AddLog(basic.WARNING, fmt.Sprintf("%v", err))
	}
	diagnosis.errorHandle(diagnosis.programDiagnosisSettings.ErrorHandle[basic.IntranetErrors], statusCode)

	diagnosis.loggerHelper.AddLog(basic.INFO, "app/diagnosis: Start DNS check")

	// =============== System DNS resolve Check ================
	diagnosis.loggerHelper.AddLog(basic.INFO, "app/diagnosis: Start system DNS check")
	statusCode, err = diagnosis.connectivityChecker.SystemResolveCheck()
	if err != nil {
		diagnosis.loggerHelper.AddLog(basic.ERROR, fmt.Sprintf("%v", err))
	}
	diagnosis.errorHandle(diagnosis.programDiagnosisSettings.ErrorHandle[basic.ResolverErrors], statusCode)

	// =============== Internet DNS Check (aliDNS, 114DNS, ...) ================
	diagnosis.loggerHelper.AddLog(basic.INFO, "app/diagnosis: Start internet DNS check")

	available, unavailable := diagnosis.connectivityChecker.InternetDnsCheck()
	if len(available) > 0 {
		diagnosis.loggerHelper.AddLog(basic.INFO,
			fmt.Sprintf("app/diagnosis: The following Internet DNS is available:\n%s", strings.Join(available, ", ")))
		if diagnosis.printHint {
			fmt.Printf("%s\n%s\n", diagnosis.programShellSettings.InteractHint.Diagnosis.InterAvailable, strings.Join(available, "\n"))
		}
	} else {
		diagnosis.loggerHelper.AddLog(basic.ERROR, "app/diagnosis: No internet DNS available")
		if diagnosis.printHint {
			fmt.Println(diagnosis.programShellSettings.InteractHint.Diagnosis.InterUnavailable)
		}
	}
	if len(unavailable) > 0 {
		diagnosis.loggerHelper.AddLog(basic.INFO,
			fmt.Sprintf("app/diagnosis: The following Internet DNS is unavailable:\n%s", strings.Join(unavailable, ", ")))
	}

	// =============== Intranet DNS Check (10.6.39.2, 202.117.0.20, ...) ================
	diagnosis.loggerHelper.AddLog(basic.INFO, "app/diagnosis: Start intranet DNS check")
	available, unavailable = diagnosis.connectivityChecker.IntranetDnsCheck()
	if len(available) > 0 {
		diagnosis.loggerHelper.AddLog(basic.INFO,
			fmt.Sprintf("app/diagnosis: The following Intranet DNS is available:\n%s", strings.Join(available, ", ")))
		if diagnosis.printHint {
			fmt.Printf("%s\n%s\n", diagnosis.programShellSettings.InteractHint.Diagnosis.IntraAvailable, strings.Join(available, "\n"))
		}
	} else {
		diagnosis.loggerHelper.AddLog(basic.ERROR, "app/diagnosis: No intranet DNS available")
		if diagnosis.printHint {
			fmt.Println(diagnosis.programShellSettings.InteractHint.Diagnosis.IntraUnavailable)
		}
	}
	if len(unavailable) > 0 {
		diagnosis.loggerHelper.AddLog(basic.INFO,
			fmt.Sprintf("app/diagnosis: The following Intranet DNS is unavailable:\n%s", strings.Join(unavailable, ", ")))
	}

	// =============== Proxy check ================
	diagnosis.loggerHelper.AddLog(basic.INFO, "app/diagnosis: Start local proxy detecting")
	proxies, programs, proxyAvailable := diagnosis.proxyChecker.ProxyCheck()
	noAvail := true
	if len(proxies) > 0 {
		var proxyList []string
		for i := 0; i < len(proxies); i++ {
			avail := "x"
			if proxyAvailable[i] {
				avail = "o"
				noAvail = false
			}
			proxyList = append(proxyList, fmt.Sprintf("%s (%s) %s", proxies[i], programs[i], avail))
		}

		diagnosis.loggerHelper.AddLog(basic.INFO, fmt.Sprint("app/diagnosis: Proxies found:", "\n", strings.Join(proxyList, "\n")))
		if diagnosis.printHint {
			fmt.Println(diagnosis.programShellSettings.InteractHint.Diagnosis.ProxyFound)
			fmt.Println(strings.Join(proxyList, "\n"))
		}
		if noAvail {
			diagnosis.loggerHelper.AddLog(basic.WARNING, "app/diagnosis: No proxy available")
			if diagnosis.printHint {
				fmt.Println(diagnosis.programShellSettings.InteractHint.Diagnosis.NoProxyAvailable)
			}
		}

	}

}
