package managers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"
)

const tokenRegexp = "bearer [\\w\\.-]+"

// RedactedValue is the text that is displayed for redacted content. (eg
// authorization tokens, passwords, etc.)
const RedactedValue = "[PRIVATE DATA HIDDEN]"

func RedactHeaders(header http.Header) http.Header {
	redactedHeaders := make(http.Header)
	for key, value := range header {
		if key == "Authorization" {
			redactedHeaders[key] = []string{RedactedValue}
		} else {
			redactedHeaders[key] = value
		}
	}
	return redactedHeaders
}

const tokenEndpoint = "token_endpoint"

var keysToSanitize = regexp.MustCompile("(?i)token|password")
var sanitizeURIParams = regexp.MustCompile(`([&?]password)=[A-Za-z0-9\-._~!$'()*+,;=:@/?]*`)
var sanitizeURLPassword = regexp.MustCompile(`([\d\w]+):\/\/([^:]+):(?:[^@]+)@`)

func SanitizeJSON(raw []byte) ([]byte, error) {
	var result interface{}
	decoder := json.NewDecoder(bytes.NewBuffer(raw))
	decoder.UseNumber()
	err := decoder.Decode(&result)
	if err != nil {
		return nil, err
	}

	sanitized := iterateAndRedact(result)

	buff := new(bytes.Buffer)
	encoder := json.NewEncoder(buff)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(sanitized)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func iterateAndRedact(blob interface{}) interface{} {
	switch v := blob.(type) {
	case string:
		return sanitizeURL(v)
	case []interface{}:
		var list []interface{}
		for _, val := range v {
			list = append(list, iterateAndRedact(val))
		}

		return list
	case map[string]interface{}:
		for key, value := range v {
			if keysToSanitize.MatchString(key) && key != tokenEndpoint {
				v[key] = RedactedValue
			} else {
				v[key] = iterateAndRedact(value)
			}
		}
		return v
	}
	return blob
}

func sanitizeURL(rawURL string) string {
	sanitized := sanitizeURLPassword.ReplaceAllString(rawURL, fmt.Sprintf("$1://$2:%s@", RedactedValue))
	sanitized = sanitizeURIParams.ReplaceAllString(sanitized, fmt.Sprintf("$1=%s", RedactedValue))
	return sanitized
}

type RequestLogger struct {
	dumpSanitizer *regexp.Regexp
}

func NewRequestLogger() *RequestLogger {
	return &RequestLogger{
		dumpSanitizer: regexp.MustCompile(tokenRegexp),
	}
}

func (display *RequestLogger) DisplayBody([]byte) error {
	log.Print(fmt.Sprintf("%s\n", RedactedValue))
	return nil
}

func (display *RequestLogger) DisplayDump(dump string) error {
	sanitized := display.dumpSanitizer.ReplaceAllString(dump, RedactedValue)
	log.Print(fmt.Sprintf("%s\n", sanitized))
	return nil
}

func (display *RequestLogger) DisplayHeader(name string, value string) error {
	log.Print(fmt.Sprintf("%s: %s\n", name, value))
	return nil
}

func (display *RequestLogger) DisplayHost(name string) error {
	log.Print(fmt.Sprintf("host: %s\n", name))
	return nil
}

func (display *RequestLogger) DisplayJSONBody(body []byte) error {
	if len(body) == 0 {
		return nil
	}

	sanitized, err := SanitizeJSON(body)
	if err != nil {
		log.Print(fmt.Sprintf("%s\n", string(body)))
		return nil
	}

	log.Print(fmt.Sprintf("%s\n", string(sanitized)))

	return nil
}

func (display *RequestLogger) DisplayMessage(msg string) error {
	log.Print(fmt.Sprintf("%s\n", msg))
	return nil
}

func (display *RequestLogger) DisplayRequestHeader(method string, uri string, httpProtocol string) error {
	log.Print(fmt.Sprintf("%s %s %s\n", method, uri, httpProtocol))
	return nil
}

func (display *RequestLogger) DisplayResponseHeader(httpProtocol string, status string) error {
	log.Print(fmt.Sprintf("%s %s\n", httpProtocol, status))
	return nil
}

func (display *RequestLogger) DisplayType(name string, requestDate time.Time) error {
	text := fmt.Sprintf("%s: [%s]", name, requestDate.Format(time.RFC3339))
	log.Println(text)
	return nil
}

func (display *RequestLogger) HandleInternalError(err error) {
	log.Println(err.Error())
}

func (display *RequestLogger) Start() error {
	return nil
}

func (display *RequestLogger) Stop() error {
	return nil
}
