// Copyright (c) 2021 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package bmppath

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/tunabay/go-bitarray"
)

// ErrInvalidWidth is the error thrown when the specified width is invalid.
var ErrInvalidWidth = errors.New("invalid width")

// ErrInvalidBitmap is the error thrown when the source bitmap data is invalid.
var ErrInvalidBitmap = errors.New("invalid bitmap")

// Vertex represents the coordinate of one of the vertices that make up the
// polyline path.
type Vertex [2]int

// String returns the string representation of a Vertex in "(x, y)" format.
func (v Vertex) String() string { return fmt.Sprintf("(%d, %d)", v[0], v[1]) }

// X returns the x-coordinate of Vertex.
func (v Vertex) X() int { return v[0] }

// Y returns the y-coordinate of Vertex.
func (v Vertex) Y() int { return v[1] }

// Path is a bitmap image represented by a set of closed paths.
type Path struct {
	Width, Height int
	Vertices      [][]Vertex
}

// NumPath returns the number of closed paths in this set of paths.
func (p *Path) NumPath() int { return len(p.Vertices) }

// PathLen returns the number of vertices of the closed path specified by the
// index n. n == 0 points to the first path, and n must be less than p.NumPath()
// otherwise it panics.
func (p *Path) PathLen(n int) int { return len(p.Vertices[n]) }

// PathString returns the string representation of the closed path specified by
// the index n.
func (p *Path) PathString(n int) string {
	f := make([]string, len(p.Vertices[n]))
	for i, v := range p.Vertices[n] {
		f[i] = v.String()
	}
	return strings.Join(f, ", ")
}

// SVGDString is identical to WriteSVGD except that it returns a string instead
// of writing to io.Writer.
func (p *Path) SVGDString() string {
	var sb strings.Builder
	_ = p.WriteSVGD(&sb)
	return sb.String()
}

// WriteSVGD converts the entire set of paths into a string representation that
// can be used as the 'd' property of the SVG <path> element, and writes it to
// w. The 'd' string written uses only relative coordinates (such as m instead
// of M) relative to the upper left corner of the source image. Therefore, it is
// possible to translate whole path by prepending commands before the 'd'
// string written.
func (p *Path) WriteSVGD(w io.Writer) error {
	var c Vertex
	for _, p := range p.Vertices {
		if err := pathSVGD(w, p, c); err != nil {
			return err
		}
		c = p[0]
	}
	return nil
}

// WriteSVG writes the vectorized bitmap image as an SVG document. It is
// recommended to use WriteSVGD instead to write customized SVG documents.
func (p *Path) WriteSVG(w io.Writer) error {
	var sb strings.Builder
	fmt.Fprintln(&sb, `<?xml version="1.0" encoding="utf-8"?>`)
	fmt.Fprint(&sb, `<svg version="1.1" xmlns="http://www.w3.org/2000/svg"`)
	// fmt.Fprintf(&sb, ` x="0px" y="0px" width="%dpx" height="%dpx"`, p.Width, p.Height)
	fmt.Fprintf(&sb, ` viewBox="0 0 %d %d">`, p.Width, p.Height)
	fmt.Fprintln(&sb)
	// fmt.Fprintf(&sb, `<rect fill="#fff" width="%d" height="%d"/>`, p.Width, p.Height)
	fmt.Fprintf(&sb, `<path fill="#fff" d="m0,0h%dv%dh-%dz"/>`, p.Width, p.Height, p.Width)
	fmt.Fprint(&sb, `<path d="`)
	if _, err := fmt.Fprint(w, sb.String()); err != nil {
		return fmt.Errorf("write failure: %w", err)
	}
	if err := p.WriteSVGD(w); err != nil {
		return err
	}
	sb.Reset()
	fmt.Fprintln(&sb, `"/>`)
	fmt.Fprintln(&sb, `</svg>`)
	if _, err := fmt.Fprint(w, sb.String()); err != nil {
		return fmt.Errorf("write failure: %w", err)
	}
	return nil
}

