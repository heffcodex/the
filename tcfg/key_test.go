package tcfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testB64Key = "LS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0"
	testRawKey = "--------------------------------"
)

func TestKey_Validate(t *testing.T) {
	t.Parallel()

	assert.NoError(t, Key(testB64Key).Validate())
	assert.NoError(t, Key(testRawKey).Validate())
	assert.ErrorContains(t, Key("foo").Validate(), "decode")                 //nolint:testifylint // don't want require here
	assert.ErrorContains(t, Key(testRawKey+testRawKey).Validate(), "decode") //nolint:testifylint // don't want require here
	assert.ErrorIs(t, Key(testB64Key+testB64Key).Validate(), ErrInvalidKeyLength)
}

func TestKey_Bytes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []byte(testRawKey), Key(testB64Key).Bytes())
	assert.Equal(t, []byte(testRawKey), Key(testRawKey).Bytes())
	assert.Panics(t, func() { Key("foo").Bytes() })
	assert.Panics(t, func() { Key(testRawKey + testRawKey).Bytes() })
	assert.Panics(t, func() { Key(testB64Key + testB64Key).Bytes() })
}

func TestKey_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, testB64Key, Key(testB64Key).String())
	assert.Equal(t, testRawKey, Key(testRawKey).String())
}

func TestKey_Derive(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "oq5GjT90InV6Ze1Vq/F+oxZFOcQRRBBwMm8zf2w6zbc", Key(testB64Key).Derive("").String())
	assert.Equal(t, "oq5GjT90InV6Ze1Vq/F+oxZFOcQRRBBwMm8zf2w6zbc", Key(testRawKey).Derive("").String())
	assert.Equal(t, "MQ1qskbONKB91roQkNuW3iuGfXOAeQNxoanndOloJC0", Key(testB64Key).Derive("some$alt").String())
	assert.Equal(t, "MQ1qskbONKB91roQkNuW3iuGfXOAeQNxoanndOloJC0", Key(testRawKey).Derive("some$alt").String())
}
