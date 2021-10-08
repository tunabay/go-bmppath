// Copyright (c) 2021 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package bmppath_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/tunabay/go-bitarray"
	"github.com/tunabay/go-bmppath"
)

func Example_invaderSVG() {
	bmp := bitarray.NewBufferFromByteSlice([]byte{
		0b_10000001,
		0b_01000010,
		0b_00111100,
		0b_01111110,
		0b_11011011,
		0b_01111110,
		0b_00100100,
		0b_11000011,
	})

	path, err := bmppath.New(bmp, 8)
	if err != nil {
		panic(err)
	}

	_ = path.WriteSVG(os.Stdout)

	// Output:
	// <?xml version="1.0" encoding="utf-8"?>
	// <svg version="1.1" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 8 8">
	// <path fill="#fff" d="m0,0h8v8h-8z"/><path d="m0,0h1v1h1v1h4v-1h1v-1h1v1h-1v1h-1v1h1v1h1v1h-1v1h-1v1h2v1h-2v-1h-1v-1h-2v1h-1v1h-2v-1h2v-1h-1v-1h-1v-1h1v-1h1v-1h-1v-1h-1zm2,4v1h1v-1zm3,0v1h1v-1z"/>
	// </svg>
}

func Example_rawPathData() {
	bmp := bitarray.NewBufferFromBitArray(
		bitarray.MustParse(strings.Join([]string{
			"11101",
			"10100",
			"11101",
		}, "")),
	)
	path, err := bmppath.New(bmp, 5)
	if err != nil {
		panic(err)
	}

	fmt.Printf("size: %dx%d\n", path.Width, path.Height)
	for i, p := range path.Vertices {
		fmt.Printf("%d:", i)
		for _, v := range p {
			fmt.Printf(" (%d, %d)", v.X(), v.Y())
		}
		fmt.Print("\n")
	}

	// Output:
	// size: 5x3
	// 0: (0, 0) (3, 0) (3, 3) (0, 3)
	// 1: (1, 1) (1, 2) (2, 2) (2, 1)
	// 2: (4, 0) (5, 0) (5, 1) (4, 1)
	// 3: (4, 2) (5, 2) (5, 3) (4, 3)
}

func ExamplePath_PathString() {
	bmp := bitarray.NewBufferFromBitArray(
		bitarray.MustParse(strings.Join([]string{
			"1101",
			"1101",
			"0001",
			"1111",
		}, "")),
	)
	path, err := bmppath.New(bmp, 4)
	if err != nil {
		panic(err)
	}

	for i := 0; i < path.NumPath(); i++ {
		fmt.Printf("%d: n=%d: %s\n", i, path.PathLen(i), path.PathString(i))
	}

	// Output:
	// 0: n=4: (0, 0), (2, 0), (2, 2), (0, 2)
	// 1: n=6: (3, 0), (4, 0), (4, 4), (0, 4), (0, 3), (3, 3)
}

func ExamplePath_SVGDString() {
	bmp := bitarray.NewBufferFromBitArray(
		bitarray.MustParse(strings.Join([]string{
			"1101",
			"1101",
		}, "")),
	)
	path, err := bmppath.New(bmp, 4)
	if err != nil {
		panic(err)
	}

	fmt.Println(path.SVGDString())

	// Output:
	// m0,0h2v2h-2zm3,0h1v2h-1z
}
