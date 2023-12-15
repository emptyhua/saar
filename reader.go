package saar

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"sync"
)

type FileReader struct {
	*io.SectionReader
	Header *Header
}

type Reader struct {
	r     io.ReadSeeker
	lock  sync.Mutex
	index []*Header
}

func NewReader(r io.ReadSeeker) *Reader {
	return &Reader{
		r: r,
	}
}

func (sr *Reader) ReadAt(p []byte, off int64) (n int, err error) {
	sr.lock.Lock()
	defer sr.lock.Unlock()

	_, err = sr.r.Seek(off, io.SeekStart)
	if err != nil {
		return
	}

	n, err = sr.r.Read(p)

	return
}

func (sr *Reader) initIndex() error {
	if sr.index != nil {
		return nil
	}

	if _, err := sr.r.Seek(0, io.SeekStart); err != nil {
		return err
	}

	tmp := make([]byte, 4)

	if _, err := sr.r.Read(tmp); err != nil {
		return err
	}

	off := binary.LittleEndian.Uint32(tmp)

	if _, err := sr.r.Seek(int64(off), io.SeekStart); err != nil {
		return fmt.Errorf("can't seek to header index:%w", err)
	}

	bs, err := ioutil.ReadAll(sr.r)
	if err != nil {
		return err
	}

	index := []*Header{}
	if err := json.Unmarshal(bs, &index); err != nil {
		return fmt.Errorf("parse header index failed:%w", err)
	}
	sr.index = index

	return nil
}

func (sr *Reader) Open(p string) (*FileReader, error) {
	sr.lock.Lock()
	defer sr.lock.Unlock()

	if err := sr.initIndex(); err != nil {
		return nil, err
	}

	var hdr *Header
	for _, oh := range sr.index {
		if oh.Path == p {
			hdr = oh
			break
		}
	}
	if hdr == nil {
		return nil, fmt.Errorf("%w:%s", ErrFileNotExist, p)
	}

	if hdr.IsDir {
		return nil, fmt.Errorf("%w:%s", ErrIsDir, p)
	}

	return &FileReader{
		SectionReader: io.NewSectionReader(sr, hdr.Offset, hdr.Size),
		Header:        hdr,
	}, nil
}

func (sr *Reader) List() ([]Header, error) {
	sr.lock.Lock()
	defer sr.lock.Unlock()

	if err := sr.initIndex(); err != nil {
		return nil, err
	}

	arr := make([]Header, 0, len(sr.index))
	for _, hdr := range sr.index {
		arr = append(arr, *hdr)
	}

	return arr, nil
}

func (sr *Reader) Close() error {
	if c, ok := sr.r.(io.Closer); ok {
		return c.Close()
	}

	return nil
}
