package proto

import (
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.ErrorLevel)
}

func TestUnmarshal(t *testing.T) {
	cases := []struct {
		desc string
		json string
		exp  interface{}
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
			exp: SDPMessage{},
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
