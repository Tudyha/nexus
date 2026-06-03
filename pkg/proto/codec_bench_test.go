package proto

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"
)

func BenchmarkCodecWriteMessage(b *testing.B) {
	codec := NewCodec()
	msg := &Message{
		Id:        "benchmark-id-001",
		Timestamp: 1700000000000,
		Type:      MessageType_HANDSHAKE,
		Payload:   make([]byte, 1024),
	}
	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := codec.WriteMessage(&buf, msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCodecReadMessage(b *testing.B) {
	codec := NewCodec()
	msg := &Message{
		Id:        "benchmark-id-001",
		Timestamp: 1700000000000,
		Type:      MessageType_HANDSHAKE,
		Payload:   make([]byte, 1024),
	}
	var buf bytes.Buffer
	if err := codec.WriteMessage(&buf, msg); err != nil {
		b.Fatal(err)
	}
	data := buf.Bytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(data)
		if _, err := codec.ReadMessage(r); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCodecRoundTrip(b *testing.B) {
	codec := NewCodec()
	msg := &Message{
		Id:        "benchmark-id-001",
		Timestamp: 1700000000000,
		Type:      MessageType_TUNNEL_OPEN,
		Payload:   make([]byte, 4096),
	}
	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := codec.WriteMessage(&buf, msg); err != nil {
			b.Fatal(err)
		}
		if _, err := codec.ReadMessage(&buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCodecMarshal(b *testing.B) {
	codec := NewCodec()
	msg := &Message{
		Id:        "benchmark-id-001",
		Timestamp: 1700000000000,
		Type:      MessageType_HEARTBEAT,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := codec.Marshal(msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCodecLargePayload(b *testing.B) {
	codec := NewCodec()
	payload := make([]byte, 64*1024)
	rand.Read(payload)
	msg := &Message{
		Id:        "benchmark-large",
		Timestamp: 1700000000000,
		Type:      MessageType_TUNNEL_OPEN,
		Payload:   payload,
	}
	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := codec.WriteMessage(&buf, msg); err != nil {
			b.Fatal(err)
		}
		if _, err := codec.ReadMessage(&buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCodecConcurrent(b *testing.B) {
	codec := NewCodec()
	b.RunParallel(func(pb *testing.PB) {
		msg := &Message{
			Id:        "benchmark-concurrent",
			Timestamp: 1700000000000,
			Type:      MessageType_HEARTBEAT,
			Payload:   []byte("ping"),
		}
		var buf bytes.Buffer
		for pb.Next() {
			buf.Reset()
			if err := codec.WriteMessage(&buf, msg); err != nil {
				b.Error(err)
				return
			}
			if _, err := codec.ReadMessage(&buf); err != nil {
				b.Error(err)
				return
			}
		}
	})
}

func BenchmarkCodecPayloadSizes(b *testing.B) {
	sizes := []int{64, 256, 1024, 4096, 16384, 65536}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("payload-%d", size), func(b *testing.B) {
			codec := NewCodec()
			payload := make([]byte, size)
			rand.Read(payload)
			msg := &Message{
				Id:        "benchmark-size",
				Timestamp: 1700000000000,
				Type:      MessageType_TUNNEL_OPEN,
				Payload:   payload,
			}
			var buf bytes.Buffer
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf.Reset()
				if err := codec.WriteMessage(&buf, msg); err != nil {
					b.Fatal(err)
				}
				if _, err := codec.ReadMessage(&buf); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
