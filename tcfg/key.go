package tcfg

import (
	"crypto/hkdf"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const (
	KeyLenBytes = 32
)

var (
	KeyEncoding         = base64.RawStdEncoding
	KeyDerivationAlgo   = sha256.New
	KeyErrorHandler     = func(e error) { panic(e) }
	ErrInvalidKeyLength = fmt.Errorf("invalid key length (must be %d either raw or b64-decoded)", KeyLenBytes)
)

type Key string

func (k Key) Validate() error {
	_, err := k.getBytes()
	return err
}

func (k Key) Bytes() []byte {
	b, err := k.getBytes()
	if err != nil {
		KeyErrorHandler(fmt.Errorf("get bytes: %w", err))
	}

	return b
}

func (k Key) String() string {
	return string(k)
}

func (k Key) Derive(salt string) Key {
	derived, err := hkdf.Extract(KeyDerivationAlgo, k.Bytes(), []byte(salt))
	if err != nil {
		KeyErrorHandler(fmt.Errorf("derive from key: %w", err))
	}

	encoded := KeyEncoding.EncodeToString(derived)

	return Key(encoded)
}

func (k Key) getBytes() ([]byte, error) {
	if len(k) == KeyLenBytes {
		return []byte(k), nil
	}

	decoded, err := KeyEncoding.DecodeString(string(k))
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	} else if len(decoded) != KeyLenBytes {
		return nil, fmt.Errorf("%w: raw=%d, dec=%d", ErrInvalidKeyLength, len(k), len(decoded))
	}

	return decoded, nil
}
