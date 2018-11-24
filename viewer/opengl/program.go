package opengl

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-gl/gl/v4.2-core/gl"
)

// Program represents an OpenGL program
type Program struct {
	prog    uint32
	shaders []*Shader
}

// NewProgram creates a new OpenGL program
func NewProgram() *Program {
	prog := gl.CreateProgram()
	return &Program{prog: prog, shaders: make([]*Shader, 0)}
}

// ShaderID represents on OpenGL shader instance
type ShaderID uint32

type Shader struct {
	shaderID     ShaderID
	shaderType   uint32
	path         string
	lastModified time.Time
}

func (p *Program) LoadShader(path string, shaderType uint32) error {
	shader, err := initShader(path, shaderType)
	if err != nil {
		return err
	}
	gl.AttachShader(p.prog, uint32(shader.shaderID))
	p.shaders = append(p.shaders, shader)
	return nil
}

// SetFloat sets the shader variable with the given name to the given value
func (p *Program) SetFloat(name string, value float32) {
	location := gl.GetUniformLocation(p.prog, gl.Str(name+"\x00"))
	gl.Uniform1f(location, value)
}

func initShader(path string, shaderType uint32) (*Shader, error) {
	stats, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	log.Printf("Loading shader from %s...", path)
	shaderSource, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	id, err := compileShader(string(shaderSource), shaderType)
	if err != nil {
		return nil, err
	}
	return &Shader{
		shaderID:     id,
		shaderType:   shaderType,
		path:         path,
		lastModified: stats.ModTime(),
	}, nil
}

// HasChanged checks if this shader has changed on disk
func (s *Shader) HasChanged() bool {
	info, err := os.Stat(s.path)
	if err != nil {
		return false
	}
	return !info.ModTime().Equal(s.lastModified)
}

// UpdateModifiedShaders checks if shaders have changed on disk and reloads them if needed
func (p *Program) UpdateModifiedShaders() *LinkedProgram {
	changed := false
	for i, s := range p.shaders {
		if s.HasChanged() {
			next, err := initShader(s.path, s.shaderType)
			if err != nil {
				continue
			}
			gl.DetachShader(p.prog, uint32(s.shaderID))
			gl.AttachShader(p.prog, uint32(next.shaderID))
			p.shaders[i] = next
			changed = true
		}
	}
	if changed {
		return p.Link()
	} else {
		var lp = LinkedProgram(*p)
		return &lp
	}
}

func compileShader(source string, shaderType uint32) (ShaderID, error) {
	shader := gl.CreateShader(shaderType)
	source = source + "\x00"
	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	defer free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("%v", log)
	}
	return ShaderID(shader), nil
}

// Link links this program so that it can be used
func (p *Program) Link() *LinkedProgram {
	gl.LinkProgram(p.prog)
	var status int32
	gl.GetProgramiv(p.prog, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(p.prog, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength)+1)
		gl.GetProgramInfoLog(p.prog, logLength, nil, gl.Str(log))
		panic("Failed to link program:\n" + log)
	}
	var lp = LinkedProgram(*p)
	return &lp
}

// LinkedProgram is a linked OpenGL program ready to be used
type LinkedProgram Program

// Use uses this program in the current OpenGL context
func (p *LinkedProgram) Use() {
	gl.UseProgram(p.prog)
}
