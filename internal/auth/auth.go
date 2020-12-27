package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type User struct {
	Name string
	Hash string
}

// TODO: use external auth provider
var userMap map[string]User

// Configure sets up the auth module based on conf.
func Configure(conf *viper.Viper) error {
	userMap = make(map[string]User)

	var userList []User
	if err := conf.UnmarshalKey("users", &userList); err != nil {
		return fmt.Errorf("unmarshal user list: %v", err)
	}
	for _, u := range userList {
		userMap[u.Name] = u
	}
	return nil
}

// UserPass implements basic username / password verification.
func UserPass(user string, pass string) bool {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s+%s", user, pass)))
	str := fmt.Sprintf("%x", sum)
	u, ok := userMap[user]
	if !ok {
		log.Warn("authentication failure: ", user)
		return false
	}
	if str != u.Hash {
		log.Warn("authentication failure: ", user)
		return false
	}
	log.Trace("authenticated user: ", user)
	return true
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
