package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	conf "github.com/spf13/viper"
)

func init() {
	log.SetLevel(log.ErrorLevel)
}

func TestBasicAuthHandler(t *testing.T) {
	conf.Set("users.testuser",
		"09d9623a149a4a0c043befcb448c9c3324be973230188ba412c008a2929f31d0")

	okResponder := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	}

	cases := []struct {
		desc       string
		user       string
		pass       string
		exp_status int
		exp_body   string
	}{
		{
			desc:       "basic auth ok",
			user:       "testuser",
			pass:       "test",
			exp_status: http.StatusOK,
			exp_body:   "OK",
		},
		{
			desc:       "basic auth wrong user",
			user:       "wrong.user",
			pass:       "test",
			exp_status: http.StatusUnauthorized,
			exp_body:   "Unauthorized",
		},
		{
			desc:       "basic auth wrong password",
			user:       "testuser",
			pass:       "wrong password",
			exp_status: http.StatusUnauthorized,
			exp_body:   "Unauthorized",
		},
		{
			desc:       "basic auth empty user",
			user:       "",
			pass:       "test",
			exp_status: http.StatusUnauthorized,
			exp_body:   "Unauthorized",
		},
		{
			desc:       "basic auth empty password",
			user:       "testuser",
			pass:       "",
			exp_status: http.StatusUnauthorized,
			exp_body:   "Unauthorized",
		},
		{
			desc:       "basic auth empty user and password",
			user:       "",
			pass:       "",
			exp_status: http.StatusUnauthorized,
			exp_body:   "Unauthorized",
		},
	}

	for _, c := range cases {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(c.user, c.pass)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(BasicAuth(okResponder))
		handler.ServeHTTP(rr, req)

		if got := rr.Code; got != c.exp_status {
			t.Errorf("%v: exp status: %v got status: %v", c.desc, c.exp_status, got)
		}

		if got := rr.Body.String(); strings.TrimSpace(got) != c.exp_body {
			t.Errorf("%v: exp body: '%v' got body: '%v'", c.desc, c.exp_body, got)
		}
	}
}

func TestBasicAuthHeader(t *testing.T) {
	exp := make(http.Header)
	exp.Add("Authorization", "Basic dGVzdHVzZXI6dGVzdA==")

	got := BasicAuthHeader("testuser", "test")
	if !reflect.DeepEqual(got, exp) {
		t.Errorf("basic auth header mismatch: exp: %v got: %v", exp, got)
	}
}

func TestPassword(t *testing.T) {
	conf.Set("users.testuser",
		"09d9623a149a4a0c043befcb448c9c3324be973230188ba412c008a2929f31d0")

	exp := false
	if got := User("testuser", "wrong password"); got != exp {
		t.Errorf("wrong password authenticated: exp: %v got: %v", exp, got)
	}

	exp = true
	if got := User("testuser", "test"); got != exp {
		t.Errorf("password failed to authenticate: exp: %v got: %v", exp, got)
	}
}