func pathSVGD(w io.Writer, p []Vertex, z Vertex) error {
	c := p[0]
	d := Vertex{c[0] - z[0], c[1] - z[1]}
	cm := ","
	if d[1] < 0 {
		cm = ""
	}
	if _, err := fmt.Fprintf(w, "m%d%s%d", d[0], cm, d[1]); err != nil {
		return fmt.Errorf("write failure: %w", err)
	}
	for i := 1; i < len(p); i++ {
		v := p[i]
		switch {
		case v[0] == c[0]:
			if _, err := fmt.Fprintf(w, "v%d", v[1]-c[1]); err != nil {
				return fmt.Errorf("write failure: %w", err)
			}
		case v[1] == c[1]:
			if _, err := fmt.Fprintf(w, "h%d", v[0]-c[0]); err != nil {
				return fmt.Errorf("write failure: %w", err)
			}
		}
		c = v
	}
	if _, err := fmt.Fprint(w, "z"); err != nil {
		return fmt.Errorf("write failure: %w", err)
	}
	return nil
}

type vertex struct {
	x, y       int
	prev, next *vertex
}

func (v *vertex) ins(chead *vertex) {
	ctail := chead.prev
	v.prev.next = chead
	chead.prev, ctail.next = v.prev, v
	v.prev = ctail
}

type path struct {
	head, tail *vertex
	nVertices  int
	deleted    bool
}

// dist finds the closest pair of vertices on the two paths p0 and p1, and
// returns the distance between them and the pair of vertices. If p0 and p1
// touch at a vertex, the distance 0 and that vertex are returned.
func dist(p0, p1 *path) (int, *vertex, *vertex) {
	mind := 0
	var nv0, nv1 *vertex
	for v0 := p0.head; ; {
		for v1 := p1.head; ; {
			dx, dy := v1.x-v0.x, v1.y-v0.y
			d := dx*dx + dy*dy
			if nv0 == nil || d < mind {
				if d == 0 {
					return 0, v0, v1
				}
				mind, nv0, nv1 = d, v0, v1
			}
			v1 = v1.next
			if v1 == p1.head {
				break
			}
		}
		if v0 = v0.next; v0 == p0.head {
			break
		}
	}

	return mind, nv0, nv1
}

func newPath(v Vertex) *path {
	v0 := &vertex{x: v[0], y: v[1]}
	return &path{head: v0, tail: v0, nVertices: 1}
}

func (p *path) addVertex(x, y int) {
	p.tail.next = &vertex{x: x, y: y, prev: p.tail}
	p.tail = p.tail.next
	p.nVertices++
}

func (p *path) close() {
	p.tail.next = p.head
	p.head.prev = p.tail
}

func (p *path) normalize() {
	mind := 0
	var minv *vertex
	v := p.head
	for {
		d := v.x*v.x + v.y*v.y
		if minv == nil || d < mind {
			mind, minv = d, v
			if d == 0 {
				break
			}
		}
		v = v.next
		if v == p.head {
			break
		}
	}
	p.head, p.tail = minv, minv.prev
}

func (p *path) pub() []Vertex {
	var ret []Vertex
	v := p.head
	for {
		ret = append(ret, Vertex{v.x, v.y})
		v = v.next
		if v == p.head {
			break
		}
	}
	return ret
}

type pathSet struct {
	width, height int
	paths         []*path
}

type pathList []*path

func (pl pathList) Len() int           { return len(pl) }
func (pl pathList) Less(i, j int) bool { return pl[i].nVertices > pl[j].nVertices }
func (pl pathList) Swap(i, j int)      { pl[i], pl[j] = pl[j], pl[i] }

func (ps *pathSet) addPath(p *path) { ps.paths = append(ps.paths, p) }

func (ps *pathSet) sort() {
	var n, x0, y0 int
	for _, p := range ps.paths {
		if !p.deleted {
			n++
			p.normalize()
		}
	}
	a := make([]*path, 0, n)
	for i := 0; i < n; i++ {
		mind, mini := 0, -1
		for j, p := range ps.paths {
			if p == nil || p.deleted {
				continue
			}
			dx, dy := p.head.x-x0, p.head.y-y0
			if d := dx*dx + dy*dy; mini == -1 || d < mind {
				mind, mini = d, j
			}
		}
		p := ps.paths[mini]
		ps.paths[mini] = nil
		x0, y0 = p.head.x, p.head.y
		a = append(a, p)
	}
	ps.paths = a
}

func (ps *pathSet) pub() [][]Vertex {
	ret := make([][]Vertex, 0, len(ps.paths))
	for _, p := range ps.paths {
		// if p.deleted {
		// 	continue
		// }
		ret = append(ret, p.pub())
	}
	return ret
}

