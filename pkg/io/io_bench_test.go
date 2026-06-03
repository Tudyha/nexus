package io

import (
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"testing"
)

func BenchmarkCopy(b *testing.B) {
	sizes := []int{4096, 16384, 65536, 262144, 1048576}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			data := make([]byte, size)
			rand.Read(data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				srcR, srcW := net.Pipe()
				dstR, dstW := net.Pipe()
				go func() {
					srcW.Write(data)
					srcW.Close()
				}()
				go func() {
					io.ReadAll(dstR)
					dstR.Close()
				}()
				Copy(srcR, dstW)
			}
		})
	}
}

func BenchmarkCopyBuf(b *testing.B) {
	sizes := []int{4096, 16384, 65536}
	bufSizes := []int{4096, 32768, 262144}
	for _, size := range sizes {
		for _, bufSize := range bufSizes {
			b.Run(fmt.Sprintf("size-%d/buf-%d", size, bufSize), func(b *testing.B) {
				data := make([]byte, size)
				rand.Read(data)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					srcR, srcW := net.Pipe()
					dstR, dstW := net.Pipe()
					go func() {
						srcW.Write(data)
						srcW.Close()
					}()
					go func() {
						io.ReadAll(dstR)
						dstR.Close()
					}()
					CopyBuf(srcR, dstW, bufSize)
				}
			})
		}
	}
}
