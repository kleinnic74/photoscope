package gps

import "math"

type Rect [4]float64

func RectFrom(x0, y0, x1, y1 float64) Rect {
	return Rect([4]float64{math.Min(x0, x1), math.Min(y0, y1), math.Max(x0, x1), math.Max(y0, y1)})
}

func RectPointSize(x0, y0, w, h float64) Rect {
	return Rect{x0, y0, x0 + w, y0 + h}
}

func (r Rect) W() float64 {
	return r[2] - r[0]
}

func (r Rect) H() float64 {
	return r[3] - r[1]
}

func (r Rect) X0() float64 {
	return r[0]
}

func (r Rect) Y0() float64 {
	return r[1]
}

func (r Rect) X1() float64 {
	return r[2]
}

func (r Rect) Y1() float64 {
	return r[3]
}

func (r Rect) Center() Point {
	return Point{(r[0] + r[2]) / 2, (r[1] + r[3]) / 2}
}

func (r Rect) HalfSize() (float64, float64) {
	return r.W() / 2, r.H() / 2
}

func (r Rect) FullyContains(other Rect) bool {
	return other[0] >= r[0] && other[2] < r[2] && other[1] >= r[1] && other[3] < r[3]
}

func (r Rect) Translate(x, y float64) Rect {
	return RectFrom(r[0]+x, r[1]+y, r[2]+x, r[3]+y)
}

type Point [2]float64

func PointFromLatLon(lat, lon float64) Point {
	return Point{lon, lat}
}

func (p Point) X() float64 {
	return p[0]
}

func (p Point) Y() float64 {
	return p[1]
}

func (p Point) Lat() float64 {
	return p[1]
}

func (p Point) Lon() float64 {
	return p[0]
}

func (p Point) In(r Rect) bool {
	return p[0] >= r[0] && p[0] < r[2] && p[1] >= r[1] && p[1] < r[3]
}
