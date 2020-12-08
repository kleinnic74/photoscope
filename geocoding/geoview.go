package geocoding

import (
	"fmt"
	"image/color"
	"io"

	"bitbucket.org/kleinnic74/photos/domain/gps"
	svg "github.com/ajstarks/svgo"
)

var (
	blue = color.RGBA{0, 40, 255, 255}

	strokeGrid   = []string{`stroke="gray"`, `stroke-width="1px"`, `fill="none"`}
	strokeQuad   = []string{`stroke="blue"`, `stroke-width="1px"`}
	strokeObject = []string{`stroke="red"`, `stroke-width="1px"`, `fill="none"`}
)

type GeoView struct {
	canvas *svg.SVG
}

func NewGeoView(out io.Writer) *GeoView {
	canvas := svg.New(out)
	return &GeoView{
		canvas: canvas,
	}
}

func xlinePath(bounds gps.Rect) string {
	center := bounds.Center()
	return fmt.Sprintf("M %f %f l %f %f", bounds.X0(), center.Y(), bounds.W(), 0.)
}

func ylinePath(bounds gps.Rect) string {
	center := bounds.Center()
	return fmt.Sprintf("M %f %f l %f %f", center.X(), bounds.Y0(), 0., bounds.W())
}

func rectPath(bounds gps.Rect) string {
	return fmt.Sprintf("M %f %f l 0 %f l %f 0 l 0 %f Z", bounds.X0(), bounds.Y0(), bounds.H(), bounds.W(), -bounds.H())
}

func (g *GeoView) Begin(bounds gps.Rect) {
	g.canvas.Startpercent(100, 100, fmt.Sprintf(`viewBox="%f %f %f %f"`, bounds.X0(), bounds.Y0(), bounds.W(), bounds.H()))
	g.canvas.Gtransform("scale(1,-1)")
	g.canvas.Path("M 0 -90 l 0 180", strokeGrid...)
	g.canvas.Path("M -180 0 l 360 0", strokeGrid...)
	g.canvas.Path("M -180 -90 L -180 90 L 180 90 L 180 -90 Z", strokeGrid...)
}

func (g *GeoView) Level(depth int, bounds gps.Rect) {
	g.canvas.Group()
	g.canvas.Path(xlinePath(bounds), strokeQuad...)
	g.canvas.Path(ylinePath(bounds), strokeQuad...)
	g.canvas.Gend()
}

func (g *GeoView) Object(bounds gps.Rect) {
	g.canvas.Group()
	g.canvas.Path(rectPath(bounds), strokeObject...)
	g.canvas.Gend()
}

func (g *GeoView) End() {
	g.canvas.Gend()
	g.canvas.End()
}