// New creates a set of paths from a binary bitmap image represented by a bit
// array. The bit array bm must be exactly width * height length.
func New(bm *bitarray.Buffer, width int) (*Path, error) {
	switch {
	case width < 1:
		return nil, fmt.Errorf("%w: %d < 1", ErrInvalidWidth, width)
	case bm == nil:
		return nil, fmt.Errorf("%w: bm == nil", ErrInvalidBitmap)
	}
	bmlen := bm.Len()
	height := bmlen / width
	switch {
	case bmlen < width:
		return nil, fmt.Errorf("%w: too short: len=%d < width=%d", ErrInvalidBitmap, bmlen, width)
	case bmlen%width != 0:
		return nil, fmt.Errorf("%w: len=%d %% width=%d != 0", ErrInvalidBitmap, bmlen, width)
	}
	ps := &pathSet{width: width, height: height}

	v := bitarray.NewBuffer((width + 1) * (height + 1) << 2)
	pix := func(x, y int) bool {
		// if x < 0 || y < 0 || width <= x || height <= y {
		// 	return false
		// }
		return bm.BitAt(width*y+x) != 0
	}
	set := func(x, y, dir int, b bool) {
		bb := byte(0)
		if b {
			bb = 1
		}
		v.PutBitAt(((width+1)*y+x)*4+dir, bb)
	}
	get := func(x, y, dir int) bool {
		ret := v.BitAt(((width+1)*y+x)*4+dir) != 0
		if ret {
			set(x, y, dir, false)
		}
		return ret
	}
	for y := 0; y < height+1; y++ {
		for x := 0; x < width; x++ {
			var mu, md bool
			if 0 < y {
				mu = pix(x, y-1)
			}
			if y < height {
				md = pix(x, y)
			}
			switch {
			case !mu && md:
				set(x, y, 1, true)
			case mu && !md:
				set(x+1, y, 3, true)
			}
		}
	}
	for x := 0; x < width+1; x++ {
		for y := 0; y < height; y++ {
			var ml, mr bool
			if 0 < x {
				ml = pix(x-1, y)
			}
			if x < width {
				mr = pix(x, y)
			}
			switch {
			case !ml && mr:
				set(x, y+1, 0, true)
			case ml && !mr:
				set(x, y, 2, true)
			}
		}
	}
	for {
		s := Vertex{-1, -1}
		for y := 0; y < height+1 && s[1] < 0; y++ {
			for x := 0; x < width+1 && s[0] < 0; x++ {
				if get(x, y, 1) {
					s = Vertex{x, y}
				}
			}
		}
		if s[0] < 0 {
			break
		}
		path := newPath(s)
		dir, cx, cy := 1, s[0]+1, s[1]
		for cx != s[0] || cy != s[1] {
			switch dir {
			case 0:
				switch {
				case get(cx, cy, 3):
					path.addVertex(cx, cy)
					dir = 3
					cx--
				case get(cx, cy, 1):
					path.addVertex(cx, cy)
					dir = 1
					cx++
				case get(cx, cy, 0):
					cy--
				}
			case 1:
				switch {
				case get(cx, cy, 0):
					path.addVertex(cx, cy)
					dir = 0
					cy--
				case get(cx, cy, 2):
					path.addVertex(cx, cy)
					dir = 2
					cy++
				case get(cx, cy, 1):
					cx++
				}
			case 2:
				switch {
				case get(cx, cy, 1):
					path.addVertex(cx, cy)
					dir = 1
					cx++
				case get(cx, cy, 3):
					path.addVertex(cx, cy)
					dir = 3
					cx--
				case get(cx, cy, 2):
					cy++
				}
			case 3:
				switch {
				case get(cx, cy, 2):
					path.addVertex(cx, cy)
					dir = 2
					cy++
				case get(cx, cy, 0):
					path.addVertex(cx, cy)
					dir = 0
					cy--
				case get(cx, cy, 3):
					cx--
				}
			}
		}
		path.close()
		ps.addPath(path)
	}
	sort.Sort(pathList(ps.paths))
	for {
		eff := false
		for i, p0 := range ps.paths {
			if p0.deleted {
				continue
			}
			for j := i + 1; j < len(ps.paths); j++ {
				p1 := ps.paths[j]
				if p1.deleted {
					continue
				}
				if d, nv0, nv1 := dist(p0, p1); d == 0 {
					nv0.ins(nv1)
					p1.deleted = true
					eff = true
				}
			}
		}
		if !eff {
			break
		}
	}
	ps.sort()

	ret := &Path{
		Width:    ps.width,
		Height:   ps.height,
		Vertices: ps.pub(),
	}

	return ret, nil
}
