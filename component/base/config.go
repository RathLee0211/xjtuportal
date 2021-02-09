package base

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type UserSettings struct {
	AuthData struct {
		Domain   string `yaml:"domain"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"auth_data"`
	SessionManage struct {
		IsAutoLogout bool     `yaml:"auto_logout"`
		LanIpCidr    string   `yaml:"lan_ip_cidr"`
		DeviceList   []string `yaml:"device_list,flow"`
	} `yaml:"session_manage"`
	Log struct {
		LogOutput []string `yaml:"log_output,flow"`
		LogLevel  string   `yaml:"log_level"`
		LogFile   string   `yaml:"log_file"`
		UseColor  bool     `yaml:"log_color"`
	} `yaml:"log"`
	Options struct {
		Mode string `yaml:"mode"`
	} `yaml:"interface"`
}

func InitUserSettings(confPath string) (*UserSettings, error) {
	userSettings := &UserSettings{}
	userSettingFile, err := ioutil.ReadFile(confPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(userSettingFile, userSettings)
	if err != nil {
		return nil, err
	}
	return userSettings, nil
}

type ErrorHandler struct {
	HintMessage string            `yaml:"hint_message"`
	LogMessage  string            `yaml:"log_message"`
	HandleFunc  func(interface{}) `yaml:"-"`
}

type ProgramSettings struct {
	Http struct {
		Header struct {
			UserAgent      string `yaml:"user_agent"`
			AcceptLanguage string `yaml:"accept_language"`
		} `yaml:"header"`
		Connect struct {
			Timeout int `yaml:"timeout"`
		} `yaml:"connect"`
	} `yaml:"http"`
	Api struct {
		NetworkCheck struct {
			InternetCheckHostname string `yaml:"internet_check_hostname"`
			IntranetCheckHostname string `yaml:"intranet_check_hostname"`
			DnsCheckHostname      string `yaml:"dns_check_hostname"`
		} `yaml:"network_check"`
		PortalServer struct {
			Hostname         string `yaml:"hostname"`
			FakeRedirectPath string `yaml:"fake_redirect_path"`
			LoginPath        string `yaml:"login_path"`
			SessionListPath  string `yaml:"session_list_path"`
			LogoutPath       string `yaml:"logout_path"`
		} `yaml:"portal_server"`
		SpeedCheckServer struct {
			Hostname  string `yaml:"hostname"`
			GetIpPath string `yaml:"get_ip_path"`
		} `yaml:"speed_check_server"`
	} `yaml:"api"`
	Errors struct {
		LoginErrors []struct {
			ErrorDescription string `yaml:"error_description"`
			ErrorCode        int    `yaml:"error_code"`
		} `yaml:"login_errors,flow"`
	} `yaml:"errors"`
	Log struct {
		Datetime  string `yaml:"datetime"`
		LogRecord string `yaml:"log_record"`
	} `yaml:"log"`
	Session struct {
		SessionRecord string `yaml:"session_record"`
		SessionInfo   string `yaml:"session_info"`
	} `yaml:"session"`
	ErrorHandle map[string]map[int]ErrorHandler `yaml:"error_handle"`
}

func InitProgramSettings(confPath string) (*ProgramSettings, error) {
	programSettings := &ProgramSettings{}
	programSettingFile, err := ioutil.ReadFile(confPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(programSettingFile, programSettings)
	if err != nil {
		return nil, err
	}
	return programSettings, nil
}

type ConfigHelper struct {
	UserSettings    *UserSettings
	ProgramSettings *ProgramSettings
}

func InitConfigHelper(
	userSettingsFile string,
	programSettingsFile string,
) (*ConfigHelper, error) {

	configHelper := &ConfigHelper{}

	userSettings, err := InitUserSettings(userSettingsFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("%s [%v]", "Read User Settings config error", err))
		return nil, err
	}
	programSettings, err := InitProgramSettings(programSettingsFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("%s [%v]", "Read Program Settings config error", err))
		return nil, err
	}

	configHelper.UserSettings = userSettings
	configHelper.ProgramSettings = programSettings

	return configHelper, nil

}
