package Inits

import (
	"bytes"
	"email/global"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)

type LogFormatter struct{}

func (t *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = gray
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	//自定义日期格式
	timeStamp := entry.Time.Format("2006-01-02 15:04:05")
	if entry.HasCaller() {
		//funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line)
		//fmt.Fprintf(b, "[ %s ] \x1b[%dm[%s]\x1b[0m %s %s %s \n", timeStamp, levelColor, entry.Level, fileVal, funcVal, entry.Message)
		fmt.Fprintf(b, "[ %s ] [ %s ] \x1b[%dm[%s]\x1b[0m %s | \x1b[%dm%s\x1b[0m \n", global.Config.Logger.Prefix, timeStamp, levelColor, entry.Level, fileVal, levelColor, entry.Message)
	} else {
		fmt.Fprintf(b, "[ %s ] [ %s ] \x1b[%dm[%s]\x1b[0m %s \n", global.Config.Logger.Prefix, timeStamp, levelColor, entry.Level, entry.Message)
	}

	return b.Bytes(), nil
}

func InitLogger() {
	qlog := logrus.New()
	qlog.SetOutput(os.Stdout)
	qlog.SetReportCaller(global.Config.Logger.ShowLine)
	qlog.SetFormatter(&LogFormatter{})
	level, err := logrus.ParseLevel(global.Config.Logger.Level)
	if err != nil {
		level = logrus.DebugLevel
	}
	qlog.SetLevel(level)
	InitDefaultLogger()
	global.Log = qlog
	global.Log.Info("配置文件读取成功...")
	global.Log.Info("日志初始化成功...")
}

// 全局Log
func InitDefaultLogger() {
	logrus.SetOutput(os.Stdout)
	logrus.SetReportCaller(global.Config.Logger.ShowLine)
	logrus.SetFormatter(&LogFormatter{})
	level, err := logrus.ParseLevel(global.Config.Logger.Level)
	if err != nil {
		level = logrus.DebugLevel
	}
	logrus.SetLevel(level)
}
