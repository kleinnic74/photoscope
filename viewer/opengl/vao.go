package opengl

import (
	"fmt"

	"github.com/go-gl/gl/v4.2-core/gl"
)

// Vao is a vertex array object
type Vao struct {
	vbo    uint32
	vao    uint32
	length int32
}

// MakeVao creates a Vao from the given points
func MakeVao(points []float32, nbComponents int32) *Vao {

	if len(points)%int(nbComponents) != 0 {
		panic(fmt.Errorf("Bad value for component or bad number of elements in array: %d not dividable by %d", len(points), nbComponents))
	}
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, nbComponents*4, nil)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, nbComponents*4, gl.PtrOffset(3*4))
	return &Vao{vbo, vao, int32(len(points)) / nbComponents}
}

// Draw draws the vao o the current context
func (vao *Vao) Draw() {
	gl.BindVertexArray(vao.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, vao.length)
}
