package proto

import (
	"reflect"
	"testing"

	"github.com/nsf/jsondiff"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.ErrorLevel)
}

func TestMarshal(t *testing.T) {
	cases := []struct {
		desc string
		data Message
		exp  string
		err  error
	}{
		{
			desc: "marshal SDPMessage",
			data: &SDPMessage{},
			exp: `{
				"type":"sdp", 
				"src":"", 
				"dst":"", 
				"sdp":{"sdp":"", "type":"unknown"} 
			}`,
			err: nil,
		},
	}

	diffOpts := jsondiff.DefaultConsoleOptions()

	for _, c := range cases {
		got, err := Marshal(c.data)
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
		exp  Message
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
			exp: &SDPMessage{},
			err: nil,
		},
	}

	for _, c := range cases {
		got, err := Unmarshal([]byte(c.json))
		if err != nil {
			t.Errorf("%v: %v", c.desc, err)
		}

		if reflect.TypeOf(c.exp) != reflect.TypeOf(got) {
			t.Errorf("%v: exp: %v got: %v", c.desc,
				reflect.TypeOf(c.exp), reflect.TypeOf(got))
		}
	}
}
