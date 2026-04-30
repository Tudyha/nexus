package io

import (
	"io"
	"sync"
)

func Copy(src io.ReadWriteCloser, dst io.ReadWriteCloser) {
	var (
		wg        sync.WaitGroup
		closeOnce [2]sync.Once
	)

	closeSrc := func() { closeOnce[0].Do(func() { src.Close() }) }
	closeDst := func() { closeOnce[1].Do(func() { dst.Close() }) }

	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(dst, src)
		if tc, ok := dst.(interface{ CloseWrite() error }); ok {
			tc.CloseWrite()
		} else {
			closeDst()
		}
	}()

	go func() {
		defer wg.Done()
		io.Copy(src, dst)
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
