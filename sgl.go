package sgl

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type LogMode int

const (
	LOG_DEBUG LogMode = iota
	LOG_INFO
	LOG_WARNING
	LOG_ERROR
	LOG_ERROR_ONCE
)

var logLevelsNames = [...]string{
	"DEBUG",
	"INFO",
	"WARNING",
	"ERROR",
	"ERROR_ONCE",
}

func (s LogMode) String() string {
	return logLevelsNames[s]
}

type logMsg struct {
	Level LogMode
	Time  time.Time
	Msg   string
}

var logChannel chan *logMsg = make(chan *logMsg, 1024)

var onceErrors map[string]bool = make(map[string]bool)

func log_worker() {
	for m := range logChannel {
		if m != nil {
			if m.Level == LOG_ERROR_ONCE {
				if _, ok := onceErrors[m.Msg]; ok {
					continue
				} else {
					onceErrors[m.Msg] = true
				}
			}

			logFile.WriteString(fmt.Sprintf("%02d-%02d-%04d | %02d:%02d:%06g | %s | %s\n",
				m.Time.Day(), m.Time.Month(), m.Time.Year(),
				m.Time.Hour(), m.Time.Minute(), float32(m.Time.Second())+float32(m.Time.Nanosecond())/(1000000.0*1000.0),
				m.Level, m.Msg))
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
					log.Fatalf("Can't open log file '%s'.\n", logFileName)
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

	logChannel <- &m
}

var logFile *os.File
var logFileName string
var minLogLevel LogMode

var maxLogSize int
var nLogs int
var currentLogIndex int

func Init(fileName string, minLevel LogMode, maxSize int, n int) {
	f, err := os.OpenFile(fileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("Can't open log file '%s'.\n", fileName)
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
	log_func(LOG_DEBUG, format, a...)
}

func Info(format string, a ...interface{}) {
	log_func(LOG_INFO, format, a...)
}

func Warning(format string, a ...interface{}) {
	log_func(LOG_WARNING, format, a...)
}

func Error(format string, a ...interface{}) {
	log_func(LOG_ERROR, format, a...)
}

func ErrorOnce(format string, a ...interface{}) {
	log_func(LOG_ERROR_ONCE, format, a...)
}
