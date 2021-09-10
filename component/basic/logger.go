package basic

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
	"xjtuportal/component/utils"
)

const (
	STDOUT = "stdout"
	FILE   = "file"
)

const (
	DEBUG   = 0
	INFO    = 1
	WARNING = 2
	ERROR   = 3
	FATAL   = 4
	MUTE    = 5
)

type logLevelHelper struct {
	logLevelName  string
	logLevelColor color.Attribute
}

type LoggerHelper struct {
	logger                *log.Logger
	userLoggerSettings    *UserLoggerSettings
	programLoggerSettings *ProgramLoggerSettings
}

var (
	logLevelNumbers = map[string]int{
		"DEBUG":   DEBUG,
		"INFO":    INFO,
		"WARNING": WARNING,
		"ERROR":   ERROR,
		"FATAL":   FATAL,
		"MUTE":    MUTE,
	}
	logLevelInfoMap = map[int]logLevelHelper{
		DEBUG:   {"DEBUG", color.FgWhite},
		INFO:    {"INFO", color.FgHiGreen},
		WARNING: {"WARNING", color.FgHiYellow},
		ERROR:   {"ERROR", color.FgHiRed},
		FATAL:   {"FATAL", color.FgHiMagenta},
		MUTE:    {},
	}
	LoggerTemp = &LoggerHelper{
		logger: log.New(os.Stdout, "", 0),
		userLoggerSettings: &UserLoggerSettings{
			Level:    "FATAL",
			UseColor: true,
		},
		programLoggerSettings: &ProgramLoggerSettings{
			LogLevelNumber: 4,
			MaxInfoLength:  1000,
			OutputFormat: struct {
				Datetime  string `yaml:"datetime"`
				LogRecord string `yaml:"log_record"`
			}{
				Datetime:  "2006-01-02 15:04:05",
				LogRecord: "[%s] [%s] %s",
			},
		},
	}
)

func InitLoggerHelper(configHelper *ConfigHelper) (*LoggerHelper, error) {

	if configHelper == nil {
		err := errors.New("configHelper is invalid")
		return nil, err
	}

	// Disable default logger
	log.SetOutput(ioutil.Discard)

	loggerHelper := &LoggerHelper{
		userLoggerSettings:    &configHelper.UserSettings.UserLoggerSettings,
		programLoggerSettings: &configHelper.ProgramSettings.ProgramLoggerSettings,
	}

	// Set output type
	logFilePath := configHelper.UserSettings.UserLoggerSettings.FilePath
	_, logOutputMap := utils.RemoveDuplicateStrings(configHelper.UserSettings.UserLoggerSettings.OutputWriter)
	_, isStdout := logOutputMap[STDOUT]
	_, isFileOut := logOutputMap[FILE]

	if isFileOut {
		parent := filepath.Dir(logFilePath)
		if _, err := os.Stat(parent); os.IsNotExist(err) {
			if err = os.MkdirAll(parent, 0644); err != nil {
				return nil, err
			}
		}
		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		if isStdout {
			logWriter := io.MultiWriter(logFile, os.Stdout)
			loggerHelper.logger = log.New(logWriter, "", 0)
		} else {
			loggerHelper.logger = log.New(logFile, "", 0)
		}
	} else {
		loggerHelper.logger = log.New(os.Stdout, "", 0)
	}

	// Set log level
	if val, ok := logLevelNumbers[loggerHelper.userLoggerSettings.Level]; ok {
		loggerHelper.programLoggerSettings.LogLevelNumber = val
	} else {
		loggerHelper.programLoggerSettings.LogLevelNumber = logLevelNumbers["WARNING"]
	}

	return loggerHelper, nil

}

func (loggerHelper *LoggerHelper) SetLogLevel(loglevel int) {
	loggerHelper.programLoggerSettings.LogLevelNumber = loglevel
}

func (loggerHelper *LoggerHelper) AddLog(logLevel int, logInfo string) {

	logLevelInfo := logLevelInfoMap[WARNING]
	if info, ok := logLevelInfoMap[logLevel]; ok {
		logLevelInfo = info
	}

	if logLevel >= loggerHelper.programLoggerSettings.LogLevelNumber {

		datetime := time.Now().Format(loggerHelper.programLoggerSettings.OutputFormat.Datetime)
		infoRune := []rune(logInfo)
		totalLen := len(infoRune)
		if totalLen > loggerHelper.programLoggerSettings.MaxInfoLength {
			infoRune = infoRune[0:loggerHelper.programLoggerSettings.MaxInfoLength]
			logInfo = fmt.Sprintf("%s ...\n(Total length: %d)", string(infoRune), totalLen)
		}
		logRecord := fmt.Sprintf(loggerHelper.programLoggerSettings.OutputFormat.LogRecord, datetime, logLevelInfo.logLevelName, logInfo)

		if loggerHelper.userLoggerSettings.UseColor {
			currentColor := color.New(logLevelInfo.logLevelColor).SprintFunc()
			logRecord = currentColor(logRecord)
		}

		loggerHelper.logger.Println(logRecord)
	}

}
