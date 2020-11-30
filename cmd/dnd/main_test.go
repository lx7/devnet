package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	conf "github.com/spf13/viper"
)

func init() {
	configure("../../configs/server.yaml")
}

func TestServerCmdConfig(t *testing.T) {
	// flag defaults
	exp := "info"
	if got := conf.GetString("loglevel"); got != exp {
		t.Errorf("get loglevel flag default: exp: '%v' got: '%v'", exp, got)
	}

	// value from config file
	exp = "/channel"
	if got := conf.GetString("server.wspath"); got != exp {
		t.Errorf("get wspath setting from config: exp: '%v' got: '%v'", exp, got)
	}

}

func TestServerCmdResponse(t *testing.T) {
	go run()

	// allow for some startup time
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://%s", conf.GetString("server.addr")))
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	exp := "OK"
	if got := string(body); got != exp {
		t.Errorf("response body: exp: '%v' got: '%v'", exp, got)
	}
}
