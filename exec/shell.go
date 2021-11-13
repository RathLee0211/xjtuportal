package exec

import (
	"bufio"
	"fmt"
	"github.com/eiannone/keyboard"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"xjtuportal/component/app"
	"xjtuportal/component/basic"
	"xjtuportal/component/device"
	"xjtuportal/component/http"
)

var (
	clearMapFunc = map[string]func(){
		basic.Windows: func() {
			cmd := exec.Command("cmd", "/c", "cls")
			cmd.Stdout = os.Stdout
			_ = cmd.Run()
		},
		basic.Linux: func() {
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			_ = cmd.Run()
		},
	}
)

func pause(hint string) {
	fmt.Println(hint)
	_, _, _ = keyboard.GetSingleKey()
}

type ShellUi struct {
	portal          *app.PortalShellHelper
	diagnosis       *app.DiagnosisShellHelper
	configHelper    *basic.ConfigHelper
	loggerHelper    *basic.LoggerHelper
	configDir       string
	loginFlag       bool
	logoutFlag      int
	showSessionFlag bool
	diagnosisFlag   bool
}

func InitShellUi(
	versionFlag bool,
	configFlag string,
	loginFlag bool,
	logoutFlag int,
	showSessionFlag bool,
	diagnosisFlag bool,
	adapterFlag bool,
) *ShellUi {

	if versionFlag {
		fmt.Println(app.ProgramInfo())
		return nil
	}

	if adapterFlag {
		ifList, _, _, err := device.GetLocalInterfaceInfo()
		if err != nil {
			fmt.Println(err)
		} else if len(ifList) == 0 {
			fmt.Println("Cannot get any interfaces")
		} else {
			for _, i := range ifList {
				fmt.Println(i)
			}
		}
		return nil
	}

	configHelper, err := basic.InitConfigHelper(
		fmt.Sprintf("%s/%s", configFlag, basic.UserConfigFile),
		fmt.Sprintf("%s/%s", configFlag, basic.ProgramConfigFile),
	)
	if err != nil {
		basic.LoggerTemp.AddLog(basic.FATAL, fmt.Sprintf("%v", err))
		pause("Press any key to exit...")
		return nil
	}

	loggerHelper, err := basic.InitLoggerHelper(configHelper)
	if err != nil {
		basic.LoggerTemp.AddLog(basic.FATAL, fmt.Sprintf("%v", err))
		pause("Press any key to exit...")
		return nil
	}
	loggerHelper.AddLog(basic.INFO, "Basic module successfully initialized")

	requestHelper, err := http.InitRequestHelper(configHelper, loggerHelper)
	if err != nil {
		basic.LoggerTemp.AddLog(basic.FATAL, fmt.Sprintf("%v", err))
		return nil
	}
	loggerHelper.AddLog(basic.DEBUG, "RequestHelper successfully initialized")

	dnsHelper, err := http.InitDnsHelper(configHelper, loggerHelper)
	if err != nil {
		basic.LoggerTemp.AddLog(basic.FATAL, fmt.Sprintf("%v", err))
		return nil
	}
	loggerHelper.AddLog(basic.DEBUG, "DnsHelper successfully initialized")

	connectivityChecker, err := http.InitConnectivityChecker(configHelper, loggerHelper, requestHelper, dnsHelper)
	if err != nil {
		basic.LoggerTemp.AddLog(basic.FATAL, fmt.Sprintf("%v", err))
		return nil
	}
	loggerHelper.AddLog(basic.DEBUG, "ConnectivityChecker successfully initialized")

	diagnosisHelper, err := app.InitDiagnosisHelper(configHelper, loggerHelper, connectivityChecker)
	if err != nil {
		basic.LoggerTemp.AddLog(basic.FATAL, fmt.Sprintf("%v", err))
		return nil
	}
	loggerHelper.AddLog(basic.DEBUG, "DiagnosisShellHelper successfully initialized")

	sessionListHelper, err := http.InitSessionListHelper(configHelper, loggerHelper, requestHelper)
	if err != nil {
		basic.LoggerTemp.AddLog(basic.FATAL, fmt.Sprintf("%v", err))
		return nil
	}
	loggerHelper.AddLog(basic.DEBUG, "SessionListHelper successfully initialized")

	interfaceHelper, err := device.InitInterfaceHelper(configHelper, loggerHelper)
	if err != nil {
		basic.LoggerTemp.AddLog(basic.FATAL, fmt.Sprintf("%v", err))
		return nil
	}
	loggerHelper.AddLog(basic.DEBUG, "InterfaceHelper successfully initialized")

	portalHelper, err := app.InitPortalShellHelper(configHelper, loggerHelper, connectivityChecker, sessionListHelper, interfaceHelper)
	if err != nil {
		basic.LoggerTemp.AddLog(basic.FATAL, fmt.Sprintf("%v", err))
		return nil
	}
	loggerHelper.AddLog(basic.DEBUG, "PortalShellHelper successfully initialized")

	shellUi := &ShellUi{
		portal:          portalHelper,
		diagnosis:       diagnosisHelper,
		configHelper:    configHelper,
		loggerHelper:    loggerHelper,
		configDir:       configFlag,
		loginFlag:       loginFlag,
		logoutFlag:      logoutFlag,
		showSessionFlag: showSessionFlag,
		diagnosisFlag:   diagnosisFlag,
	}
	basic.LoggerTemp.AddLog(basic.INFO, "All modules successfully initialized")
	return shellUi

}

