package sgl

import (
	"fmt"
	"log"
	"os"
)

type LogMode int

const (
	LOG_DEBUG LogMode = iota
	LOG_INFO
	LOG_WARNING
	LOG_ERROR
	LOG_ERROR_ONCE
)

var LogLevelsNames = [...]string{
	"DEBUG:",
	"INFO:",
	"WARNING:",
	"ERROR:",
	"ERROR_ONCE:",
}

func (s LogMode) String() string {
	return LogLevelsNames[s]
}

type LogMsg struct {
	Level LogMode

	Msg string
}

var logChannel chan *LogMsg = make(chan *LogMsg, 1024)

var onceErrors map[string]bool = make(map[string]bool)

func StringToLogLevel(ll string) LogMode {
	res := LOG_DEBUG
	for _, i := range LogLevelsNames {
		if ll == i[:len(i)-1] {
			return res
		}
		res++
	}
	return res
}

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
			log.Println(m.Level, m.Msg)
		} else {
			logFile.Close()
		}
	}
}

func log_func(level LogMode, format string, a ...interface{}) {
	if level < minLogLevel {
		return
	}

	m := LogMsg{Level: level}

	m.Msg = fmt.Sprintf(format, a...)

	logChannel <- &m
}

var logFile os.File
var minLogLevel LogMode

func Init(logFileName string, flags int, minLevel LogMode) {
	if logFileName != "" {
		logFile, err := os.OpenFile(logFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Printf("Can't open log file '%s'.\n", logFileName)
		} else {
			log.SetOutput(logFile)
		}
	}

	log.SetFlags(flags)

	minLogLevel = minLevel

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
