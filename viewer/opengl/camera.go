package opengl

import (
	"github.com/go-gl/mathgl/mgl32"
)

type Camera struct {
	width, height int
	pos           mgl32.Vec3
	projection    mgl32.Mat4
	dirty         bool
}

func NewOrthoCamera(viewportWidth, viewportHeight int) *Camera {
	pos := mgl32.Vec3{0., 0., -20.}
	aspectRatio := float32(viewportHeight) / float32(viewportWidth)
	projection := mgl32.Ortho2D(-1, 1, -1*aspectRatio, 1*aspectRatio)
	//projection := mgl32.Ident4()
	//	projection = projection.Mul4(mgl32.Translate3D(pos.X(), pos.Y(), pos.Z()))
	return &Camera{
		width:      viewportWidth,
		height:     viewportHeight,
		pos:        pos,
		projection: projection,
		dirty:      false,
	}
}

func (c *Camera) Translate(dx, dy float32) {
	c.pos = c.pos.Add(mgl32.Vec3{dx, dy, 0})
	c.dirty = true
}

func (c *Camera) Projection() mgl32.Mat4 {
	if c.dirty {
		aspectRatio := float32(c.height) / float32(c.width)
		projection := mgl32.Ortho2D(-1, 1, -1*aspectRatio, 1*aspectRatio)
		//		c.projection = projection.Mul4(mgl32.Translate3D(c.pos.X(), c.pos.Y(), c.pos.Z()))
		c.projection = projection
		c.dirty = false
	}
	return c.projection
}

func (c *Camera) Resize(width, height int) {
	c.dirty = c.width != width || c.height != height
	c.width, c.height = width, height
}
