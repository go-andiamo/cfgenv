package cfgenv

import "encoding/base64"

const (
	encodingBase64       = "base64"
	encodingBase64Url    = "base64url"
	encodingRawBase64    = "rawBase64"
	encodingRawBase64Url = "rawBase64url"
)

// Decoder is an option that can be passed to Load or LoadAs
// and provides support for decoding values
//
// Values that are encoded can be denoted with field tags, e.g.
//
//	type MyConfig struct {
//	  Key string `env:"encoding=base64"`
//	}
type Decoder interface {
	// Encoding returns the encoding (e.g. "base64") that this Decoder supports
	Encoding() string
	// Decode returns the decoded value
	Decode(value string) (string, error)
}

// NewBase64Decoder returns a new Decoder for decoding base64
func NewBase64Decoder() Decoder {
	return &base64Decoder{}
}

type base64Decoder struct{}

var _ Decoder = (*base64Decoder)(nil)

func (d *base64Decoder) Encoding() string {
	return encodingBase64
}

func (d *base64Decoder) Decode(value string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// NewBase64UrlDecoder returns a new Decoder for decoding base64url
func NewBase64UrlDecoder() Decoder {
	return &base64UrlDecoder{}
}

type base64UrlDecoder struct{}

var _ Decoder = (*base64UrlDecoder)(nil)

func (d *base64UrlDecoder) Encoding() string {
	return encodingBase64Url
}

func (d *base64UrlDecoder) Decode(value string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// NewRawBase64Decoder returns a new Decoder for decoding raw base64 (no padding)
func NewRawBase64Decoder() Decoder {
	return &rawBase64Decoder{}
}

type rawBase64Decoder struct{}

var _ Decoder = (*rawBase64Decoder)(nil)

func (d *rawBase64Decoder) Encoding() string {
	return encodingRawBase64
}

func (d *rawBase64Decoder) Decode(value string) (string, error) {
	data, err := base64.RawStdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// NewRawBase64UrlDecoder returns a new Decoder for decoding raw base64url (no padding)
func NewRawBase64UrlDecoder() Decoder {
	return &rawBase64UrlDecoder{}
}

type rawBase64UrlDecoder struct{}

var _ Decoder = (*rawBase64UrlDecoder)(nil)

func (d *rawBase64UrlDecoder) Encoding() string {
	return encodingRawBase64Url
}

func (d *rawBase64UrlDecoder) Decode(value string) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
