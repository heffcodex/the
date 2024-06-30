package tcfg

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/hkdf"
)

type App struct {
	Name            string `mapstructure:"name" json:"name" yaml:"name"`
	Key             Key    `mapstructure:"key" json:"key" yaml:"key"`
	Env             Env    `mapstructure:"env" json:"env" yaml:"env"`
	LogLevel        string `mapstructure:"logLevel" json:"logLevel" yaml:"logLevel"`
	ShutdownTimeout int    `mapstructure:"shutdownTimeout" json:"shutdownTimeout" yaml:"shutdownTimeout"`
}

// -----------------------------------------------------------------------------------------------------------------------------------------

type Env string

const (
	EnvDev   Env = "dev"
	EnvTest  Env = "test"
	EnvStage Env = "staging"
	EnvProd  Env = "production"
)

func (e Env) IsEmpty() bool {
	return e == ""
}

func (e Env) String() string {
	return string(e)
}

// -----------------------------------------------------------------------------------------------------------------------------------------

const (
	KeyLenBytes = 32
)

var (
	KeyEncoding         = base64.StdEncoding
	ErrInvalidKeyLength = fmt.Errorf("invalid key length (must be %d either raw or decoded)", KeyLenBytes)
)

type Key string

func (k Key) Validate() error {
	_, err := k.getBytes()
	return err
}

func (k Key) Bytes() []byte {
	b, err := k.getBytes()
	if err != nil {
		panic(err)
	}

	return b
}

func (k Key) String() string {
	return string(k)
}

func (k Key) Derive(salt string) Key {
	derived := hkdf.Extract(sha256.New, k.Bytes(), []byte(salt))
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
