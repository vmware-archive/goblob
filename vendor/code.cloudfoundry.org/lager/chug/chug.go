package chug

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Entry struct {
	Isboshlog bool
	Raw     []byte
	Log     LogEntry
}

type LogEntry struct {
	Timestamp time.Time
	LogLevel  boshlog.LogLevel

	Source  string
	Message string
	Session string

	Error error
	Trace string

	Data boshlog.Data
}

func Chug(reader io.Reader, out chan<- Entry) {
	scanner := bufio.NewReader(reader)
	for {
		line, err := scanner.ReadBytes('\n')
		if line != nil {
			out <- entry(bytes.TrimSuffix(line, []byte{'\n'}))
		}
		if err != nil {
			break
		}
	}
	close(out)
}

func entry(raw []byte) (entry Entry) {
	copiedBytes := make([]byte, len(raw))
	copy(copiedBytes, raw)
	entry = Entry{
		Isboshlog: false,
		Raw:     copiedBytes,
	}

	rawString := string(raw)
	idx := strings.Index(rawString, "{")
	if idx == -1 {
		return
	}

	var boshlogLog boshlog.LogFormat
	decoder := json.NewDecoder(strings.NewReader(rawString[idx:]))
	err := decoder.Decode(&boshlogLog)
	if err != nil {
		return
	}

	entry.Log, entry.Isboshlog = convertboshlogLog(boshlogLog)

	return
}

func convertboshlogLog(boshlogLog boshlog.LogFormat) (LogEntry, bool) {
	timestamp, err := strconv.ParseFloat(boshlogLog.Timestamp, 64)

	if err != nil {
		return LogEntry{}, false
	}

	data := boshlogLog.Data

	var logErr error
	if boshlogLog.LogLevel == boshlog.ERROR || boshlogLog.LogLevel == boshlog.FATAL {
		dataErr, ok := boshlogLog.Data["error"]
		if ok {
			errorString, ok := dataErr.(string)
			if !ok {
				return LogEntry{}, false
			}
			logErr = errors.New(errorString)
			delete(boshlogLog.Data, "error")
		}
	}

	var logTrace string
	dataTrace, ok := boshlogLog.Data["trace"]
	if ok {
		logTrace, ok = dataTrace.(string)
		if !ok {
			return LogEntry{}, false
		}
		delete(boshlogLog.Data, "trace")
	}

	var logSession string
	dataSession, ok := boshlogLog.Data["session"]
	if ok {
		logSession, ok = dataSession.(string)
		if !ok {
			return LogEntry{}, false
		}
		delete(boshlogLog.Data, "session")
	}

	return LogEntry{
		Timestamp: time.Unix(0, int64(timestamp*1e9)),
		LogLevel:  boshlogLog.LogLevel,
		Source:    boshlogLog.Source,
		Message:   boshlogLog.Message,
		Session:   logSession,

		Error: logErr,
		Trace: logTrace,

		Data: data,
	}, true
}
