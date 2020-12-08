package geocoding

import (
	"fmt"
	"log"

	"bitbucket.org/kleinnic74/photos/domain/gps"
)

type entry struct {
	bounds gps.Rect
	data   interface{}
}

type ResultFunc func(interface{}, gps.Rect)

type quadtree struct {
	count  int
	Bounds gps.Rect

	root     *node
	capacity int
	maxDepth int
}

type node struct {
	bounds   gps.Rect
	quads    [4]*node
	entries  []entry
	capacity int
	depth    int
}

type Visitor interface {
	Begin(bounds gps.Rect)
	Level(depth int, bounds gps.Rect)
	Object(bounds gps.Rect)
	End()
}

func NewQuadTree(bounds gps.Rect) *quadtree {
	return &quadtree{Bounds: bounds, capacity: 20, maxDepth: 10}
}

func (qt *quadtree) InsertRect(r gps.Rect, o interface{}) {
	if !qt.Bounds.FullyContains(r) {
		panic(fmt.Sprintf("Rect %v not in vounds %v", r, qt.Bounds))
	}
	if qt.root == nil {
		qt.root = newNode(r, qt.capacity, qt.maxDepth)
	} else {
		if !qt.root.FullyContains(r) {
			qt.root = qt.root.grow(r)
		}
	}
	qt.root.add(r, o)
	qt.count++
}

func (qt *quadtree) Find(p gps.Point) (result []interface{}) {
	if qt.root == nil {
		return
	}
	qt.root.findFunc(p, func(o interface{}, _ gps.Rect) {
		result = append(result, o)
	})
	return
}

func (qt *quadtree) Visit(v Visitor) {
	v.Begin(qt.root.bounds)
	qt.root.visit(v)
	v.End()
}

func (qt *quadtree) FindFunc(p gps.Point, f ResultFunc) {
	if qt.root == nil {
		return
	}
	qt.root.findFunc(p, f)
	return
}

func newNode(bounds gps.Rect, capacity int, depth int) *node {
	log.Printf("New QuadTree node: %v", bounds)
	return &node{bounds: bounds, capacity: capacity, depth: depth}
}

func (n *node) FullyContains(r gps.Rect) bool {
	return n.bounds.FullyContains(r)
}

func (n *node) add(r gps.Rect, o interface{}) {
	e := entry{r, o}
	if n.quads[0] == nil {
		// Not subdivided yet
		if len(n.entries) < n.capacity || n.depth == 0 {
			n.entries = append(n.entries, e)
			return
		}
		n.split()
	}
	quad := n.choose(r)
	switch quad {
	case -1:
		n.entries = append(n.entries, e)
	default:
		n.quads[quad].add(r, o)
	}
}

func (n *node) split() {
	hw, hh := n.bounds.HalfSize()
	log.Printf("Splitting to [%f/%f]", n.bounds[0]+hw, n.bounds[1]+hh)
	n.quads[0] = newNode(gps.RectFrom(n.bounds[0], n.bounds[1], n.bounds[0]+hw, n.bounds[1]+hh), n.capacity, n.depth-1)
	n.quads[1] = newNode(gps.RectFrom(n.bounds[0], n.bounds[1]+hh, n.bounds[0]+hw, n.bounds[3]), n.capacity, n.depth-1)
	n.quads[2] = newNode(gps.RectFrom(n.bounds[0]+hw, n.bounds[1], n.bounds[2], n.bounds[1]+hh), n.capacity, n.depth-1)
	n.quads[3] = newNode(gps.RectFrom(n.bounds[0]+hw, n.bounds[1]+hh, n.bounds[2], n.bounds[3]), n.capacity, n.depth-1)
	entries := n.entries
	n.entries = nil
	for _, e := range entries {
		quad := n.choose(e.bounds)
		switch quad {
		case -1:
			// Does not fit in any quadrant
			n.entries = append(n.entries, e)
		default:
			n.quads[quad].add(e.bounds, e.data)
		}
	}
}

func (n *node) grow(r gps.Rect) *node {
	root := n
	for !root.FullyContains(r) {
		var xmin, ymin float64
		dx0, dx1 := root.bounds.X0()-r.X0(), r.X1()-root.bounds.X1()
		var previousIndex int
		left := dx0 > dx1
		if left {
			xmin = root.bounds.X0() - root.bounds.W()
			previousIndex += 2
		} else {
			xmin = root.bounds.X0()
		}
		dy0, dy1 := root.bounds.Y0()-r.Y0(), r.Y1()-root.bounds.Y1()
		below := dy0 > dy1
		if below {
			ymin = root.bounds.Y0() - root.bounds.H()
			previousIndex += 1
		} else {
			ymin = root.bounds.Y0()
		}
		newRoot := newNode(gps.RectPointSize(xmin, ymin, root.bounds.W()*2, root.bounds.H()*2), n.capacity, root.depth+1)
		for i := 0; i < 4; i++ {
			if i == previousIndex {
				newRoot.quads[i] = root
			} else {
				dx := float64(i/2) * root.bounds.W()
				dy := float64(i%2) * root.bounds.H()
				r := gps.RectPointSize(xmin+dx, ymin+dy, root.bounds.W(), root.bounds.H())
				newRoot.quads[i] = newNode(r, n.capacity, root.depth)
			}
		}
		root = newRoot
	}
	return root
}

func (n *node) choose(r gps.Rect) int {
	for i := 0; i < 4; i++ {
		if n.quads[i].bounds.FullyContains(r) {
			return i
		}
	}
	return -1
}

func (n *node) findFunc(p gps.Point, f ResultFunc) {
	if !p.In(n.bounds) {
		return
	}
	for _, e := range n.entries {
		if p.In(e.bounds) {
			f(e.data, e.bounds)
		}
	}
	if n.quads[0] != nil {
		quad := 0
		dx, dy := p.X()-(n.bounds[0]+n.bounds.W()/2), p.Y()-(n.bounds[1]+n.bounds.H()/2)
		if dx > 0 {
			quad += 2
		}
		if dy > 0 {
			quad++
		}
		n.quads[quad].findFunc(p, f)
	}
	return
}

func (n *node) visit(v Visitor) {
	v.Level(n.depth, n.bounds)
	for _, e := range n.entries {
		v.Object(e.bounds)
	}
	if n.quads[0] != nil {
		for i := range n.quads {
			n.quads[i].visit(v)
		}
	}
}
