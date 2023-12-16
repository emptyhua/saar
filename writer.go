package saar

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

type Writer struct {
	w      io.Writer
	lock   sync.Mutex
	closed bool
	err    error
	index  []*Header
	hdr    *Header
	offset int64
	size   int64
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

func (sw *Writer) saveLastHeader() {
	if sw.hdr == nil {
		return
	}

	sw.hdr.Offset = sw.offset - sw.size
	sw.hdr.Size = sw.size
	sw.index = append(sw.index, sw.hdr)
	sw.size = 0
	sw.hdr = nil
}

func (sw *Writer) WriteHeader(hdr Header) error {
	sw.lock.Lock()
	defer sw.lock.Unlock()

	if sw.closed {
		return ErrClosed
	}

	if sw.err != nil {
		return sw.err
	}

	for _, oh := range sw.index {
		if oh.Path == hdr.Path {
			return fmt.Errorf("%w:%s", ErrExist, hdr.Path)
		}
	}

	sw.saveLastHeader()

	sw.hdr = &hdr

	return nil
}

func (sw *Writer) Write(p []byte) (n int, err error) {
	sw.lock.Lock()
	defer sw.lock.Unlock()

	if sw.closed {
		err = ErrClosed
		return
	}

	if sw.err != nil {
		err = sw.err
		return
	}

	if sw.hdr == nil {
		err = ErrHeaderRequired
		return
	}

	n, err = sw.w.Write(p)
	sw.offset += int64(n)
	sw.size += int64(n)

	if err != nil {
		sw.err = err
	}

	return
}

func (sw *Writer) writeIndex() error {
	jsbs, err := json.Marshal(sw.index)
	if err != nil {
		return err
	}

	if _, err := sw.w.Write(jsbs); err != nil {
		return err
	}

	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(len(jsbs)))

	if _, err := sw.w.Write(tmp); err != nil {
		return err
	}

	return nil
}

func (sw *Writer) Close() error {
	sw.lock.Lock()
	defer sw.lock.Unlock()

	if sw.closed {
		return ErrClosed
	}

	sw.closed = true

	if sw.err == nil {
		sw.saveLastHeader()

		if err := sw.writeIndex(); err != nil {
			return err
		}
	}

	if c, ok := sw.w.(io.Closer); ok {
		if err := c.Close(); err != nil {
			return err
		}
	}

	return nil
}
