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
	header := make([]byte, 8)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	p.magic = binary.BigEndian.Uint16(header[0:2])
	p.version = header[2]
	p.codecType = header[3]
	p.length = binary.BigEndian.Uint32(header[4:8])
	if p.magic != Magic {
		return nil, fmt.Errorf("invalid magic: 0x%X", p.magic)
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
	buf := make([]byte, 8+len(data))
	binary.BigEndian.PutUint16(buf[0:2], Magic)
	buf[2] = Version
	buf[3] = 1 // codecType: protobuf
	binary.BigEndian.PutUint32(buf[4:8], uint32(len(data)))
	copy(buf[8:], data)
	if _, err := w.Write(buf); err != nil {
		return fmt.Errorf("write message: %w", err)
	}
	return nil
}

func (c *codec) Marshal(msg any) ([]byte, error) {
	return proto.Marshal(msg.(proto.Message))
}

func (c *codec) Unmarshal(b []byte, msg any) error {
	return proto.Unmarshal(b, msg.(proto.Message))
}
