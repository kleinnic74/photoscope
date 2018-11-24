package opengl

import "github.com/go-gl/gl/v4.2-core/gl"

// Drawable can be drawn into the current context
type Drawable interface {
	Draw()
}

type Scene struct {
	objects []Drawable
}

func NewScene() *Scene {
	return &Scene{objects: make([]Drawable, 0)}
}

func (s *Scene) Add(o Drawable) {
	s.objects = append(s.objects, o)
}

func (s *Scene) Draw() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	for _, o := range s.objects {
		o.Draw()
	}
}
