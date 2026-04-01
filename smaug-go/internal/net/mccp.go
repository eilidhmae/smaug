package net

import (
	"bytes"
	"compress/zlib"
)

// CompressData compresses data using zlib (MCCP2).
func CompressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecompressData decompresses zlib data.
func DecompressData(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MCCPStartSequence returns the telnet sequence to begin MCCP2 compression.
// IAC SB COMPRESS2 IAC SE — after this, all output is zlib compressed.
func MCCPStartSequence() []byte {
	return TelnetSubneg(TELOPT_COMPRESS2, nil)
}
