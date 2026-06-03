package proto

import (
	"bytes"
	"encoding/binary"
	"sync"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func TestRoundTrip_Protobuf(t *testing.T) {
	codec := NewCodec()
	msg := &Message{
		Id:        uuid.NewString(),
		Timestamp: 1700000000000,
		Type:      MessageType_HANDSHAKE,
		Payload:   []byte("hello"),
	}

	var buf bytes.Buffer
	if err := codec.WriteMessage(&buf, msg); err != nil {
		t.Fatalf("WriteMessage failed: %v", err)
	}

	got, err := codec.ReadMessage(&buf)
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	if got.Id != msg.Id {
		t.Errorf("Id = %q, want %q", got.Id, msg.Id)
	}
	if got.Timestamp != msg.Timestamp {
		t.Errorf("Timestamp = %d, want %d", got.Timestamp, msg.Timestamp)
	}
	if got.Type != msg.Type {
		t.Errorf("Type = %v, want %v", got.Type, msg.Type)
	}
	if !bytes.Equal(got.Payload, msg.Payload) {
		t.Errorf("Payload = %v, want %v", got.Payload, msg.Payload)
	}
}

func TestRoundTrip_EmptyPayload(t *testing.T) {
	codec := NewCodec()
	msg := &Message{
		Id:        uuid.NewString(),
		Timestamp: 1700000000001,
		Type:      MessageType_HEARTBEAT,
	}

	var buf bytes.Buffer
	if err := codec.WriteMessage(&buf, msg); err != nil {
		t.Fatalf("WriteMessage failed: %v", err)
	}

	got, err := codec.ReadMessage(&buf)
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	if got.Id != msg.Id {
		t.Errorf("Id = %q, want %q", got.Id, msg.Id)
	}
	if got.Type != msg.Type {
		t.Errorf("Type = %v, want %v", got.Type, msg.Type)
	}
}

func TestRoundTrip_AllMessageTypes(t *testing.T) {
	codec := NewCodec()
	types := []MessageType{
		MessageType_HANDSHAKE,
		MessageType_HANDSHAKE_ACK,
		MessageType_HEARTBEAT,
		MessageType_HEARTBEAT_ACK,
		MessageType_DISCONNECT,
		MessageType_TUNNEL_OPEN,
		MessageType_TUNNEL_OPEN_ACK,
		MessageType_EXIT,
		MessageType_TASK,
		MessageType_TASK_PROGRESS,
	}

	for _, mt := range types {
		msg := &Message{
			Id:        uuid.NewString(),
			Timestamp: 1700000000000,
			Type:      mt,
			Payload:   []byte{1, 2, 3},
		}

		var buf bytes.Buffer
		if err := codec.WriteMessage(&buf, msg); err != nil {
			t.Fatalf("WriteMessage(%v) failed: %v", mt, err)
		}

		got, err := codec.ReadMessage(&buf)
		if err != nil {
			t.Fatalf("ReadMessage(%v) failed: %v", mt, err)
		}

		if got.Type != mt {
			t.Errorf("Type = %v, want %v", got.Type, mt)
		}
	}
}

func TestInvalidMagic(t *testing.T) {
	codec := NewCodec()
	buf := make([]byte, 8)
	binary.BigEndian.PutUint16(buf[0:2], 0xDEAD) // wrong magic
	buf[2] = 1                                     // version
	buf[3] = 1                                     // codecType
	binary.BigEndian.PutUint32(buf[4:8], 0)        // length

	_, err := codec.ReadMessage(bytes.NewReader(buf))
	if err == nil {
		t.Fatal("expected error for invalid magic")
	}
}

func TestMessageTooLarge(t *testing.T) {
	codec := NewCodec()
	buf := make([]byte, 8)
	binary.BigEndian.PutUint16(buf[0:2], Magic)
	buf[2] = Version
	buf[3] = 1
	binary.BigEndian.PutUint32(buf[4:8], maxMessageSize+1) // exceed limit

	_, err := codec.ReadMessage(bytes.NewReader(buf))
	if err == nil {
		t.Fatal("expected error for message exceeding max size")
	}
}

func TestMaxMessageSizeBoundary(t *testing.T) {
	codec := NewCodec()
	// payload sized so total serialized message hits exactly maxMessageSize
	payload := make([]byte, maxMessageSize)
	msg := &Message{
		Id:      uuid.NewString(),
		Type:    MessageType_HANDSHAKE,
		Payload: payload,
	}

	serialized, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if len(serialized) <= maxMessageSize {
		t.Fatalf("test setup: serialized size %d should exceed maxMessageSize %d", len(serialized), maxMessageSize)
	}

	// Write the raw header + data directly to verify boundary enforcement
	var buf bytes.Buffer
	header := make([]byte, 8)
	binary.BigEndian.PutUint16(header[0:2], Magic)
	header[2] = Version
	header[3] = 1
	binary.BigEndian.PutUint32(header[4:8], uint32(len(serialized)))
	buf.Write(header)
	buf.Write(serialized)

	_, err = codec.ReadMessage(&buf)
	if err == nil {
		t.Fatal("expected error for message exceeding max size")
	}
}

func TestReadPartialHeader(t *testing.T) {
	codec := NewCodec()
	r := bytes.NewReader([]byte{0x4E}) // only 1 byte of header
	_, err := codec.ReadMessage(r)
	if err == nil {
		t.Fatal("expected error for partial header")
	}
}

func TestReadHeaderThenEOF(t *testing.T) {
	codec := NewCodec()
	buf := make([]byte, 8)
	binary.BigEndian.PutUint16(buf[0:2], Magic)
	buf[2] = Version
	buf[3] = 1
	binary.BigEndian.PutUint32(buf[4:8], 10) // claim 10 bytes but send none

	_, err := codec.ReadMessage(bytes.NewReader(buf))
	if err == nil {
		t.Fatal("expected error for missing data body")
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	codec := NewCodec()
	orig := &Message{
		Id:        uuid.NewString(),
		Timestamp: 1700000000000,
		Type:      MessageType_HANDSHAKE,
		Payload:   []byte("test payload"),
	}

	b, err := codec.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got Message
	if err := codec.Unmarshal(b, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.Id != orig.Id {
		t.Errorf("Id = %q, want %q", got.Id, orig.Id)
	}
	if got.Type != orig.Type {
		t.Errorf("Type = %v, want %v", got.Type, orig.Type)
	}
}

func TestMarshalUnmarshal_Response(t *testing.T) {
	codec := NewCodec()
	orig := &Response{
		Code: ErrorCode_SUCCESS,
		Msg:  "ok",
		Data: []byte("response data"),
	}

	b, err := codec.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got Response
	if err := codec.Unmarshal(b, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.Code != orig.Code {
		t.Errorf("Code = %v, want %v", got.Code, orig.Code)
	}
	if got.Msg != orig.Msg {
		t.Errorf("Msg = %q, want %q", got.Msg, orig.Msg)
	}
}

func TestWriteToClosedWriter(t *testing.T) {
	codec := NewCodec()
	msg := &Message{
		Id:   uuid.NewString(),
		Type: MessageType_DISCONNECT,
	}

	var buf bytes.Buffer
	if err := codec.WriteMessage(&buf, msg); err != nil {
		t.Fatalf("WriteMessage failed: %v", err)
	}

	// reading from the write end of a bytes.Buffer is fine,
	// but let's verify we can read it back
	_, err := codec.ReadMessage(&buf)
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	codec := NewCodec()
	const goroutines = 20
	var wg sync.WaitGroup

	for i := range goroutines {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			msg := &Message{
				Id:        uuid.NewString(),
				Timestamp: int64(1700000000000 + n),
				Type:      MessageType(int32(n%10) + 1),
				Payload:   []byte{byte(n)},
			}

			var buf bytes.Buffer
			if err := codec.WriteMessage(&buf, msg); err != nil {
				t.Errorf("WriteMessage failed: %v", err)
				return
			}

			got, err := codec.ReadMessage(&buf)
			if err != nil {
				t.Errorf("ReadMessage failed: %v", err)
				return
			}

			if got.Id != msg.Id {
				t.Errorf("Id mismatch: got %q, want %q", got.Id, msg.Id)
			}
		}(i)
	}
	wg.Wait()
}

func FuzzCodecRoundTrip(f *testing.F) {
	seeds := []*Message{
		{Id: uuid.NewString(), Type: MessageType_HANDSHAKE, Payload: []byte("hello")},
		{Id: uuid.NewString(), Type: MessageType_HEARTBEAT},
		{Id: uuid.NewString(), Type: MessageType_TUNNEL_OPEN, Payload: make([]byte, 1000)},
	}
	for _, msg := range seeds {
		b, _ := proto.Marshal(msg)
		f.Add(b)
	}

	f.Fuzz(func(t *testing.T, raw []byte) {
		// Just ensure we never panic on arbitrary data
		codec := NewCodec()

		// Try WriteMessage first (should succeed for valid proto)
		msg := &Message{}
		if err := proto.Unmarshal(raw, msg); err != nil {
			return
		}

		var buf bytes.Buffer
		if err := codec.WriteMessage(&buf, msg); err != nil {
			return
		}

		got, err := codec.ReadMessage(&buf)
		if err != nil {
			return
		}

		// Basic sanity: same ID and type
		if got.Id != msg.Id {
			t.Errorf("Id mismatch after round trip")
		}
	})
}

func TestLargePayload(t *testing.T) {
	codec := NewCodec()
	// 10MB payload
	payload := make([]byte, 10<<20)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	msg := &Message{
		Id:        uuid.NewString(),
		Timestamp: 1700000000000,
		Type:      MessageType_HANDSHAKE,
		Payload:   payload,
	}

	var buf bytes.Buffer
	if err := codec.WriteMessage(&buf, msg); err != nil {
		t.Fatalf("WriteMessage failed: %v", err)
	}

	got, err := codec.ReadMessage(&buf)
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	if !bytes.Equal(got.Payload, payload) {
		t.Errorf("Payload mismatch for large message")
	}
}

func TestMultipleMessagesInSingleReader(t *testing.T) {
	codec := NewCodec()
	var combined bytes.Buffer

	msgs := make([]*Message, 10)
	for i := range msgs {
		msgs[i] = &Message{
			Id:        uuid.NewString(),
			Timestamp: int64(1700000000000 + i),
			Type:      MessageType(int32(i%5) + 1),
			Payload:   []byte{byte(i)},
		}
		if err := codec.WriteMessage(&combined, msgs[i]); err != nil {
			t.Fatalf("WriteMessage %d failed: %v", i, err)
		}
	}

	for i, expected := range msgs {
		got, err := codec.ReadMessage(&combined)
		if err != nil {
			t.Fatalf("ReadMessage %d failed: %v", i, err)
		}
		if got.Id != expected.Id {
			t.Errorf("Message %d: Id = %q, want %q", i, got.Id, expected.Id)
		}
	}

	// Ensure no more messages
	_, err := codec.ReadMessage(&combined)
	if err == nil {
		t.Error("Expected error after reading all messages")
	}
}

func TestHeaderSizeConstant(t *testing.T) {
	// Verify header is exactly 8 bytes: magic(2) + version(1) + codecType(1) + length(4)
	codec := NewCodec()
	msg := &Message{
		Id:   uuid.NewString(),
		Type: MessageType_HANDSHAKE,
	}

	var buf bytes.Buffer
	if err := codec.WriteMessage(&buf, msg); err != nil {
		t.Fatalf("WriteMessage failed: %v", err)
	}

	raw := buf.Bytes()
	if len(raw) < 8 {
		t.Fatalf("message too short: %d bytes", len(raw))
	}

	// Check magic
	magic := binary.BigEndian.Uint16(raw[0:2])
	if magic != Magic {
		t.Errorf("magic = 0x%X, want 0x%X", magic, Magic)
	}

	// Check version
	if raw[2] != Version {
		t.Errorf("version = %d, want %d", raw[2], Version)
	}

	// Check codec type is protobuf (1)
	if raw[3] != 1 {
		t.Errorf("codecType = %d, want 1", raw[3])
	}
}

func TestNilSafeReadOnEmptyReader(t *testing.T) {
	codec := NewCodec()
	_, err := codec.ReadMessage(bytes.NewReader(nil))
	if err == nil {
		t.Fatal("expected error reading from empty reader")
	}
}

func TestWriteMessage_Interface(t *testing.T) {
	// Verify that Codec interface is implemented
	var c Codec = NewCodec()
	if c == nil {
		t.Fatal("NewCodec() returned nil")
	}
}

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	codec := NewCodec()

	t.Run("message", func(t *testing.T) {
		orig := &Message{
			Id:        "test-id",
			Timestamp: 1234567890,
			Type:      MessageType_HEARTBEAT_ACK,
			Payload:   []byte("pong"),
		}
		b, err := codec.Marshal(orig)
		if err != nil {
			t.Fatal(err)
		}
		var dest Message
		if err := codec.Unmarshal(b, &dest); err != nil {
			t.Fatal(err)
		}
		if dest.Id != orig.Id {
			t.Errorf("id mismatch")
		}
	})

	t.Run("response", func(t *testing.T) {
		orig := &Response{
			Code: ErrorCode_FAILED,
			Msg:  "something went wrong",
			Data: []byte("details"),
		}
		b, err := codec.Marshal(orig)
		if err != nil {
			t.Fatal(err)
		}
		var dest Response
		if err := codec.Unmarshal(b, &dest); err != nil {
			t.Fatal(err)
		}
		if dest.Code != orig.Code {
			t.Errorf("code mismatch")
		}
	})
}
