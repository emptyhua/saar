package saar

import (
	"fmt"
)

type Header struct {
	Path   string
	IsDir  bool
	Offset int64
	Size   int64
	Extra  interface{}
}

var ErrFileNotExist = fmt.Errorf("saar:file not exist")
var ErrFileExist = fmt.Errorf("saar:file existed")
var ErrIsDir = fmt.Errorf("saar:can't open directory")
var ErrHeaderRequired = fmt.Errorf("WriteHeader() required")
var ErrClosed = fmt.Errorf("saar:file closed")
