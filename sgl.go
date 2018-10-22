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
	ShowFileLine    Flags = 1 << 0
	PrintToStdout         = 1 << 1
	BufferedLogging       = 1 << 2
)

type logMode int

const (
	DebugLevel logMode = iota
	InfoLevel
	WarningLevel
	ErrorLevel
	ErrorOnceLevel
	PanicLevel
)

var logLevelsNames = [...]string{
	"DEBUG",
	"INFO",
	"WARNING",
	"ERROR",
	"ERROR_ONCE",
}

func stringToLogLevel(ll string) logMode {
	res := DebugLevel
	for _, i := range logLevelsNames {
		if ll == i {
			return res
		}
		res++
	}
	return res
}

func (s logMode) String() string {
	return logLevelsNames[s]
}

type logMsg struct {
	Level logMode
	Time  time.Time
	Msg   string
	File  string
	Line  int
}

var logChannel chan *logMsg

var onceErrors = make(map[string]bool)

func logWriter(m *logMsg) string {
	var msg string
	if !enabledFileLine {
		msg = fmt.Sprintf("%02d-%02d-%04d | %02d:%02d:%.3f | %s | %s\n",
			m.Time.Day(), m.Time.Month(), m.Time.Year(),
			m.Time.Hour(), m.Time.Minute(), float32(m.Time.Second())+float32(m.Time.Nanosecond())/(1000000.0*1000.0),
			m.Level, m.Msg)
	} else {
		msg = fmt.Sprintf("%02d-%02d-%04d | %02d:%02d:%.3f | %s:%d | %s | %s\n",
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

	return msg
}

func logWorker() {
	for {
		m := <-logChannel
		if m != nil {
			if m.Level == ErrorOnceLevel {
				if _, ok := onceErrors[m.Msg]; ok {
					continue
				} else {
					onceErrors[m.Msg] = true
				}
			}

			logWriter(m)

			fi, err := logFile.Stat()
			if err != nil {
				continue
			}

			if nLogs != 0 && fi.Size() > int64(maxLogSize) {
				logFile.Close()
				os.Rename(logFileName, logFileName+"."+strconv.FormatUint(currentLogIndex, 10))
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

func logFunc(level logMode, format string, a ...interface{}) {
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
var minLogLevel logMode

var maxLogSize uint64
var nLogs uint64
var currentLogIndex uint64

var enabledFileLine bool
var printToStdOut bool
var bufferedLogging bool

func Init(fileName string, minLevel logMode, maxSize uint64, n uint64, flags Flags) {
	enabledFileLine = (flags & ShowFileLine) != 0
	printToStdOut = (flags & PrintToStdout) != 0
	bufferedLogging = (flags & BufferedLogging) != 0

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
	currentLogIndex = uint64(1)

	go logWorker()
}

func Deinit() {
	logChannel <- nil
}

func Debug(format string, a ...interface{}) {
	logFunc(DebugLevel, format, a...)
}

func Info(format string, a ...interface{}) {
	logFunc(InfoLevel, format, a...)
}

func Warning(format string, a ...interface{}) {
	logFunc(WarningLevel, format, a...)
}

func Error(format string, a ...interface{}) {
	logFunc(ErrorLevel, format, a...)
}

func ErrorOnce(format string, a ...interface{}) {
	logFunc(ErrorOnceLevel, format, a...)
}

func Panic(format string, a ...interface{}) {
	m := logMsg{Level: PanicLevel, Time: time.Now(), Msg: fmt.Sprintf(format, a...)}
	if enabledFileLine {
		_, f, l, _ := runtime.Caller(2)
		m.File = f
		m.Line = l
	}
	logWriter(&m)
	panic(logWriter(&m))
}
