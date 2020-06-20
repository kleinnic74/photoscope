package opengl

import (
	"image"
	"image/draw"

	// Register PNG image type
	"image/jpeg"
	_ "image/png"
	"log"
	"os"
	"time"

	"github.com/disintegration/imaging"
	"github.com/go-gl/gl/v4.2-core/gl"
)

// Texture represents an OpenGL texture
type Texture struct {
	textureID    uint32
	path         string
	lastModified time.Time
}

func init() {
	log.Printf("Default JPEG quality: %#v", jpeg.DefaultQuality)
}

// LoadTexture loads a texture from the given image file
func LoadTexture(path string) (*Texture, error) {
	log.Printf("Loading texture from '%s'...", path)
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	img = imaging.FlipV(img)
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(bounds.Dx()), int32(bounds.Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
	return &Texture{id, path, info.ModTime()}, nil
}

// BindTexture binds this texture to the current openGL context
func (t *Texture) BindTexture() {
	gl.BindTexture(gl.TEXTURE_2D, t.textureID)
}
