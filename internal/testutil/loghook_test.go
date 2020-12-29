package testutil

import (
	"io/ioutil"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogHook(t *testing.T) {
	h := &LogHook{}
	l := zerolog.New(ioutil.Discard).Hook(h)

	assert.Nil(t, h.Entry(zerolog.ErrorLevel))

	l.Error().Msg("error")
	require.NotNil(t, h.Entry(zerolog.ErrorLevel))
	assert.Equal(t, h.Entry(zerolog.ErrorLevel).Level, zerolog.ErrorLevel)

	h.Reset()
	assert.Nil(t, h.Entry(zerolog.ErrorLevel))
}
