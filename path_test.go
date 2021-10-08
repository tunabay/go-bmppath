// Copyright (c) 2021 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package bmppath_test

import (
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/tunabay/go-bitarray"
	"github.com/tunabay/go-bmppath"
)

func TestNew_error(t *testing.T) {
	buf := bitarray.NewBuffer(100)
	var nilbuf *bitarray.Buffer

	if path, err := bmppath.New(buf, 0); !errors.Is(err, bmppath.ErrInvalidWidth) {
		t.Errorf("width=0: unexpected return: path=%+v, err=%+v", path, err)
	}
	if path, err := bmppath.New(buf, -1); !errors.Is(err, bmppath.ErrInvalidWidth) {
		t.Errorf("width=-1: unexpected return: path=%+v, err=%+v", path, err)
	}
	if path, err := bmppath.New(nilbuf, 8); !errors.Is(err, bmppath.ErrInvalidBitmap) {
		t.Errorf("bmp=nil: unexpected return: path=%+v, err=%+v", path, err)
	}
	if path, err := bmppath.New(buf.Slice(0, 7), 8); !errors.Is(err, bmppath.ErrInvalidBitmap) {
		t.Errorf("7 < 8: unexpected return: path=%+v, err=%+v", path, err)
	}
	if path, err := bmppath.New(buf.Slice(0, 0), 8); !errors.Is(err, bmppath.ErrInvalidBitmap) {
		t.Errorf("0 < 8: unexpected return: path=%+v, err=%+v", path, err)
	}
	if path, err := bmppath.New(buf.Slice(0, 15), 8); !errors.Is(err, bmppath.ErrInvalidBitmap) {
		t.Errorf("15 %% 8 != 0: unexpected return: path=%+v, err=%+v", path, err)
	}
	if path, err := bmppath.New(buf.Slice(0, 17), 8); !errors.Is(err, bmppath.ErrInvalidBitmap) {
		t.Errorf("17 %% 8 != 0: unexpected return: path=%+v, err=%+v", path, err)
	}
}

func Test_svgFile(t *testing.T) {
	outPath := filepath.Join(os.TempDir(), "go-bmppath-test-out.svg")
	const width = 37
	h := "00000000000000000000000000000000000000fed4abf804" +
		"119090402e8c4eba0174b0b5d00babd3ae804154550403fa" +
		"aaafe000019000007f4c31880086d63f4020deeb90008123" +
		"4e20053fbd5380109aaf6c031b5534000986af1900477688" +
		"6806a0a9ffc026d3e2f80171a607b00abb74ff80007f546c" +
		"03fb8cea001054931900baba2ff005d49cea002eaa9fe401" +
		"054d59a00fe1276e00000000000000000000000000000000" +
		"00000000"

	src, _ := hex.DecodeString(h)
	buf := bitarray.NewBufferFromByteSlice(src).Slice(0, width*width)
	path, err := bmppath.New(buf, width)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	f, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o0644)
	if err != nil {
		t.Fatalf("OpenFile(): %s: %v", outPath, err)
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(outPath)
	}()

	if err := path.WriteSVG(f); err != nil {
		t.Errorf("WriteSVG(): %s", err)
	}
}
