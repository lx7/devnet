package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	conf "github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})
	configure("../../configs/signald.yaml")
}

func TestServerCmd_Config(t *testing.T) {
	// flag defaults
	exp := "info"
	if got := conf.GetString("loglevel"); got != exp {
		t.Errorf("get loglevel flag default: exp: '%v' got: '%v'", exp, got)
	}

	// value from config file
	exp = "/channel"
	if got := conf.GetString("signaling.wspath"); got != exp {
		t.Errorf("get wspath setting from config: exp: '%v' got: '%v'", exp, got)
	}

}

func TestServerCmd_Response(t *testing.T) {
	go run()

	// allow for some startup time
	time.Sleep(10 * time.Millisecond)

	client := &http.Client{}
	url := "http://" + conf.GetString("signaling.addr")
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", "Basic dGVzdHVzZXI6dGVzdA==")
	res, err := client.Do(req)
	require.NoError(t, err)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, "OK", string(body))
}
