package io

import (
	"bytes"
	"io"
	"testing"
)

func TestDatagramStreamRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	s := NewDatagramStream(&buf)

	payloads := [][]byte{
		[]byte("hello"),
		[]byte(""),
		[]byte("x"),
		make([]byte, 65535),
	}
	for _, p := range payloads {
		if err := s.WriteDatagram(p); err != nil {
			t.Fatalf("WriteDatagram(%d bytes): %v", len(p), err)
		}
	}

	r := NewDatagramStream(&buf)
	for _, expected := range payloads {
		got, err := r.ReadDatagram()
		if err != nil {
			t.Fatalf("ReadDatagram: %v", err)
		}
		if !bytes.Equal(got, expected) {
			t.Fatalf("payload mismatch: got %d bytes, want %d bytes", len(got), len(expected))
		}
	}
}

func TestDatagramStreamEmpty(t *testing.T) {
	var buf bytes.Buffer
	s := NewDatagramStream(&buf)

	if err := s.WriteDatagram([]byte{}); err != nil {
		t.Fatalf("WriteDatagram empty: %v", err)
	}

	got, err := s.ReadDatagram()
	if err != nil {
		t.Fatalf("ReadDatagram: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty, got %d bytes", len(got))
	}
}

func TestDatagramStreamTooLarge(t *testing.T) {
	var buf bytes.Buffer
	s := NewDatagramStream(&buf)

	// Max allowed is 65536
	data := make([]byte, 65537)
	s.WriteDatagram(data)

	r := NewDatagramStream(&buf)
	_, err := r.ReadDatagram()
	if err != io.ErrUnexpectedEOF {
		t.Fatalf("expected ErrUnexpectedEOF, got %v", err)
	}
}

func TestDatagramStreamPartialRead(t *testing.T) {
	var buf bytes.Buffer
	// Write a valid header but wrong body
	header := []byte{0x00, 0x00, 0x00, 0x05} // length=5
	buf.Write(header)
	buf.Write([]byte{0x01}) // only 1 byte instead of 5

	r := NewDatagramStream(&buf)
	_, err := r.ReadDatagram()
	if err != io.ErrUnexpectedEOF {
		t.Fatalf("expected ErrUnexpectedEOF, got %v", err)
	}
}

func TestDatagramStreamMultipleDatagrams(t *testing.T) {
	var buf bytes.Buffer
	s := NewDatagramStream(&buf)

	count := 100
	for i := 0; i < count; i++ {
		data := []byte{byte(i), byte(i >> 8)}
		if err := s.WriteDatagram(data); err != nil {
			t.Fatalf("WriteDatagram %d: %v", i, err)
		}
	}

	r := NewDatagramStream(&buf)
	for i := 0; i < count; i++ {
		got, err := r.ReadDatagram()
		if err != nil {
			t.Fatalf("ReadDatagram %d: %v", i, err)
		}
		if len(got) != 2 || got[0] != byte(i) || got[1] != byte(i>>8) {
			t.Fatalf("payload mismatch at %d: %v", i, got)
		}
	}
}
