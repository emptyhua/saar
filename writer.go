package saar

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

type Writer struct {
	w      io.WriteSeeker
	lock   sync.Mutex
	closed bool
	err    error
	index  []*Header
	hdr    *Header
	offset int64
	size   int64
}

func NewWriter(w io.WriteSeeker) *Writer {
	return &Writer{
		w: w,
	}
}

func (sw *Writer) initOffset() error {
	if sw.offset != 0 {
		return nil
	}

	if _, err := sw.w.Seek(0, io.SeekStart); err != nil {
		return err
	}

	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, 0)

	if _, err := sw.w.Write(tmp); err != nil {
		return err
	}

	sw.offset = 4
	return nil
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

func (sw *Writer) WriteHeader(hdr Header) (rterr error) {
	sw.lock.Lock()
	defer sw.lock.Unlock()

	if sw.closed {
		return ErrClosed
	}

	if sw.err != nil {
		return sw.err
	}

	defer func() {
		if rterr != nil {
			sw.err = rterr
		}
	}()

	if err := sw.initOffset(); err != nil {
		return err
	}

	for _, oh := range sw.index {
		if oh.Path == hdr.Path {
			return fmt.Errorf("%w:%s", ErrFileExist, hdr.Path)
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
	if err := sw.initOffset(); err != nil {
		return err
	}

	jsbs, err := json.Marshal(sw.index)
	if err != nil {
		return err
	}

	if _, err := sw.w.Write(jsbs); err != nil {
		return err
	}

	if _, err := sw.w.Seek(0, io.SeekStart); err != nil {
		return err
	}

	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(sw.offset))

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
