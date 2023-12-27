package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type Settings struct {
	Path       string `yaml:"path"`
	Name       string `yaml:"name"`
	Ext        string `yaml:"ext"`
	TimeFormat string `yanl:"time_format"`
}

type logLevel int

const (
	DEBUG logLevel = iota
	INFO
	WARNING
	ERROR
	FATAL
)

const (
	flags              = log.LstdFlags
	defaultCallerDepth = 2
	bufferSize         = 1e5
)

type logEntry struct {
	msg   string
	level logLevel
}

var levelFlags = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

type Logger struct {
	logFile   *os.File
	logger    *log.Logger
	entryChan chan *logEntry
	entryPool *sync.Pool
}

var DefaultLogger = NewStdoutLogger()

// NewStdoutLogger 创建一个 logger，日志输出到 stdout
func NewStdoutLogger() *Logger {
	logger := &Logger{
		logFile:   nil,
		logger:    log.New(os.Stdout, "", flags),
		entryChan: make(chan *logEntry, bufferSize),
		entryPool: &sync.Pool{
			New: func() interface{} {
				return &logEntry{}
			},
		},
	}
	go func() {
		for e := range logger.entryChan {
			_ = logger.logger.Output(0, e.msg)
			logger.entryPool.Put(e)
		}
	}()
	return logger
}

// NewFileLogger 创建一个 logger，日志输出到 stdout 和日志文件
func NewFileLogger(settings *Settings) (*Logger, error) {
	fileName := fmt.Sprintf("%s-%s.%s", settings.Name, time.Now().Format(settings.TimeFormat), settings.Ext)
	logFile, err := MustOpen(fileName, settings.Path)
	if err != nil {
		return nil, fmt.Errorf("logging.Join err: %s", err)
	}
	logger := &Logger{
		logFile:   logFile,
		logger:    log.New(io.MultiWriter(os.Stdout, logFile), "", flags),
		entryChan: make(chan *logEntry, bufferSize),
		entryPool: &sync.Pool{
			New: func() interface{} {
				return &logEntry{}
			},
		},
	}
	go func() {
		for e := range logger.entryChan {
			fileName := fmt.Sprintf("%s-%s.%s", settings.Name, time.Now().Format(settings.TimeFormat), settings.Ext)
			if path.Join(settings.Path, fileName) != logger.logFile.Name() {
				logFile, err := MustOpen(fileName, settings.Path)
				if err != nil {
					panic(fmt.Sprintf("Open log %s failed: %s", fileName, err.Error()))
				}
				logger.logFile = logFile
				logger.logger = log.New(io.MultiWriter(os.Stdout, logFile), "", flags)
			}
			_ = logger.logger.Output(0, e.msg)
			logger.entryPool.Put(e)
		}
	}()
	return logger, nil
}

func Setup(settings *Settings) {
	logger, err := NewFileLogger(settings)
	if err != nil {
		panic(err)
	}
	DefaultLogger = logger
}

func (logger *Logger) Output(level logLevel, callerDepth int, msg string) {
	var formattedMsg string
	_, file, line, ok := runtime.Caller(callerDepth)
	if ok {
		formattedMsg = fmt.Sprintf("[%s][%s:%d] %s", levelFlags[level], filepath.Base(file), line, msg)
	} else {
		formattedMsg = fmt.Sprintf("[%s] %s", levelFlags[level], msg)
	}
	entry := logger.entryPool.Get().(*logEntry)
	entry.msg = formattedMsg
	entry.level = level
	logger.entryChan <- entry
}

func Debug(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	DefaultLogger.Output(DEBUG, defaultCallerDepth, msg)
}

func Debugf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	DefaultLogger.Output(DEBUG, defaultCallerDepth, msg)
}

func Info(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	DefaultLogger.Output(INFO, defaultCallerDepth, msg)
}

func Infof(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	DefaultLogger.Output(INFO, defaultCallerDepth, msg)
}

func Warn(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	DefaultLogger.Output(WARNING, defaultCallerDepth, msg)
}

func Error(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	DefaultLogger.Output(ERROR, defaultCallerDepth, msg)
}

func Errorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	DefaultLogger.Output(ERROR, defaultCallerDepth, msg)
}

func Fatal(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	DefaultLogger.Output(FATAL, defaultCallerDepth, msg)
}
