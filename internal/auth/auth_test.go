package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	conf "github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicAuthHandler(t *testing.T) {
	conf.Set("users.testuser",
		"09d9623a149a4a0c043befcb448c9c3324be973230188ba412c008a2929f31d0")

	okResponder := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	}

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
			handler := http.HandlerFunc(BasicAuth(okResponder))
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}

func TestBasicAuthHeader(t *testing.T) {
	want := make(http.Header)
	want.Add("Authorization", "Basic dGVzdHVzZXI6dGVzdA==")

	header := BasicAuthHeader("testuser", "test")
	assert.Equal(t, want, header)
}

func TestUserPass(t *testing.T) {
	conf.Set("users.testuser",
		"09d9623a149a4a0c043befcb448c9c3324be973230188ba412c008a2929f31d0")

	auth := UserPass("testuser", "wrong password")
	assert.Equal(t, auth, false, "wrong password should not match")

	auth = UserPass("testuser", "test")
	assert.Equal(t, auth, true, "right password should match")
}
