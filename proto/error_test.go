package proto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	tests := []struct {
		desc    string
		give    error
		wantStr string
	}{
		{
			desc: "InvalidMessageType",
			give: InvalidMessageTypeError{
				_type: messageTypeSDP,
				_json: []byte("{}"),
			},
			wantStr: "unknown message type",
		},
		{
			desc: "UnexpectedMessageType",
			give: UnexpectedMessageTypeError{
				_type: "other",
				_exp:  messageTypeSDP,
			},
			wantStr: "unexpected message type",
		},
	}

	for _, tt := range tests {
		assert.Contains(t, tt.give.Error(), tt.wantStr)
	}
}
