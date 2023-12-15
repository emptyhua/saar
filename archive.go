package saar

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type ProgressFunc func(p1 string, p2 string, err error)

func CreateArchive(progFunc ProgressFunc, dst string, srcs ...string) (rterr error) {
	dstFp, err := os.Create(dst)
	if err != nil {
		return err
	}

	dstAr := NewWriter(dstFp)

	defer func() {
		if err := dstAr.Close(); err != nil {
			if rterr == nil {
				rterr = err
			}
		}
	}()

	af := func(p string, rp string, isDir bool) error {
		hdr := Header{Path: rp, IsDir: isDir}

		if err := dstAr.WriteHeader(hdr); err != nil {
			return err
		}

		if isDir {
			return nil
		}

		rfp, err := os.Open(p)
		if err != nil {
			return err
		}
		defer rfp.Close()

		_, err = io.Copy(dstAr, rfp)
		return err
	}

	for _, src := range srcs {
		src, err = filepath.Abs(src)
		if err != nil {
			return err
		}

		err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("%s:%w", p, err)
			}

			rp, err := filepath.Rel(src, p)
			if err != nil {
				return err
			}

			rp = filepath.ToSlash(rp)

			if rp == "." {
				return nil
			}

			err = af(p, rp, d.IsDir())
			progFunc(p, rp, err)

			return err
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func ExtractArchive(progFunc ProgressFunc, src string, dst string) error {
	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}

	srcFp, err := os.Open(src)
	if err != nil {
		return err
	}

	srcAr := NewReader(srcFp)
	defer srcAr.Close()

	hdrs, err := srcAr.List()
	if err != nil {
		return err
	}

	xf := func(hdr Header, lp string) error {
		if hdr.IsDir {
			err := os.MkdirAll(lp, 0755)
			if err != nil {
				return err
			}
		} else {
			lfp, err := os.Create(lp)
			if err != nil {
				return err
			}
			defer lfp.Close()

			afp, err := srcAr.Open(hdr.Path)
			if err != nil {
				return err
			}

			_, err = io.Copy(lfp, afp)
			return err
		}

		return nil
	}

	for _, hdr := range hdrs {
		lp := filepath.Join(dst, filepath.FromSlash(hdr.Path))
		err := xf(hdr, lp)
		progFunc(hdr.Path, lp, err)
		if err != nil {
			return err
		}
	}

	return nil
}
