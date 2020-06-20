// temporal.go
// Implements an algorithm to detect events based on timestamps of the given objects
// Objects with timestamps close to each other are considered to be in the same cluster
// Based on https://www.fxpal.com/publications/temporal-event-clustering-for-digital-photo-collections-2.pdf
package classification

import (
	"math"
	"sort"
	"time"
)

type Timestamped interface {
	Timestamp() time.Time
}

type TimestampedData []Timestamped

func (t TimestampedData) Len() int {

	return len(t)
}

func (t TimestampedData) Less(i, j int) bool {
	return t[i].Timestamp().Before(t[j].Timestamp())
}

func (t TimestampedData) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type DistanceFunc func(i, j time.Time) float64

func TimestampDistanceK(k float64) DistanceFunc {
	return func(i, j time.Time) float64 {
		return math.Min(1, math.Exp(-math.Abs(float64(i.Unix())-float64(j.Unix()))/k))
	}
}

type DistanceMatrix [][]float64

func NewDistanceMatrix(data TimestampedData) DistanceMatrix {
	// k = 5 hours (5h * 60m * 60sec => 5hrs in seconds)
	return NewDistanceMatrixWithDistanceFunc(data, TimestampDistanceK(60*60*5))
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
