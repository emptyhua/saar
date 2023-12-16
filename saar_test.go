package saar

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

// common read write
func TestReadWrite(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	arw := NewWriter(buf)

	_, err := arw.Write([]byte("ccc"))
	assert.ErrorIs(t, err, ErrHeaderRequired)

	err = arw.WriteHeader(Header{Path: "1.txt"})
	assert.Nil(t, err)

	err = arw.WriteHeader(Header{Path: "2.txt"})
	assert.Nil(t, err)

	err = arw.WriteHeader(Header{Path: "1.txt"})
	assert.ErrorIs(t, err, ErrExist)

	_, err = arw.Write([]byte("222"))
	assert.Nil(t, err)

	err = arw.Close()
	assert.Nil(t, err)

	ard := NewReader(bytes.NewReader(buf.Bytes()))

	rd1, err := ard.Open("1.txt")
	assert.Nil(t, err)
	bs1, err := ioutil.ReadAll(rd1)
	assert.Nil(t, err)
	assert.Equal(t, bs1, []byte(""))

	rd2, err := ard.Open("2.txt")
	assert.Nil(t, err)
	bs2, err := ioutil.ReadAll(rd2)
	assert.Nil(t, err)
	assert.Equal(t, bs2, []byte("222"))

	_, err = ard.Open("3.txt")
	assert.ErrorIs(t, err, ErrNotExist)
}

func TestEmptyArchive(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	arw := NewWriter(buf)
	err := arw.Close()
	assert.Nil(t, err)
	ard := NewReader(bytes.NewReader(buf.Bytes()))
	_, err = ard.Open("fdfdf")
	assert.ErrorIs(t, err, ErrNotExist)
}

// read malformed file
func TestReadMalformed(t *testing.T) {
	bufrd1 := bytes.NewReader([]byte(""))
	arrd1 := NewReader(bufrd1)
	_, err := arrd1.Open("dddd")
	t.Log("read malformed 1", err)
	assert.ErrorIs(t, err, ErrMalformed)

	bufrd2 := bytes.NewReader([]byte("12"))
	arrd2 := NewReader(bufrd2)
	_, err = arrd2.Open("dddd")
	t.Log("read malformed 2", err)
	assert.ErrorIs(t, err, ErrMalformed)

	bufrd3 := bytes.NewReader([]byte("1234"))
	arrd3 := NewReader(bufrd3)
	_, err = arrd3.Open("dddd")
	t.Log("read malformed 3", err)
	assert.ErrorIs(t, err, ErrMalformed)

	bufrd4 := bytes.NewReader([]byte("12340001"))
	arrd4 := NewReader(bufrd4)
	_, err = arrd4.Open("dddd")
	t.Log("read malformed 4", err)
	assert.ErrorIs(t, err, ErrMalformed)
}
