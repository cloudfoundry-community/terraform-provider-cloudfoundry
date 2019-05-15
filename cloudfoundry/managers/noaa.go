package managers

import (
	"crypto/tls"
	"fmt"
	noaaconsumer "github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"net/http"
	"strings"
	"time"
)

const LogTimestampFormat = "2006-01-02T15:04:05.00-0700"

type NOAATokenStore interface {
	AccessToken() string
}

type NOAAClient struct {
	consumer    *noaaconsumer.Consumer
	store       NOAATokenStore
	maxMessages int
}

func NewNOAAClient(trafficControllerUrl string, skipSslValidation bool, store NOAATokenStore, maxMessages int) *NOAAClient {
	consumer := noaaconsumer.New(trafficControllerUrl, &tls.Config{
		InsecureSkipVerify: skipSslValidation,
	}, http.ProxyFromEnvironment)
	return &NOAAClient{
		consumer:    consumer,
		store:       store,
		maxMessages: maxMessages,
	}
}

func (c NOAAClient) RecentLogs(appGUID string) (string, error) {
	logMsgs, err := c.consumer.RecentLogs(appGUID, c.store.AccessToken())
	if err != nil {
		return "", err
	}
	maxLen := c.maxMessages
	if maxLen < 0 {
		maxLen = len(logMsgs)
	}
	if maxLen-1 < 0 {
		return "", nil
	}
	logs := ""
	for i := maxLen - 1; i >= 0; i-- {
		logMsg := logMsgs[i]
		t := time.Unix(0, logMsg.GetTimestamp()).In(time.Local).Format(LogTimestampFormat)
		typeMessage := "OUT"
		if logMsg.GetMessageType() != events.LogMessage_OUT {
			typeMessage = "ERR"
		}
		header := fmt.Sprintf("%s [%s/%s] %s ",
			t,
			logMsg.GetSourceType(),
			logMsg.GetSourceInstance(),
			typeMessage,
		)
		message := string(logMsg.GetMessage())
		for _, line := range strings.Split(message, "\n") {
			logs += fmt.Sprintf("%s%s\n", header, strings.TrimRight(line, "\r\n"))
		}
	}
	return logs, nil
}
