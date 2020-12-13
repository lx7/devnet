package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/lx7/devnet/proto"

	pb "github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	conf "github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetLevel(log.InfoLevel)

	conf.Set("users.testuser",
		"09d9623a149a4a0c043befcb448c9c3324be973230188ba412c008a2929f31d0")

	s := New("127.0.0.1:40100")
	go s.Serve("/channel")
	time.Sleep(20 * time.Millisecond)
}

func TestServer_Echo(t *testing.T) {
	// connect websocket
	header := make(http.Header)
	header.Add("Authorization", "Basic dGVzdHVzZXI6dGVzdA==")
	url := "ws://127.0.0.1:40100/channel"
	ws, _, err := websocket.DefaultDialer.Dial(url, header)
	require.NoError(t, err)
	defer ws.Close()

	// define cases
	tests := []struct {
		desc     string
		give     *proto.Frame
		giveType int
		want     *proto.Frame
	}{
		{
			desc:     "echo",
			give:     &proto.Frame{Src: "testuser", Dst: "testuser"},
			giveType: websocket.BinaryMessage,
			want:     &proto.Frame{Src: "testuser", Dst: "testuser"},
		},
	}

	// execute test cases
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			out, err := tt.give.Marshal()
			require.NoError(t, err, "marshal should not cause an error")
			err = ws.WriteMessage(tt.giveType, out)
			require.NoError(t, err, "ws write should not cause an error")
			time.Sleep(10 * time.Millisecond)

			_, in, err := ws.ReadMessage()
			require.NoError(t, err, "ws read should not cause an error")

			have := &proto.Frame{}
			err = have.Unmarshal(in)
			require.NoError(t, err, "unmarshal should not cause an error")
			if !pb.Equal(tt.want, have) {
				t.Errorf("want: %v\nhave: %v\n", tt.want, have)
			}
		})
	}
}

func TestServer_Auth(t *testing.T) {
	// define cases
	tests := []struct {
		desc     string
		giveAuth string
		wantErr  error
	}{
		{
			desc:     "no auth header",
			giveAuth: "",
			wantErr:  websocket.ErrBadHandshake,
		},
		{
			desc:     "invalid auth header",
			giveAuth: "Basic kDgmmNnabzatzZmvAV",
			wantErr:  websocket.ErrBadHandshake,
		},
		{
			desc:     "correct auth header",
			giveAuth: "Basic dGVzdHVzZXI6dGVzdA==",
			wantErr:  nil,
		},
	}

	url := "ws://127.0.0.1:40100/channel"

	// run tests
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			header := make(http.Header)
			if tt.giveAuth != "" {
				header.Add("Authorization", tt.giveAuth)
			}
			ws, _, err := websocket.DefaultDialer.Dial(url, header)
			require.Equal(t, tt.wantErr, err)

			if ws != nil {
				ws.Close()
			}
		})
	}
}
