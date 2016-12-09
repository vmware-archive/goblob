package boshlogtest

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega/gbytes"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type TestLogger struct {
	boshlog.Logger
	*TestSink
}

type TestSink struct {
	boshlog.Sink
	buffer *gbytes.Buffer
}

func NewTestLogger(component string) *TestLogger {
	logger := boshlog.NewLogger(component)

	testSink := NewTestSink()
	logger.RegisterSink(testSink)
	logger.RegisterSink(boshlog.NewWriterSink(ginkgo.GinkgoWriter, boshlog.DEBUG))

	return &TestLogger{logger, testSink}
}

func NewTestSink() *TestSink {
	buffer := gbytes.NewBuffer()

	return &TestSink{
		Sink:   boshlog.NewWriterSink(buffer, boshlog.DEBUG),
		buffer: buffer,
	}
}

func (s *TestSink) Buffer() *gbytes.Buffer {
	return s.buffer
}

func (s *TestSink) Logs() []boshlog.LogFormat {
	logs := []boshlog.LogFormat{}

	decoder := json.NewDecoder(bytes.NewBuffer(s.buffer.Contents()))
	for {
		var log boshlog.LogFormat
		if err := decoder.Decode(&log); err == io.EOF {
			return logs
		} else if err != nil {
			panic(err)
		}
		logs = append(logs, log)
	}

	return logs
}

func (s *TestSink) LogMessages() []string {
	logs := s.Logs()
	messages := make([]string, 0, len(logs))
	for _, log := range logs {
		messages = append(messages, log.Message)
	}
	return messages
}
