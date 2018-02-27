package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/pivotal-golang/lager"
)

const (
	DEFAULT_PORT     = "8080"
	CF_FORWARDED_URL = "X-Cf-Forwarded-Url"
	DEFAULT_USERNAME = "admin"
	DEFAULT_PASSWORD = "letmein"
)

var c *config

type config struct {
	username string
	password string
	port     string
}

func main() {

	logger := lager.NewLogger("p-basic-auth-router")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))
	c = configFromEnvironmentVariables()

	http.Handle("/", wrapper(newProxy()))
	logger.Fatal("http-listen", http.ListenAndServe(":"+getPort(), nil))
}

func configFromEnvironmentVariables() *config {
	conf := &config{
		username: getEnv("basicauth_username", DEFAULT_USERNAME),
		password: getEnv("basicauth_password", DEFAULT_PASSWORD),
		port:     getPort(),
	}
	return conf
}

func newProxy() http.Handler {
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			forwardedURL := req.Header.Get(CF_FORWARDED_URL)
			url, err := url.Parse(forwardedURL)
			if err != nil {
				log.Fatalln(err.Error())
			}

			req.URL = url
			req.Host = url.Host
			logger := lager.NewLogger("proxy")
			logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

			logger.Debug("X-Cf-Forwarded-URL", lager.Data{
				"X-Cf-Forwarded-Url": req.Header.Get(CF_FORWARDED_URL),
			})

			logger.Debug("X-CF-Proxy-Signature", lager.Data{
				"X-CF-Proxy-Signature": req.Header.Get("X-CF-Proxy-Signature"),
			})

			logger.Debug("X-CF-Proxy-Metadata", lager.Data{
				"X-CF-Proxy-Metadata": req.Header.Get("X-CF-Proxy-Metadata"),
			})

		},
	}
	return proxy
}

func getPort() string {
	var port string
	if port = os.Getenv("PORT"); len(port) == 0 {
		port = DEFAULT_PORT
	}
	return port
}

func getEnv(env string, defaultValue string) string {
	var (
		v string
	)
	if v = os.Getenv(env); len(v) == 0 {
		log.Printf("using default value for %v", env)
		return defaultValue
	}

	return env
}
func wrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (len(c.username) > 0) && (len(c.password) > 0) && !auth(r, c.username, c.password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="REALM"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func auth(r *http.Request, user, pass string) bool {
	if username, password, ok := r.BasicAuth(); ok {
		return username == user && password == pass
	}
	return false
}
