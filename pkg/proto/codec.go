package proto

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

type Codec interface {
	ReadMessage(io.Reader) (*Message, error)
	WriteMessage(w io.Writer, msg *Message) error
	Marshal(msg any) ([]byte, error)
	Unmarshal(b []byte, msg any) error
}

const (
	Magic   uint16 = 0x4E4F // 协议魔数
	Version uint8  = 1      // 协议版本

	maxMessageSize = 64 << 20 // 64 MB 上限，防止内存爆炸
)

type codec struct{}

func NewCodec() Codec {
	return &codec{}
}

type packet struct {
	magic     uint16 // 协议魔数
	version   uint8  // 协议版本
	codecType uint8  // 数据编码类型：1-protobuf，2-json
	length    uint32 // 数据长度
	data      []byte // 实际数据
}

func (c *codec) ReadMessage(r io.Reader) (*Message, error) {
	p := new(packet)
	if err := binary.Read(r, binary.BigEndian, &p.magic); err != nil {
		return nil, fmt.Errorf("read magic: %w", err)
	}
	if p.magic != Magic {
		return nil, fmt.Errorf("invalid magic: 0x%X", p.magic)
	}
	if err := binary.Read(r, binary.BigEndian, &p.version); err != nil {
		return nil, fmt.Errorf("read version: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &p.codecType); err != nil {
		return nil, fmt.Errorf("read codec type: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &p.length); err != nil {
		return nil, fmt.Errorf("read length: %w", err)
	}
	if p.length > maxMessageSize {
		return nil, fmt.Errorf("message too large: %d bytes", p.length)
	}
	p.data = make([]byte, p.length)
	if _, err := io.ReadFull(r, p.data); err != nil {
		return nil, fmt.Errorf("read data: %w", err)
	}
	var msg Message
	switch p.codecType {
	case 1: // protobuf
		if err := proto.Unmarshal(p.data, &msg); err != nil {
			return nil, fmt.Errorf("proto unmarshal: %w", err)
		}
	case 2: // json
		if err := json.Unmarshal(p.data, &msg); err != nil {
			return nil, fmt.Errorf("json unmarshal: %w", err)
		}
	}
	return &msg, nil
}

func (c *codec) WriteMessage(w io.Writer, msg *Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("proto marshal: %w", err)
	}
	p := &packet{
		magic:     Magic,
		version:   Version,
		codecType: 1,
		length:    uint32(len(data)),
		data:      data,
	}
	if err := binary.Write(w, binary.BigEndian, p.magic); err != nil {
		return fmt.Errorf("write magic: %w", err)
	}
	if err := binary.Write(w, binary.BigEndian, p.version); err != nil {
		return fmt.Errorf("write version: %w", err)
	}
	if err := binary.Write(w, binary.BigEndian, p.codecType); err != nil {
		return fmt.Errorf("write codec type: %w", err)
	}
	if err := binary.Write(w, binary.BigEndian, p.length); err != nil {
		return fmt.Errorf("write length: %w", err)
	}
	if _, err := w.Write(p.data); err != nil {
		return fmt.Errorf("write data: %w", err)
	}
	return nil
}

func (c *codec) Marshal(msg any) ([]byte, error) {
	return proto.Marshal(msg.(proto.Message))
}

func (c *codec) Unmarshal(b []byte, msg any) error {
	return proto.Unmarshal(b, msg.(proto.Message))
}
