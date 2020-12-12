package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	conf "github.com/spf13/viper"
)

func init() {
	log.SetLevel(log.ErrorLevel)
}

// UserPass implements basic username / password verification.
func UserPass(user string, pass string) bool {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s+%s", user, pass)))
	str := fmt.Sprintf("%x", sum)
	truth := conf.GetString(fmt.Sprintf("users.%s", user))
	if str == truth {
		log.Trace("authenticated user: ", user)
		return true
	} else {
		log.Warn("authentication failure: ", user)
		return false
	}
}

// BasicAuth provides an authentication wrapper for http.HandlerFunc.
func BasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if ok && UserPass(user, pass) {
			next.ServeHTTP(w, r)
		} else {
			code := http.StatusUnauthorized
			http.Error(w, http.StatusText(code), code)
			return
		}
	})
}

// BasicAuthHeader returns an Authorization http.Header for BasicAuth.
func BasicAuthHeader(user, pass string) http.Header {
	cred := user + ":" + pass
	auth := base64.StdEncoding.EncodeToString([]byte(cred))
	header := make(http.Header)
	header.Add("Authorization", "Basic "+auth)
	return header
}
