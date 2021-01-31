package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/hlog"
	"github.com/spf13/viper"
)

type User struct {
	Name string
	Hash string
	Key  string
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
		return false
	}
	if str != u.Hash {
		return false
	}
	return true
}

// BasicAuth provides an authentication wrapper for http.HandlerFunc.
func BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if ok && UserPass(user, pass) {
			hlog.FromRequest(r).Info().
				Str("user", user).
				Msg("user authorized")
			next.ServeHTTP(w, r)
		} else {
			hlog.FromRequest(r).Warn().
				Str("user", user).
				Msg("authorization failed")
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

// UserAuthKey returns auth keys in the format used by pion/turn
// TODO: implement handling for multiple realms
func UserAuthKey(user string, realm string) ([]byte, error) {
	u, ok := userMap[user]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}

	data, err := hex.DecodeString(u.Key)
	if err != nil {
		return nil, err
	}
	return data, nil
}
