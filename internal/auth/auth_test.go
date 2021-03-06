package auth

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	conf := viper.New()
	conf.SetConfigFile("../../configs/signald.yaml")
	if err := conf.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("read config file")
	}

	err := Configure(conf.Sub("auth"))
	if err != nil {
		log.Fatal().Err(err).Msg("configure")
	}
}

func TestAuth_BasicAuth_Handler(t *testing.T) {
	okResponder := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	tests := []struct {
		desc     string
		giveUser string
		givePass string
		wantCode int
		wantBody string
	}{
		{
			desc:     "basic auth ok",
			giveUser: "testuser",
			givePass: "test",
			wantCode: http.StatusOK,
			wantBody: "OK",
		},
		{
			desc:     "basic auth wrong user",
			giveUser: "wrong.user",
			givePass: "test",
			wantCode: http.StatusUnauthorized,
			wantBody: "Unauthorized",
		},
		{
			desc:     "basic auth wrong password",
			giveUser: "testuser",
			givePass: "wrong password",
			wantCode: http.StatusUnauthorized,
			wantBody: "Unauthorized",
		},
		{
			desc:     "basic auth empty user",
			giveUser: "",
			givePass: "test",
			wantCode: http.StatusUnauthorized,
			wantBody: "Unauthorized",
		},
		{
			desc:     "basic auth empty password",
			giveUser: "testuser",
			givePass: "",
			wantCode: http.StatusUnauthorized,
			wantBody: "Unauthorized",
		},
		{
			desc:     "basic auth empty user and password",
			giveUser: "",
			givePass: "",
			wantCode: http.StatusUnauthorized,
			wantBody: "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			require.NoError(t, err)

			req.SetBasicAuth(tt.giveUser, tt.givePass)

			rr := httptest.NewRecorder()
			handler := BasicAuth(okResponder)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}

func TestAuth_BasicAuthHeader(t *testing.T) {
	want := make(http.Header)
	want.Add("Authorization", "Basic dGVzdHVzZXI6dGVzdA==")

	header := BasicAuthHeader("testuser", "test")
	assert.Equal(t, want, header)
}

func TestAuth_UserPass(t *testing.T) {
	auth := UserPass("testuser", "wrong password")
	assert.Equal(t, auth, false, "wrong password should not match")

	auth = UserPass("testuser", "test")
	assert.Equal(t, auth, true, "right password should match")
}

func TestAuth_UserAuthKey(t *testing.T) {
	key, err := UserAuthKey("unknown user", "devnet.test")
	assert.Error(t, err, "unknown user should cause an error")
	assert.Nil(t, key, "should return nil for unknown user")

	key, err = UserAuthKey("testuser", "devnet.test")
	assert.NoError(t, err, "known user should not cause an error")
	want, _ := hex.DecodeString("dcadec4f59a9793b5ebd7e278dd4f28a")
	assert.Equal(t, want, key, "key should match")
}
