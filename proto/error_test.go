package proto

import (
	"strings"
	"testing"
)

func TestError(t *testing.T) {
	cases := []struct {
		desc string
		err  error
		exp  string
	}{
		{
			desc: "InvalidMessageType",
			err: InvalidMessageTypeError{
				_type: MessageTypeSDP,
				_json: []byte("{}"),
			},
			exp: "unknown message type",
		},
		{
			desc: "UnexpectedMessageType",
			err: UnexpectedMessageTypeError{
				_type: "other",
				_exp:  MessageTypeSDP,
			},
			exp: "unexpected message type",
		},
	}

	for _, c := range cases {
		if got := c.err.Error(); !strings.Contains(got, c.exp) {
			t.Errorf("%v: expected substring: '%v' got: '%v'", c.desc, c.exp, got)
		}
	}
}
