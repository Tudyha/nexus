package io

import (
	"io"
	"sync"
)

const defaultBufSize = 128 * 1024

// Copy 使用默认缓冲区大小（128KB）双向拷贝数据。
func Copy(src io.ReadWriteCloser, dst io.ReadWriteCloser) {
	copyWithBuf(src, dst, defaultBufSize)
}

// CopyBuf 使用指定缓冲区大小双向拷贝数据。
func CopyBuf(src io.ReadWriteCloser, dst io.ReadWriteCloser, bufSize int) {
	copyWithBuf(src, dst, bufSize)
}

func copyWithBuf(src io.ReadWriteCloser, dst io.ReadWriteCloser, bufSize int) {
	var (
		wg        sync.WaitGroup
		closeOnce [2]sync.Once
	)

	closeSrc := func() { closeOnce[0].Do(func() { src.Close() }) }
	closeDst := func() { closeOnce[1].Do(func() { dst.Close() }) }

	wg.Add(2)

	go func() {
		defer wg.Done()
		io.CopyBuffer(dst, src, make([]byte, bufSize))
		if tc, ok := dst.(interface{ CloseWrite() error }); ok {
			tc.CloseWrite()
		} else {
			closeDst()
		}
	}()

	go func() {
		defer wg.Done()
		io.CopyBuffer(src, dst, make([]byte, bufSize))
		if tc, ok := src.(interface{ CloseWrite() error }); ok {
			tc.CloseWrite()
		} else {
			closeSrc()
		}
	}()

	wg.Wait()
	closeSrc()
	closeDst()
}
