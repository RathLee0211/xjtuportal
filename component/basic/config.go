package basic

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

const (
	// System type
	Windows = "windows"
	Linux   = "linux"

	// Error handlers
	LoginErrors      = "login_errors"
	LogoutErrors     = "logout_errors"
	GetSessionErrors = "get_session_errors"
	InternetErrors   = "internet_check_errors"
	IntranetErrors   = "intranet_check_errors"
	ResolverErrors   = "resolve_check_errors"

	// Login error return codes
	SessionOverload = 39

	// UI modes
	InteractMode = "interact"
)

type UserOnlineSettings struct {
	AuthData struct {
		Domain   string `yaml:"domain"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"auth_data"`
}

type UserDeviceSettings struct {
	KnownMacList []string `yaml:"known_mac_list,flow"`
	UseInterface bool     `yaml:"use_interface"`
}

type UserPortalSettings struct {
	IsAutoLogout bool `yaml:"auto_logout"`
}

type UserLoggerSettings struct {
	OutputWriter []string `yaml:"output_writer,flow"`
	Level        string   `yaml:"level"`
	FilePath     string   `yaml:"file_path"`
	UseColor     bool     `yaml:"color"`
}

type UserUISettings struct {
	Mode string `yaml:"mode"`
}

type UserSettings struct {
	UserOnlineSettings UserOnlineSettings `yaml:"online"`
	UserDeviceSettings UserDeviceSettings `yaml:"device"`
	UserAppSettings    struct {
		UserPortalSettings UserPortalSettings `yaml:"portal"`
	} `yaml:"session"`
	UserLoggerSettings UserLoggerSettings `yaml:"logger"`
	UserUISettings     UserUISettings     `yaml:"ui"`
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

type ProgramRequestSettings struct {
	Header  map[string]string `yaml:"header"`
	Connect struct {
		Timeout int `yaml:"timeout"`
	} `yaml:"connect"`
}

type ProgramDnsSettings struct {
	Connect struct {
		Timeout int `yaml:"timeout"`
	} `yaml:"connect"`
	Testing struct {
		Times int `yaml:"times"`
	}
}

type ProgramConnectivitySettings struct {
	Http struct {
		Internet string `yaml:"internet"`
		Intranet string `yaml:"intranet"`
	} `yaml:"http"`
	Dns struct {
		Domain struct {
			Internet string `yaml:"internet"`
			Intranet string `yaml:"intranet"`
		} `yaml:"domain"`
		Server struct {
			Internet []string `yaml:"internet,flow"`
			Intranet []string `yaml:"intranet,flow"`
		} `yaml:"server"`
	} `yaml:"dns"`
}

type ProgramOnlineSettings struct {
	PortalServer struct {
		Hostname         string `yaml:"hostname"`
		FakeRedirectPath string `yaml:"fake_redirect_path"`
		OnlinePath       string `yaml:"online_path"`
	} `yaml:"portal_server"`
	BootStrapUrl string `yaml:"bootstrap_url"`
}

type ProgramSessionSettings struct {
	PortalServer struct {
		Hostname        string `yaml:"hostname"`
		SessionListPath string `yaml:"session_list_path"`
		LogoutPath      string `yaml:"logout_path"`
	} `yaml:"portal_server"`
	SpeedCheckServer struct {
		Hostname  string `yaml:"hostname"`
		GetIpPath string `yaml:"get_ip_path"`
	} `yaml:"speed_check_server"`
}

type ProgramLoggerSettings struct {
	LogLevelNumber int `yaml:"-"`
	OutputFormat   struct {
		Datetime  string `yaml:"datetime"`
		LogRecord string `yaml:"log_record"`
	} `yaml:"output_format"`
	MaxInfoLength int `yaml:"max_info_length"`
}

type ErrorHandler struct {
	ErrorCode        int    `yaml:"error_code"`
	HintMessage      string `yaml:"hint_message"`
	ErrorDescription string `yaml:"error_description"`
	LogLevel         string `yaml:"log_level"`
	LogMessage       string `yaml:"log_message"`
}

func (error *ErrorHandler) LogHandledError(loggerHelper *LoggerHelper, printHint bool) {
	if printHint {
		fmt.Println(error.HintMessage)
	}
	if loggerHelper != nil {
		loggerHelper.AddLog(logLevelNumbers[error.LogLevel],
			fmt.Sprintf("basic/config/LogHandledError: %s", error.LogMessage))
	}
}

type ProgramPortalSettings struct {
	SessionList struct {
		SessionRecord  string `yaml:"session_record"`
		SessionInfo    string `yaml:"session_info"`
		CurrentSession string `yaml:"current_session"`
	} `yaml:"session_list"`
	ErrorHandle map[string]map[int]ErrorHandler `yaml:"error_handle"`
}

type ProgramDiagnosisSettings struct {
	ErrorHandle map[string]map[int]ErrorHandler `yaml:"error_handle"`
}

type ProgramShellSettings struct {
	InteractHint struct {
		BasicHint struct {
			ReturnMain    string `yaml:"return_main"`
			KeySelect     string `yaml:"key_select"`
			CommandSelect string `yaml:"command_select"`
			SelectError   string `yaml:"select_error"`
			Pause         string `yaml:"pause"`
			Success       string `yaml:"success"`
			Failed        string `yaml:"failed"`
		} `yaml:"basic_hint"`
		MainMenu struct {
			Banner string `yaml:"banner"`
		} `yaml:"main_menu"`
		QuickSetting struct {
			Banner     string `yaml:"banner"`
			Username   string `yaml:"username"`
			Password   string `yaml:"password"`
			AutoLogout string `yaml:"auto_logout"`
			Confirm    string `yaml:"confirm"`
		} `yaml:"quick_setting"`
		SessionList struct {
			Banner string `yaml:"banner"`
		} `yaml:"session_list"`
		Diagnosis struct {
			Banner           string `yaml:"banner"`
			IntraAvailable   string `yaml:"intranet_dns_available"`
			InterAvailable   string `yaml:"internet_dns_available"`
			IntraUnavailable string `yaml:"intranet_dns_unavailable"`
			InterUnavailable string `yaml:"internet_dns_unavailable"`
		} `yaml:"diagnosis"`
	} `yaml:"interact_hint"`
}

type ProgramSettings struct {
	ProgramRequestSettings      ProgramRequestSettings      `yaml:"request"`
	ProgramDnsSettings          ProgramDnsSettings          `yaml:"dns"`
	ProgramConnectivitySettings ProgramConnectivitySettings `yaml:"connectivity"`
	ProgramOnlineSettings       ProgramOnlineSettings       `yaml:"online"`
	ProgramSessionSettings      ProgramSessionSettings      `yaml:"session"`
	ProgramLoggerSettings       ProgramLoggerSettings       `yaml:"logger"`
	ProgramAppSettings          struct {
		ProgramPortalSettings    ProgramPortalSettings    `yaml:"portal"`
		ProgramDiagnosisSettings ProgramDiagnosisSettings `yaml:"diagnosis"`
	} `yaml:"app"`
	ProgramUiSettings struct {
		ProgramShellSettings ProgramShellSettings `yaml:"shell"`
	} `yaml:"ui"`
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
		err = errors.New(fmt.Sprintf("basic/config: Read User Settings config error [%v]", err))
		return nil, err
	}
	programSettings, err := InitProgramSettings(programSettingsFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("basic/config: Read Program Settings config error [%v]", err))
		return nil, err
	}

	configHelper.UserSettings = userSettings
	configHelper.ProgramSettings = programSettings

	return configHelper, nil

}
