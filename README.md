[![Go Reference](https://pkg.go.dev/badge/github.com/tunabay/go-bmppath.svg)](https://pkg.go.dev/github.com/tunabay/go-bmppath)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

# go-bmppath

## Overview

Package bmppath converts a monochrome 1-bit bitmap image into a set of vector
paths.

Note that this package is by no means a sophisticated tool for tracing raster
images in detail. It only has the ability to create vector paths with as few
paths as possible. The vector paths created keep the shape of the square pixels
and look jagged.

Originally this code was written for the purpose of converting a QR Code to a
print format.

### Examples

| source bitmap | paths |
|:---:|:---:|
|![](https://raw.githubusercontent.com/tunabay/go-bmppath/image/example-1-src.svg)|![](https://raw.githubusercontent.com/tunabay/go-bmppath/image/example-1-path.svg)|
|![](https://raw.githubusercontent.com/tunabay/go-bmppath/image/example-2-src.svg)|![](https://raw.githubusercontent.com/tunabay/go-bmppath/image/example-2-path.svg)|

Note: I draw these images manually to give an overview. Therefore, the details
may be slightly different from the actual paths output by the code. It's a known
mistake that multiple arrows are drawn on one path :-)

## Usage

```
import (
	"os"

	"github.com/tunabay/go-bitarray"
	"github.com/tunabay/go-bmppath"
)

func main() {
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
}
```
[Run in Go Playground](https://play.golang.org/p/aHcMc3Bd2Kg)

## Documentation

- https://pkg.go.dev/github.com/tunabay/go-bmppath

## License

go-bmppath is available under the MIT license. See the [LICENSE](LICENSE) file
for more information.
