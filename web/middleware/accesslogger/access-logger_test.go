package accesslogger

import (
	"fmt"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type OkHandlerStruct struct{}

func (O OkHandlerStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if r.Body != nil {
		body, _ := ioutil.ReadAll(r.Body)
		_, _ = w.Write(body)
	}
}

func TestApacheCombined(t *testing.T) {
	tests := []struct {
		name    string
		request http.Request
		want    string
	}{
		{
			name: "happy path",
			request: http.Request{
				Method:     http.MethodGet,
				RemoteAddr: "1.2.3.4:1234",
				URL: &url.URL{
					Scheme: "HTTP/1.1",
					Path:   "/foobar",
				},
				Header: http.Header{
					"User-Agent":    []string{"this go test"},
					"Referrer":      []string{"http://nowhere"},
					"Authorization": []string{"Basic bXktdXNlcjpmb29iYXIK"}, // my-user:foobar
				},
				Body: io.NopCloser(strings.NewReader("Hello, world!")),
			},
			// <remote host> <ident> <user> [<time>] "<method> <path> <protocol>" <status code> <size> "<referrer>" "<user agent>"
			want: "1.2.3.4 - my-user [%s] \"GET /foobar HTTP/1.1\" 200 13 \"http://nowhere\" \"this go test\"",
		},
		{
			name: "happy path - no size",
			request: http.Request{
				Method:     http.MethodGet,
				RemoteAddr: "1.2.3.4:1234",
				URL: &url.URL{
					Scheme: "HTTP/1.1",
					Path:   "/foobar",
				},
				Header: http.Header{
					"User-Agent":    []string{"this go test"},
					"Referrer":      []string{"http://nowhere"},
					"Authorization": []string{"Basic bXktdXNlcjpmb29iYXIK"}, // my-user:foobar
				},
			},
			// <remote host> <ident> <user> [<time>] "<method> <path> <protocol>" <status code> <size> "<referrer>" "<user agent>"
			want: "1.2.3.4 - my-user [%s] \"GET /foobar HTTP/1.1\" 200 - \"http://nowhere\" \"this go test\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := logrusTest.NewGlobal()
			w := httptest.NewRecorder()
			ApacheCombined(OkHandlerStruct{}).ServeHTTP(w, &tt.request)
			assert.Equal(t, fmt.Sprintf(tt.want, time.Now().Format("02/Jan/2006 15:04:05 -0700")), hook.LastEntry().Message)
		})
	}
}
