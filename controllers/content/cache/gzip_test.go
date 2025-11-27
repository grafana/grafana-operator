package cache

import (
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
