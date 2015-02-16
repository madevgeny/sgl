package sgl

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"
)

type Flags int

const (
	SHOW_FILE_LINE   Flags = 1 << 0
	PRINT_TO_STDOUT        = 1 << 1
	BUFFERED_LOGGING       = 1 << 2
)

type LogMode int

const (
	DEBUG LogMode = iota
	INFO
	WARNING
	ERROR
	ERROR_ONCE
)

var logLevelsNames = [...]string{
	"DEBUG",
	"INFO",
	"WARNING",
	"ERROR",
	"ERROR_ONCE",
}

func StringToLogLevel(ll string) LogMode {
	res := DEBUG
	for _, i := range logLevelsNames {
		if ll == i {
			return res
		}
		res++
	}
	return res
}

func (s LogMode) String() string {
	return logLevelsNames[s]
}

type logMsg struct {
	Level LogMode
	Time  time.Time
	Msg   string
	File  string
	Line  int
}

var logChannel chan *logMsg

var onceErrors map[string]bool = make(map[string]bool)

func log_worker() {
	for {
		m := <-logChannel
		if m != nil {
			if m.Level == ERROR_ONCE {
				if _, ok := onceErrors[m.Msg]; ok {
					continue
				} else {
					onceErrors[m.Msg] = true
				}
			}

			var msg string
			if !enabledFileLine {
				msg = fmt.Sprintf("%02d-%02d-%04d | %02d:%02d:%06g | %s | %s\n",
					m.Time.Day(), m.Time.Month(), m.Time.Year(),
					m.Time.Hour(), m.Time.Minute(), float32(m.Time.Second())+float32(m.Time.Nanosecond())/(1000000.0*1000.0),
					m.Level, m.Msg)
			} else {
				msg = fmt.Sprintf("%02d-%02d-%04d | %02d:%02d:%06g | %s:%d | %s | %s\n",
					m.Time.Day(), m.Time.Month(), m.Time.Year(),
					m.Time.Hour(), m.Time.Minute(), float32(m.Time.Second())+float32(m.Time.Nanosecond())/(1000000.0*1000.0),
					path.Base(m.File), m.Line,
					m.Level, m.Msg)
			}
			if printToStdOut {
				fmt.Print(msg)
			}
			logFile.WriteString(msg)
			logFile.Sync()

			fi, err := logFile.Stat()
			if err != nil {
				continue
			}

			if fi.Size() > int64(maxLogSize) {
				logFile.Close()
				os.Rename(logFileName, logFileName+"."+strconv.Itoa(currentLogIndex))
				currentLogIndex++
				if currentLogIndex > nLogs {
					currentLogIndex = 1
				}

				logFile, err = os.OpenFile(logFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
				if err != nil {
					panic(fmt.Sprintf("Can't open log file '%s'.\n", logFileName))
				}

				onceErrors = make(map[string]bool)
			}
		} else {
			logFile.Close()
		}
	}
}

func log_func(level LogMode, format string, a ...interface{}) {
	if level < minLogLevel {
		return
	}

	m := logMsg{Level: level, Time: time.Now(), Msg: fmt.Sprintf(format, a...)}
	if enabledFileLine {
		_, f, l, _ := runtime.Caller(2)
		m.File = f
		m.Line = l
	}

	logChannel <- &m
}

var logFile *os.File
var logFileName string
var minLogLevel LogMode

var maxLogSize int
var nLogs int
var currentLogIndex int

var enabledFileLine bool
var printToStdOut bool
var bufferedLogging bool

func Init(fileName string, minLevel LogMode, maxSize int, n int, flags Flags) {
	enabledFileLine = (flags & SHOW_FILE_LINE) != 0
	printToStdOut = (flags & PRINT_TO_STDOUT) != 0
	bufferedLogging = (flags & BUFFERED_LOGGING) != 0

	if !bufferedLogging {
		logChannel = make(chan *logMsg, 0)
	} else {
		logChannel = make(chan *logMsg, 1024)
	}

	f, err := os.OpenFile(fileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(fmt.Sprintf("Can't open log file '%s'.\n", fileName))
	}
	logFile = f

	logFileName = fileName
	minLogLevel = minLevel
	maxLogSize = maxSize
	nLogs = n
	currentLogIndex = 1

	go log_worker()
}

func Deinit() {
	logChannel <- nil
}

func Debug(format string, a ...interface{}) {
	log_func(DEBUG, format, a...)
}

func Info(format string, a ...interface{}) {
	log_func(INFO, format, a...)
}

func Warning(format string, a ...interface{}) {
	log_func(WARNING, format, a...)
}

func Error(format string, a ...interface{}) {
	log_func(ERROR, format, a...)
}

func ErrorOnce(format string, a ...interface{}) {
	log_func(ERROR_ONCE, format, a...)
}
