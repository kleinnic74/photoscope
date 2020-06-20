package main

import (
	"log"
	"runtime"

	"bitbucket.org/kleinnic74/photos/viewer/opengl"

	"github.com/go-gl/gl/v4.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	fpsTarget = 60
)

var (
	square = []float32{
		// Point        // Texture coords
		-0.5, 0.5, 0, 0, 1,
		-0.5, -0.5, 0, 0, 0,
		0.5, -0.5, 0, 1, 0,

		-0.5, 0.5, 0, 0, 1,
		0.5, 0.5, 0, 1, 1,
		0.5, -0.5, 0, 1, 0,
	}
)

func main() {
	runtime.LockOSThread()
	window := initGlfw()
	defer glfw.Terminate()

	program := initOpenGL()
	if err := program.LoadShader("assets/texturevertex.frag", gl.VERTEX_SHADER); err != nil {
		panic(err)
	}
	if err := program.LoadShader("assets/texturefrag.frag", gl.FRAGMENT_SHADER); err != nil {
		panic(err)
	}
	texture, err := opengl.LoadTexture("assets/tileset.png")
	if err != nil {
		panic(err)
	}

	lp := program.Link()
	scene := opengl.NewScene()
	scene.Add(opengl.MakeVao(square, 5))

	w, h := window.GetSize()
	camera := opengl.NewOrthoCamera(w, h)
	fps := opengl.NewFps(fpsTarget)
	for !window.ShouldClose() {
		fps.BeginFrame()
		lp.Use()
		lp.SetMat4("projection", camera.Projection())
		texture.BindTexture()
		draw(scene, window, lp)
		lp = program.UpdateModifiedShaders()
		fps.EndFrame()
	}
}

func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(400, 300, "Photo Viewer", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

func initOpenGL() *opengl.Program {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	return opengl.NewProgram()
}

func draw(scene opengl.Drawable, window *glfw.Window, program *opengl.LinkedProgram) {
	program.Use()
	scene.Draw()
	window.SwapBuffers()
	glfw.PollEvents()
}
