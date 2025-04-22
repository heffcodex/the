package tcfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testB64RawKey = "LS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0"
	testB64PadKey = "LS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0="
	testRawKey    = "--------------------------------"
	testPadKey    = "--------------------------------"
)

func TestKey_Validate(t *testing.T) {
	t.Parallel()

	assert.NoError(t, Key(testB64RawKey).Validate())
	assert.NoError(t, Key(testB64PadKey).Validate())
	assert.NoError(t, Key(testRawKey).Validate())
	assert.ErrorContains(t, Key("foo").Validate(), "decode")                 //nolint:testifylint // don't want require here
	assert.ErrorContains(t, Key(testRawKey+testRawKey).Validate(), "decode") //nolint:testifylint // don't want require here
	assert.ErrorIs(t, Key(testB64RawKey+testB64RawKey).Validate(), ErrInvalidKeyLength)
}

func TestKey_Bytes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []byte(testRawKey), Key(testB64RawKey).Bytes())
	assert.Equal(t, []byte(testPadKey), Key(testB64PadKey).Bytes())
	assert.Equal(t, []byte(testRawKey), Key(testRawKey).Bytes())
	assert.Panics(t, func() { Key("foo").Bytes() })
	assert.Panics(t, func() { Key(testRawKey + testRawKey).Bytes() })
	assert.Panics(t, func() { Key(testB64RawKey + testB64RawKey).Bytes() })
}

func TestKey_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, testB64RawKey, Key(testB64RawKey).String())
	assert.Equal(t, testB64PadKey, Key(testB64PadKey).String())
	assert.Equal(t, testRawKey, Key(testRawKey).String())
}

func TestKey_Derive(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "oq5GjT90InV6Ze1Vq/F+oxZFOcQRRBBwMm8zf2w6zbc", Key(testB64RawKey).Derive("").String())
	assert.Equal(t, "oq5GjT90InV6Ze1Vq/F+oxZFOcQRRBBwMm8zf2w6zbc", Key(testB64RawKey).Derive("").String())
	assert.Equal(t, "oq5GjT90InV6Ze1Vq/F+oxZFOcQRRBBwMm8zf2w6zbc", Key(testRawKey).Derive("").String())
	assert.Equal(t, "MQ1qskbONKB91roQkNuW3iuGfXOAeQNxoanndOloJC0", Key(testB64RawKey).Derive("some$alt").String())
	assert.Equal(t, "MQ1qskbONKB91roQkNuW3iuGfXOAeQNxoanndOloJC0", Key(testB64RawKey).Derive("some$alt").String())
	assert.Equal(t, "MQ1qskbONKB91roQkNuW3iuGfXOAeQNxoanndOloJC0", Key(testRawKey).Derive("some$alt").String())
}
