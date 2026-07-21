package cache

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// MaxGunzipSize caps decompressed gzip payload size to avoid OOM from
// highly compressed inputs (zip bombs). See grafana/grafana-operator#2815.
const MaxGunzipSize = 10 * 1024 * 1024 // 10 MiB

func Gunzip(compressed []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	data, err := io.ReadAll(io.LimitReader(gz, MaxGunzipSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxGunzipSize {
		return nil, fmt.Errorf("gzip content exceeds maximum allowed size of %d bytes", MaxGunzipSize)
	}
	return data, nil
}

func Gzip(content []byte) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	gz := gzip.NewWriter(buf)

	_, err := gz.Write(content)
	if err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return io.ReadAll(buf)
}
