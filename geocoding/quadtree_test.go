package geocoding

import (
	"testing"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	"github.com/stretchr/testify/assert"
)

var data = []struct {
	Rects []gps.Rect
	Point gps.Point
	In    []int
}{
	{Rects: []gps.Rect{
		gps.RectFrom(16.3551666, 48.2018494, 16.3751666, 48.2218494),
		gps.RectFrom(16.2433526, 48.0455922, 16.3233526, 48.1255922),
	},
		Point: gps.Point{16.3651666, 48.2118494},
		In:    []int{0},
	},
}

func TestQuadTreeMany(t *testing.T) {
	qt := NewQuadTree(gps.WorldBounds)

	for _, d := range data {
		for nb, rect := range d.Rects {
			qt.InsertRect(rect, nb)
		}
		objs := qt.Find(d.Point)
		if len(objs) != len(d.In) {
			t.Errorf("Bad number of results, expected %d, got %d", len(d.In), len(objs))
		}
		for i, a := range objs {
			actual := a.(int)
			assert.Equal(t, d.In[i], actual)
		}
	}
}

func TestQuadTreeSplit(t *testing.T) {
	qt := NewQuadTree(gps.WorldBounds)
	r := gps.RectFrom(16.2433526, 48.0455922, 16.3233526, 48.1255922)
	for i := 0; i < 100; i++ {
		qt.InsertRect(r, i)
		r = r.Translate(r.W()+0.01, r.H()+0.01)
	}
	matching := qt.Find(gps.PointFromLatLon(48.0855922, 16.2833526))
	if len(matching) != 1 {
		t.Fatalf("Expected to match exactly one rectangle but found %d", len(matching))
	}
	actual := matching[0].(int)
	if actual != 0 {
		t.Errorf("Bad index returned, expected 0, got %d", actual)
	}
}

func TestQuadTreeSingle(t *testing.T) {
	qt := NewQuadTree(gps.WorldBounds)
	qt.InsertRect(gps.RectFrom(-40, -30, -35, -20), "one")
	qt.InsertRect(gps.RectFrom(20, 30, 31, 50), "two")
	qt.InsertRect(gps.RectFrom(28, 45, 33, 49), "three")

	items := qt.Find(gps.Point{30, 30})
	assert.Equal(t, 1, len(items))
	assert.Equal(t, "two", items[0])
}

func TestNodeSplitAndAdd(t *testing.T) {
	node := newNode(gps.RectFrom(-1, -1, 1, 1), 5, 2)
	node.split()
	for i := 0; i < 4; i++ {
		assert.Equal(t, 1, node.quads[i].depth, "Bad depth for node %d", i)
	}
	assert.Equal(t, gps.RectFrom(-1, -1, 0, 0), node.quads[0].bounds)
	assert.Equal(t, gps.RectFrom(-1, 0, 0, 1), node.quads[1].bounds)
	assert.Equal(t, gps.RectFrom(0, -1, 1, 0), node.quads[2].bounds)
	assert.Equal(t, gps.RectFrom(0, 0, 1, 1), node.quads[3].bounds)
	node.add(gps.RectFrom(-0.6, 0.2, -0.4, 0.4), "quadOne")
	node.add(gps.RectFrom(0.6, -0.6, 0.8, -0.4), "quadTwo")
	assert.Equal(t, "quadOne", node.quads[1].entries[0].data)
	assert.Equal(t, "quadTwo", node.quads[2].entries[0].data)
}
