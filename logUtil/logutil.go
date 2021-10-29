package logUtil

import (
	"fmt"
	"github.com/xukgo/gsaber/utils/fileUtil"
	"github.com/xukgo/log4z"
	"go.uber.org/zap"
)

var commonLogger *zap.Logger

func InitLogger() {
	confPath := fileUtil.GetAbsUrl("conf/log4z.xml")
	loggerMap := log4z.InitLogger(confPath,
		log4z.WithTimeKey("timestamp"), log4z.WithTimeFormat("2006-01-02 15:04:05.999"))
	commonLogger = getLoggerOrConsole(loggerMap, "Common")
}
func getLoggerOrConsole(dict map[string]*zap.Logger, key string) *zap.Logger {
	logger, ok := dict[key]
	if ok {
		fmt.Printf("info: get logger %s success\r\n", key)
	} else {
		fmt.Printf("warnning: log4z get logger (%s) failed\r\n", key)
		fmt.Printf("warnning: now set logger %s to default console logger\r\n", key)
		logger = log4z.GetConsoleLogger()
	}
	return logger
}

func Common() *zap.Logger {
	return commonLogger
}