package base

import (
	"auto-portal-auth/component/utils"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"
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

type LogLevelHelper struct {
	LogLevelName        string
	LogLevelColorString string
}

var (
	LogLevelNumbers = map[string]int{
		"DEBUG":   DEBUG,
		"INFO":    INFO,
		"WARNING": WARNING,
		"ERROR":   ERROR,
		"FATAL":   FATAL,
		"MUTE":    MUTE,
	}
	LogLevelInfo = map[int]LogLevelHelper{
		DEBUG:   {"DEBUG", "%s"},
		INFO:    {"INFO", "\033[1;92m%s\033[0m"},
		WARNING: {"WARNING", "\033[1;93m%s\033[0m"},
		ERROR:   {"ERROR", "\033[1;91m%s\033[0m"},
		FATAL:   {"FATAL", "\033[1;95m%s\033[0m"},
		MUTE:    {},
	}
)

type LoggerHelper struct {
	Logger            *log.Logger
	LogLevel          int
	LogRecordFormat   string
	LogDatetimeFormat string
	UseColor          bool
}

func InitLoggerHelper(configHelper *ConfigHelper) (*LoggerHelper, error) {

	if configHelper == nil {
		err := errors.New("configHelper is invalid")
		return nil, err
	}

	loggerHelper := &LoggerHelper{}

	// Set output type
	logFilePath := configHelper.UserSettings.Log.LogFile
	_, logOutputMap := utils.RemoveDuplicateStrings(configHelper.UserSettings.Log.LogOutput)
	_, isStdout := logOutputMap[STDOUT]
	_, isFileOut := logOutputMap[FILE]

	if isFileOut {
		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		if isStdout {
			logWriter := io.MultiWriter(logFile, os.Stdout)
			loggerHelper.Logger = log.New(logWriter, "", 0)
		} else {
			loggerHelper.Logger = log.New(logFile, "", 0)
		}
	} else {
		loggerHelper.Logger = log.New(os.Stdout, "", 0)
	}

	// Set log level
	if val, ok := LogLevelNumbers[configHelper.UserSettings.Log.LogLevel]; ok {
		loggerHelper.LogLevel = val
	} else {
		loggerHelper.LogLevel = LogLevelNumbers["WARNING"]
	}

	// Set log record format
	loggerHelper.LogRecordFormat = configHelper.ProgramSettings.Log.LogRecord
	loggerHelper.LogDatetimeFormat = configHelper.ProgramSettings.Log.Datetime
	loggerHelper.UseColor = configHelper.UserSettings.Log.UseColor

	return loggerHelper, nil

}

func (loggerHelper *LoggerHelper) AddLog(logLevel int, logInfo string) {

	level := LogLevelInfo[WARNING].LogLevelName
	color := LogLevelInfo[DEBUG].LogLevelColorString

	if info, ok := LogLevelInfo[logLevel]; ok {
		level = info.LogLevelName
		if loggerHelper.UseColor {
			color = info.LogLevelColorString
		}
	}

	if logLevel >= loggerHelper.LogLevel {
		datetime := time.Now().Format(loggerHelper.LogDatetimeFormat)
		logRecord := fmt.Sprintf(loggerHelper.LogRecordFormat, datetime, level, logInfo)
		loggerHelper.Logger.Printf(color, logRecord)
	}

}
