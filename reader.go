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

	if _, err := sr.r.Seek(-4, io.SeekEnd); err != nil {
		return fmt.Errorf("%w:seek to index size failed:%v", ErrMalformed, err)
	}

	tmp := make([]byte, 4)

	if _, err := sr.r.Read(tmp); err != nil {
		return err
	}

	size := int64(binary.LittleEndian.Uint32(tmp))

	if _, err := sr.r.Seek(-4-size, io.SeekEnd); err != nil {
		return fmt.Errorf("%w:seek to index failed:%v", ErrMalformed, err)
	}

	bs, err := ioutil.ReadAll(io.LimitReader(sr.r, size))
	if err != nil {
		return fmt.Errorf("%w:read index failed:%v", ErrMalformed, err)
	}

	index := []*Header{}
	if err := json.Unmarshal(bs, &index); err != nil {
		return fmt.Errorf("%w:parse index failed:%v", ErrMalformed, err)
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
		return nil, fmt.Errorf("%w:%s", ErrNotExist, p)
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
