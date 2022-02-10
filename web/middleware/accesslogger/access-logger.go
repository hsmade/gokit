package accesslogger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"net/http"
	"strings"
	"time"
)

// CombinedFormatAccessLoggerMiddleware apache combined log format
func CombinedFormatAccessLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newW := negroni.NewResponseWriter(w)
		next.ServeHTTP(newW, r)

		user, _, ok := r.BasicAuth()
		if !ok {
			user = "-"
		}

		size := "-"
		if newW.Size() > 0 {
			size = fmt.Sprintf("%d", newW.Size())
		}

		logrus.Infof("%s - %s [%v] \"%s %s %s\" %d %s \"%s\" \"%s\"",
			strings.Split(r.RemoteAddr, ":")[0],
			user,
			time.Now().Format("02/Jan/2006 15:04:05 -0700"),
			r.Method,
			r.URL.Path,
			r.URL.Scheme,
			newW.Status(),
			size,
			r.Header.Get("Referrer"),
			r.Header.Get("User-Agent"),
		)
	})
}
