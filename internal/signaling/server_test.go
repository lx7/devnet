package signaling

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lx7/devnet/proto"
	"github.com/rs/zerolog"

	"github.com/gorilla/websocket"
	"github.com/posener/wstest"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "google.golang.org/protobuf/proto"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})
}

var conf *viper.Viper

func init() {
	conf = viper.New()
	conf.SetConfigFile("../../configs/signald.yaml")
	if err := conf.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("config file")
	}
	conf.Set("signaling.addr", "127.0.0.1:40100")
	conf.Set("signaling.tls", "true")
	conf.Set("signaling.tls_crt", "../../test/localhost.crt")
	conf.Set("signaling.tls_key", "../../test/localhost.key")
}

func TestServer_Echo(t *testing.T) {
	// run server
	s := NewServer(conf)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.sw.Run()
	}()

	// connect websocket
	d := wstest.NewDialer(http.HandlerFunc(s.serveWS))
	header := make(http.Header)
	header.Add("Authorization", "Basic dGVzdHVzZXI6dGVzdA==")
	ws, _, err := d.Dial("ws://127.0.0.1/channel", header)
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
			desc:     "client config",
			give:     nil,
			giveType: websocket.BinaryMessage,
			want: &proto.Frame{
				Dst: "testuser",
				Payload: &proto.Frame_Config{Config: &proto.Config{
					Webrtc: &proto.Config_WebRTC{
						Iceservers: []*proto.Config_WebRTC_ICEServer{
							&proto.Config_WebRTC_ICEServer{
								Url: "stun:127.0.0.1:19302",
							},
						},
					},
				}},
			},
		},
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
			if tt.give != nil {
				out, err := tt.give.Marshal()
				require.NoError(t, err)
				err = ws.WriteMessage(tt.giveType, out)
				require.NoError(t, err)
				time.Sleep(10 * time.Millisecond)
			}
			_, in, err := ws.ReadMessage()
			require.NoError(t, err)

			have := &proto.Frame{}
			err = have.Unmarshal(in)
			require.NoError(t, err)
			if !pb.Equal(tt.want, have) {
				t.Errorf("want: %v\nhave: %v\n", tt.want, have)
			}
		})
	}
	ws.Close()
	s.sw.Shutdown()
	wg.Wait()
}

func TestServer_WSHandler(t *testing.T) {
	// create new server instance
	s := NewServer(conf)

	tests := []struct {
		desc     string
		give     *http.Request
		wantCode int
		wantBody string
	}{
		{
			desc: "without basic auth",
			give: func() *http.Request {
				req, err := http.NewRequest("GET", "/", nil)
				req.RemoteAddr = "0.0.0.0:80"
				require.NoError(t, err)
				return req
			}(),
			wantCode: http.StatusUnauthorized,
			wantBody: "Unauthorized",
		},
		{
			desc: "no ws upgrade token",
			give: func() *http.Request {
				req, err := http.NewRequest("GET", "/channel", nil)
				require.NoError(t, err)

				req.SetBasicAuth("testuser", "testpass")
				return req
			}(),
			wantCode: http.StatusBadRequest,
			wantBody: "Bad Request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(s.serveWS)

			l := zerolog.GlobalLevel()
			if tt.wantCode >= 300 {
				zerolog.SetGlobalLevel(zerolog.FatalLevel)
			}
			handler.ServeHTTP(rr, tt.give)
			zerolog.SetGlobalLevel(l)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}

func TestServer_OKHandler(t *testing.T) {
	// create new server instance
	s := NewServer(conf)

	tests := []struct {
		desc     string
		give     *http.Request
		wantCode int
		wantBody string
	}{
		{
			desc: "simple request",
			give: func() *http.Request {
				req, err := http.NewRequest("GET", "/", nil)
				require.NoError(t, err)
				return req
			}(),
			wantCode: http.StatusOK,
			wantBody: "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(s.serveOK)
			handler.ServeHTTP(rr, tt.give)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}

func TestServer_AuthRequest(t *testing.T) {
	// run server
	s := NewServer(conf)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.Serve(); err != nil {
			log.Fatal().Err(err).Msg("serve")
		}
	}()
	time.Sleep(10 * time.Millisecond)

	// define cases
	tests := []struct {
		desc     string
		giveAuth string
		giveURL  string
		wantCode int
		wantErr  error
	}{
		{
			desc:     "ws: no auth header",
			giveURL:  "wss://localhost:40100/channel",
			giveAuth: "",
			wantErr:  websocket.ErrBadHandshake,
		},
		{
			desc:     "ws: invalid auth header",
			giveURL:  "wss://localhost:40100/channel",
			giveAuth: "Basic kDgmmNnabzatzZmvAV",
			wantErr:  websocket.ErrBadHandshake,
		},
		{
			desc:     "ws: correct auth header",
			giveURL:  "wss://localhost:40100/channel",
			giveAuth: "Basic dGVzdHVzZXI6dGVzdA==",
		},
		{
			desc:     "plain: no auth header",
			giveURL:  "https://localhost:40100/",
			giveAuth: "",
			wantCode: 401,
		},
		{
			desc:     "plain: invalid auth header",
			giveURL:  "https://localhost:40100/",
			giveAuth: "Basic kDgmmNnabzatzZmvAV",
			wantCode: 401,
		},
		{
			desc:     "plain: correct auth header",
			giveURL:  "https://localhost:40100/",
			giveAuth: "Basic dGVzdHVzZXI6dGVzdA==",
			wantCode: 200,
		},
	}

	// run tests
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			header := make(http.Header)
			if tt.giveAuth != "" {
				header.Add("Authorization", tt.giveAuth)
			}

			if strings.HasPrefix(tt.giveURL, "wss://") {
				dialer := &websocket.Dialer{
					TLSClientConfig: configWithFakeCertPool(),
				}
				ws, _, err := dialer.Dial(tt.giveURL, header)
				require.Equal(t, tt.wantErr, err)
				if ws != nil {
					ws.Close()
				}
			} else {
				tr := &http.Transport{TLSClientConfig: configWithFakeCertPool()}
				client := &http.Client{Transport: tr}

				req, err := http.NewRequest("GET", tt.giveURL, nil)
				require.NoError(t, err)

				req.Header.Add("Authorization", tt.giveAuth)
				res, err := client.Do(req)
				require.Equal(t, tt.wantErr, err)
				assert.Equal(t, tt.wantCode, res.StatusCode)
			}
		})
	}
	s.Shutdown()
	wg.Wait()
}

func configWithFakeCertPool() *tls.Config {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	certs, err := ioutil.ReadFile(conf.GetString("signaling.tls_crt"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to append test cert to ca pool")
	}

	if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
		log.Fatal().Msg("no certs appended")
	}

	return &tls.Config{RootCAs: rootCAs}
}
