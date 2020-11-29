package proto

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.ErrorLevel)
}

func TestMarshal(t *testing.T) {
	cases := []struct {
		desc string
		data SDPMessage
		exp  string
	}{
		{
			desc: "marshal SDPMessage",
			data: SDPMessage{
				Src: "user 1",
				Dst: "user 2",
				SDP: webrtc.SessionDescription{
					Type: webrtc.SDPTypeOffer,
					SDP:  "sdp",
				},
			},
			exp: `{
				"type":"sdp", 
				"src":"user 1", 
				"dst":"user 2", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`,
		},
	}

	diffOpts := jsondiff.DefaultConsoleOptions()

	for _, c := range cases {
		got, err := c.data.MarshalJSON()
		if err != nil {
			t.Errorf("%v: %v", c.desc, err)
		}

		res, diff := jsondiff.Compare(got, []byte(c.exp), &diffOpts)
		if res != jsondiff.FullMatch {
			t.Errorf("%v: diff: %v", c.desc, diff)
		}

	}
}

func TestUnmarshal(t *testing.T) {
	cases := []struct {
		desc string
		json string
		exp  SDPMessage
		err  error
	}{
		{
			desc: "unmarshal SDPMessage",
			json: `{
				"type":"sdp", 
				"src":"user 1", 
				"dst":"user 2", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`,
			exp: SDPMessage{
				Src: "user 1",
				Dst: "user 2",
				SDP: webrtc.SessionDescription{
					Type: webrtc.SDPTypeOffer,
					SDP:  "sdp",
				},
			},
			err: nil,
		},
		{
			desc: "unmarshal invalid type",
			json: `{
				"type":"12345", 
				"src":"user 1", 
				"dst":"user 2", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`,
			exp: SDPMessage{},
			err: UnexpectedMessageTypeError{},
		},
	}

	for _, c := range cases {
		var got SDPMessage
		err := json.Unmarshal([]byte(c.json), &got)
		if err != c.err && reflect.TypeOf(err) != reflect.TypeOf(c.err) {
			t.Errorf("%v: %v", c.desc, err)
		}

		if !reflect.DeepEqual(c.exp, got) {
			t.Errorf("%v: exp: %v got: %v", c.desc, c.exp, got)
		}

	}
}
