package io

import (
	"encoding/binary"
	"io"
)

// DatagramStream wraps a stream with length-prefixed datagram framing.
// Format: [4-byte length (big-endian)][payload]
// Read and Write can safely be called concurrently from different goroutines
// when the underlying stream supports full-duplex Read/Write.
type DatagramStream struct {
	rw io.ReadWriter
}

func NewDatagramStream(rw io.ReadWriter) *DatagramStream {
	return &DatagramStream{rw: rw}
}

func (d *DatagramStream) ReadDatagram() ([]byte, error) {
	var length uint32
	if err := binary.Read(d.rw, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	if length > 65536 {
		return nil, io.ErrUnexpectedEOF
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(d.rw, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (d *DatagramStream) WriteDatagram(data []byte) error {
	if err := binary.Write(d.rw, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}
	_, err := d.rw.Write(data)
	return err
}