func (shellUi *ShellUi) clearScreen() {
	if value, ok := clearMapFunc[runtime.GOOS]; ok {
		value()
	}
}

func (shellUi *ShellUi) getInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	str, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	regex := regexp.MustCompile(`[\r|\n]`)
	str = regex.ReplaceAllString(str, "")
	return str, nil

}

func (shellUi *ShellUi) quickSettingInteract() bool {

	interactHint := shellUi.configHelper.ProgramSettings.ProgramUiSettings.ProgramShellSettings.InteractHint
	fmt.Println(interactHint.QuickSetting.Banner)

	fmt.Println(interactHint.QuickSetting.Username)
	username, _ := shellUi.getInput()

	fmt.Println(interactHint.QuickSetting.Password)
	password, _ := shellUi.getInput()

	fmt.Println(interactHint.QuickSetting.AutoLogout)
	autoLogout, _ := shellUi.getInput()

	// Confirmation
	fmt.Println("==========================")
	fmt.Printf("%s%s\n", interactHint.QuickSetting.Username, username)
	fmt.Printf("%s%s\n", interactHint.QuickSetting.Password, password)
	fmt.Printf("%s%s\n", interactHint.QuickSetting.AutoLogout, autoLogout)

	fmt.Println(interactHint.QuickSetting.Confirm)
	confirm, _ := shellUi.getInput()

	// Write new config to file
	if confirm == "y" {
		shellUi.configHelper.UserSettings.UserOnlineSettings.AuthData.Username = username
		shellUi.configHelper.UserSettings.UserOnlineSettings.AuthData.Password = password
		shellUi.configHelper.UserSettings.UserAppSettings.UserPortalSettings.IsAutoLogout = autoLogout == "y"
		ret, err := yaml.Marshal(&shellUi.configHelper.UserSettings)
		if err != nil { // Convert to yaml error
			shellUi.loggerHelper.AddLog(basic.DEBUG, fmt.Sprintf("%v", err))
			fmt.Println(interactHint.BasicHint.Failed)
			return false
		}
		// Write to file
		err = ioutil.WriteFile(fmt.Sprintf("%s/%s", shellUi.configDir, basic.UserConfigFile), ret, 0644)
		if err != nil { // Write to file err
			shellUi.loggerHelper.AddLog(basic.DEBUG, fmt.Sprintf("%v", err))
			fmt.Println(interactHint.BasicHint.Failed)
			return false
		}
		fmt.Println(interactHint.BasicHint.Success)
		return true
	} else {
		fmt.Println(interactHint.BasicHint.Failed)
		return false
	}
}

