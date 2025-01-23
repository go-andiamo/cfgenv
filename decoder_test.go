package cfgenv

import (
	"encoding/base64"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBase64Decoder_Decode(t *testing.T) {
	dec := NewBase64Decoder()
	require.Equal(t, encodingBase64, dec.Encoding())
	value := base64.StdEncoding.EncodeToString([]byte("hello world"))
	decoded, err := dec.Decode(value)
	require.NoError(t, err)
	require.Equal(t, "hello world", decoded)
}

func TestBase64Decoder_Decode_Error(t *testing.T) {
	dec := NewBase64Decoder()
	_, err := dec.Decode("not properly encoded")
	require.Error(t, err)
}

func TestBase64UrlDecoder_Decode(t *testing.T) {
	dec := NewBase64UrlDecoder()
	require.Equal(t, encodingBase64Url, dec.Encoding())
	value := base64.URLEncoding.EncodeToString([]byte("hello world"))
	decoded, err := dec.Decode(value)
	require.NoError(t, err)
	require.Equal(t, "hello world", decoded)
}

func TestBase64UrlDecoder_Decode_Error(t *testing.T) {
	dec := NewBase64UrlDecoder()
	_, err := dec.Decode("not properly encoded")
	require.Error(t, err)
}

func TestRawBase64Decoder_Decode(t *testing.T) {
	dec := NewRawBase64Decoder()
	require.Equal(t, encodingRawBase64, dec.Encoding())
	value := base64.RawStdEncoding.EncodeToString([]byte("hello world"))
	decoded, err := dec.Decode(value)
	require.NoError(t, err)
	require.Equal(t, "hello world", decoded)
}

func TestRawBase64Decoder_Decode_Error(t *testing.T) {
	dec := NewRawBase64Decoder()
	_, err := dec.Decode("not properly encoded")
	require.Error(t, err)
}

func TestRawBase64UrlDecoder_Decode(t *testing.T) {
	dec := NewRawBase64UrlDecoder()
	require.Equal(t, encodingRawBase64Url, dec.Encoding())
	value := base64.RawURLEncoding.EncodeToString([]byte("hello world"))
	decoded, err := dec.Decode(value)
	require.NoError(t, err)
	require.Equal(t, "hello world", decoded)
}

func TestRawBase64UrlDecoder_Decode_Error(t *testing.T) {
	dec := NewRawBase64UrlDecoder()
	_, err := dec.Decode("not properly encoded")
	require.Error(t, err)
}
