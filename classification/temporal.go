// Package classification implements an algorithm to detect events based on timestamps of the given objects
// Objects with timestamps close to each other are considered to be in the same cluster
// Based on https://www.fxpal.com/publications/temporal-event-clustering-for-digital-photo-collections-2.pdf
package classification

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type Timestamped interface {
	Timestamp() time.Time
}

type TimestampedData []Timestamped

func (t TimestampedData) Len() int { return len(t) }

func (t TimestampedData) Less(i, j int) bool {
	return t[i].Timestamp().Before(t[j].Timestamp())
}

func (t TimestampedData) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type DistanceFunc func(i, j time.Time) float64

func TimestampDistance(d time.Duration) DistanceFunc {
	k := d.Seconds()
	return func(i, j time.Time) float64 {
		ti, tj := float64(i.Unix()), float64(j.Unix())
		return math.Exp(-math.Abs(ti-tj) / k)
	}
}

type DistanceMatrix [][]float64

func NewDistanceMatrix(data TimestampedData) DistanceMatrix {
	// k = 5 hours (5h * 60m * 60sec => 5hrs in seconds)
	return NewDistanceMatrixWithDistanceFunc(data, TimestampDistance(5*time.Hour))
}

func NewDistanceMatrixWithDistanceFunc(data TimestampedData, d DistanceFunc) DistanceMatrix {
	// Data must be sorted in ascending order
	sort.Sort(data)
	var mat DistanceMatrix
	size := data.Len()
	mat = make([][]float64, size)
	for i := range mat {
		mat[i] = make([]float64, size)
		for j := range mat[i] {
			mat[i][j] = d(data[i].Timestamp(), data[j].Timestamp())
		}
	}
	return mat
}

func (m DistanceMatrix) Dimension() int {
	return len(m)
}

func (m DistanceMatrix) NoveltyScores() ([]float64, float64, float64) {
	kernelSize, n := 6, len(m)
	scores := make([]float64, n)
	kernel := NewGaussianCherckerboardKernel(kernelSize)
	var min, max float64
	min = math.MaxFloat64
	for i := 0; i < len(m); i++ {
		var score float64
		for x := -kernelSize; x <= kernelSize; x++ {
			for y := -kernelSize; y <= kernelSize; y++ {
				indexX, indexY := i+x, i+y
				if indexX >= 0 && indexY >= 0 && indexX < n && indexY < n {
					score += kernel[x+kernelSize][y+kernelSize] * m[indexX][indexY]
				}
			}
		}
		scores[i] = score
		min, max = math.Min(min, score), math.Max(max, score)
	}
	return scores, min, max
}

type Kernel [][]float64

func NewGaussianCherckerboardKernel(l int) Kernel {
	if l%2 != 0 {
		panic(fmt.Errorf("Kernel size must be a multiple of 2, was %d", l))
	}
	k := make(Kernel, 2*l+1)
	for i := -l; i <= l; i++ {
		a := i + l
		k[a] = make([]float64, 2*l+1)
		for j := -l; j <= l; j++ {
			b := j + l
			x, y := float64(i)/float64(l), float64(j)/float64(l)
			v := math.Exp(-(x*x + y*y))
			k[a][b] = v * float64(sign(i)*sign(j))
		}
	}
	return k
}

func sign(i int) int {
	if i == 0 {
		return 0
	}
	if i < 0 {
		return -1
	}
	return 1
}
