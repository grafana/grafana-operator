package cache

import (
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompressDecompress(t *testing.T) {
	contentJSON := []byte(`{"dummyField": "dummyData"}`)

	compressed, err := Gzip(contentJSON)
	require.NoError(t, err)

	decompressed, err := Gunzip(compressed)
	require.NoError(t, err)

	require.JSONEq(t, string(contentJSON), string(decompressed))
}

func TestGunzipRejectsOversizedPayload(t *testing.T) {
	var raw bytes.Buffer
	gz := gzip.NewWriter(&raw)
	// Highly compressible zeros exceed MaxGunzipSize after inflate.
	_, err := gz.Write(make([]byte, MaxGunzipSize+1024))
	require.NoError(t, err)
	require.NoError(t, gz.Close())

	_, err = Gunzip(raw.Bytes())
	require.Error(t, err)
	require.Contains(t, err.Error(), "maximum allowed size")
}
