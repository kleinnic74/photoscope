// Package classification implements an algorithm to detect events based on timestamps of the given objects
// Objects with timestamps close to each other are considered to be in the same cluster
// Based on https://www.fxpal.com/publications/temporal-event-clustering-for-digital-photo-collections-2.pdf
package classification

import (
	"math"
	"time"
)

type TimestampedData interface {
	Len() int
	Get(int) time.Time
}

type DistanceFunc func(i, j time.Time) float64

func TimestampDistance(d time.Duration) DistanceFunc {
	k := d.Seconds()
	return func(i, j time.Time) float64 {
		ti, tj := float64(i.Unix()), float64(j.Unix())
		return math.Exp(-math.Abs(ti-tj) / k)
	}
}

type DistanceClassifier struct {
	d DistanceFunc
}

func NewDistanceClassifier(d DistanceFunc) (mat DistanceClassifier) {
	return DistanceClassifier{d: d}
}

type NoveltyScore struct {
	Score      float64
	Derivative float64
	Boundary   bool
}
type NoveltyScores struct {
	Scores   []NoveltyScore
	Min, Max float64
}

func (s NoveltyScores) Normalized(i int) float64 {
	return (s.Scores[i].Score - s.Min) / (s.Max - s.Min)
}

func (m DistanceClassifier) SelfSimilarityMatrix(data TimestampedData) [][]float64 {
	size := data.Len()
	// Calculate self-similarity matrix
	ssm := make([][]float64, size)
	for i := range ssm {
		ssm[i] = make([]float64, size)
		for j := range ssm[i] {
			ssm[i][j] = m.d(data.Get(i), data.Get(j))
		}
	}
	return ssm
}

func (m DistanceClassifier) NoveltyScores(ssm [][]float64, kernelSize int) (scores NoveltyScores) {
	n := len(ssm)
	scores.Scores = make([]NoveltyScore, n)
	scores.Min = math.Inf(1)
	scores.Max = math.Inf(-1)
	kernel := NewGaussianCherckerboardKernel(kernelSize)
	for i := 0; i < n; i++ {
		var score float64
		for x := -kernelSize; x <= kernelSize; x++ {
			for y := -kernelSize; y <= kernelSize; y++ {
				indexX, indexY := i+x, i+y
				if indexX >= 0 && indexY >= 0 && indexX < n && indexY < n {
					score += kernel[x+kernelSize][y+kernelSize] * ssm[indexX][indexY]
				}
			}
		}
		scores.Scores[i].Score = score
		if i > 0 {
			scores.Scores[i].Derivative = score - scores.Scores[i-1].Score
			scores.Scores[i-1].Boundary = scores.Scores[i].Derivative < 0 && scores.Scores[i-1].Derivative > 0
		}
		scores.Min = math.Min(scores.Min, score)
		scores.Max = math.Max(scores.Max, score)
	}
	return scores
}

type Cluster struct {
	First, Count int
}

func (m DistanceClassifier) Clusters(data TimestampedData) (clusters []Cluster) {
	ssm := m.SelfSimilarityMatrix(data)
	noveltyScores := m.NoveltyScores(ssm, 3)
	var cluster Cluster
	for i, s := range noveltyScores.Scores {
		if s.Boundary {
			if i-cluster.First > 0 {
				clusters = append(clusters, cluster)
			}
			cluster = Cluster{First: i}
		}
		cluster.Count++
	}
	clusters = append(clusters, cluster)
	return
}

type Kernel [][]float64

func NewGaussianCherckerboardKernel(l int) Kernel {
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
