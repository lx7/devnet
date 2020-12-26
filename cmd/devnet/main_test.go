package main

import (
	"testing"
	"time"

	"github.com/lx7/devnet/internal/server"
	"github.com/lx7/devnet/internal/testutil"
	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
	conf "github.com/spf13/viper"
)

func init() {
	configure("../../configs/client.yaml")

}

func TestClientCmd_Config(t *testing.T) {
	// define cases
	tests := []struct {
		desc string
		give string
		want string
	}{
		{
			desc: "get log level from config",
			give: "log.level",
			want: "info",
		},
		{
			desc: "get user from config",
			give: "auth.user",
			want: "user1",
		},
		{
			desc: "get signaling url from config",
			give: "signaling.url",
			want: "ws://localhost:8080/channel",
		},
		{
			desc: "get video encoder from config",
			give: "video.encode",
			want: "software",
		},
		{
			desc: "get video decoder from config",
			give: "video.decode",
			want: "auto",
		},
	}

	// execute cases
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			have := conf.GetString(tt.give)
			assert.Equal(t, tt.want, have)
		})
	}
}

func TestClientCmd_Run(t *testing.T) {
	hook := testutil.NewLogHook()

	conf.Set("users.testuser",
		"09d9623a149a4a0c043befcb448c9c3324be973230188ba412c008a2929f31d0")
	conf.Set("signaling.url", "ws://127.0.0.1:40101/channel")
	conf.Set("auth.user", "testuser")
	conf.Set("auth.pass", "test")

	// create test server
	sconf := conf.New()
	sconf.SetConfigFile("../../configs/server.yaml")
	if err := sconf.ReadInConfig(); err != nil {
		log.Fatal("failed reading server config: ", err)
	}
	sconf.Set("signaling.addr", "127.0.0.1:40101")
	s := server.New(sconf)
	go s.Serve()
	time.Sleep(100 * time.Millisecond)

	go func() {
		time.Sleep(1 * time.Second)

		quit()
	}()

	// start client
	exitcode := run()
	assert.Equal(t, 0, exitcode, "run should exit with code 0")

	errorlog := hook.Entry(log.ErrorLevel)
	if errorlog != nil {
		t.Errorf("runtime error: '%v'", errorlog.Message)
	}
}