func (shellUi *ShellUi) logoutInteract() {

	interactHint := shellUi.configHelper.ProgramSettings.ProgramUiSettings.ProgramShellSettings.InteractHint

	err := shellUi.portal.DoListSession()
	if err != nil {
		return
	}
	fmt.Println(interactHint.BasicHint.CommandSelect)
	sessionIndexStr, err := shellUi.getInput()
	if err != nil {
		fmt.Println(interactHint.BasicHint.SelectError)
		return
	}
	sessionIndex, err := strconv.Atoi(sessionIndexStr)
	if err != nil {
		fmt.Println(interactHint.BasicHint.SelectError)
		return
	}
	shellUi.portal.DoLogout(sessionIndex)
}

func (shellUi *ShellUi) interactExec() (exit bool) {

	exit = true

	interactHint := shellUi.configHelper.ProgramSettings.ProgramUiSettings.ProgramShellSettings.InteractHint
	for {
		// Clear terminal
		shellUi.clearScreen()
		// Main menu
		fmt.Println(interactHint.MainMenu.Banner)
		fmt.Println(interactHint.BasicHint.KeySelect)
		char, _, err := keyboard.GetSingleKey()
		if err != nil {
			fmt.Println(interactHint.BasicHint.SelectError)
			pause(interactHint.BasicHint.Pause)
			continue
		}
		// Command handle
		switch char {
		case '0':
			{
				shellUi.clearScreen()
				fmt.Println(app.ProgramInfo())
				pause(interactHint.BasicHint.Pause)
			}
		case '1':
			{
				shellUi.clearScreen()
				ret := shellUi.quickSettingInteract()
				pause(interactHint.BasicHint.Pause)
				exit = false
				if ret {
					return
				}
			}
		case '2':
			{
				shellUi.clearScreen()
				shellUi.portal.DoLogin()
				pause(interactHint.BasicHint.Pause)
			}
		case '3':
			{
				shellUi.clearScreen()
				_ = shellUi.portal.DoListSession()
				pause(interactHint.BasicHint.Pause)
			}
		case '4':
			{
				shellUi.clearScreen()
				shellUi.logoutInteract()
				pause(interactHint.BasicHint.Pause)
			}
		case '5':
			{
				shellUi.clearScreen()
				shellUi.diagnosis.DoDiagnosis()
				pause(interactHint.BasicHint.Pause)
			}
		case '6':
			{
				shellUi.clearScreen()
				shellUi.loggerHelper.SetLogLevel(basic.DEBUG)
				fmt.Println(interactHint.BasicHint.Success)
				pause(interactHint.BasicHint.Pause)
			}
		case '7':
			{
				shellUi.clearScreen()
				ifList, _, _, err := device.GetLocalInterfaceInfo()
				if err != nil {
					fmt.Println(interactHint.BasicHint.Failed)
					shellUi.loggerHelper.AddLog(basic.ERROR, err.Error())
				} else if len(ifList) == 0 {
					fmt.Println(interactHint.BasicHint.Failed)
					shellUi.loggerHelper.AddLog(basic.ERROR, "Cannot get any interfaces")
				} else {
					for _, i := range ifList {
						fmt.Println(i)
					}
				}
				pause(interactHint.BasicHint.Pause)
			}
		case 'q':
			{
				return
			}
		default:
			{
				shellUi.clearScreen()
				fmt.Println(interactHint.BasicHint.SelectError)
				pause(interactHint.BasicHint.Pause)
			}
		}
	}
}

func (shellUi *ShellUi) Exec() (exit bool) {

	exit = true

	if !shellUi.loginFlag &&
		shellUi.logoutFlag == -1 &&
		!shellUi.showSessionFlag &&
		!shellUi.diagnosisFlag {
		if shellUi.configHelper.UserSettings.UserUISettings.Mode == basic.InteractMode {
			exit = shellUi.interactExec()
			return
		} else {
			shellUi.portal.DoLogin()
			return
		}
	}

	if shellUi.loginFlag {
		shellUi.portal.DoLogin()
		return
	}

	if shellUi.logoutFlag > -1 {
		_ = shellUi.portal.DoListSession()
		shellUi.portal.DoLogout(shellUi.logoutFlag)
		return
	}

	if shellUi.showSessionFlag {
		_ = shellUi.portal.DoListSession()
		return
	}

	if shellUi.diagnosisFlag {
		shellUi.diagnosis.DoDiagnosis()
		return
	}

	return

}
