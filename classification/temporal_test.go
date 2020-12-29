package classification_test

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"bitbucket.org/kleinnic74/photos/classification"
	"github.com/stretchr/testify/assert"
)

type WithTimestamp time.Time

func (t WithTimestamp) Timestamp() time.Time {
	return time.Time(t)
}

func TestNewDistanceMatrix(t *testing.T) {
	data := []classification.Timestamped{
		parseTime("2018-02-24T15:30:30Z"),
		parseTime("2018-01-13T16:30:00Z"),
		parseTime("2018-01-24T15:25:00Z"),
	}
	mat := classification.NewDistanceMatrixWithDistanceFunc(classification.TimestampDistance(12 * time.Hour))
	ssm := mat.SelfSimilarityMatrix(data)
	if len(ssm) != 3 {
		t.Fatalf("Expected self-similarity matrix of 3x3, but got: %dx%d", len(ssm), len(ssm[0]))
	}
}

func parseTime(value string) WithTimestamp {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return WithTimestamp(t)
}

func TestGaussianCheckerboardKernel(t *testing.T) {
	k := classification.NewGaussianCherckerboardKernel(6)
	assert.Equal(t, 13, len(k))
	assert.Equal(t, 13, len(k[0]))
	img := image.NewRGBA(image.Rect(0, 0, len(k), len(k)))
	for i := 0; i < len(k); i++ {
		for j := 0; j < len(k[i]); j++ {
			v := uint8(math.Abs(k[i][j]) * 255.)
			var r, g, b uint8
			if k[i][j] < 0 {
				r = v
			} else {
				g = v
			}
			img.SetRGBA(i, j, color.RGBA{R: r, G: g, B: b, A: 255})
			fmt.Fprintf(os.Stdout, "\t%.4f", k[i][j])
		}
		fmt.Println()
	}
	out, _ := os.Create("gaussianCheckerboard.png")
	defer out.Close()
	png.Encode(out, img)
}

func TestNoveltyScores(t *testing.T) {
	data := loadEventData(t)
	m := classification.NewDistanceMatrixWithDistanceFunc(classification.TimestampDistance(12 * time.Hour))
	ssm := m.SelfSimilarityMatrix(data)
	img := image.NewRGBA(image.Rect(0, 0, len(data), 2*len(data)))
	for i := 0; i < len(ssm); i++ {
		for j := 0; j < len(ssm); j++ {
			g := uint8(255. * ssm[i][j])
			img.Set(i, j, color.RGBA{g, g, g, 255})
		}
	}
	scores := m.NoveltyScores(ssm, 3)
	assert.Equal(t, len(data), len(scores.Scores))
	draw.Draw(img, image.Rect(0, len(data), len(data), 2*len(data)), image.White, image.ZP, draw.Src)
	for i := 0; i < len(data); i++ {
		y := 2*len(data) - int(scores.Normalized(i)*float64(len(data)))
		var c color.Color
		switch scores.Scores[i].Boundary {
		case true:
			c = color.RGBA{R: 255, A: 255}
		default:
			c = color.Black
		}
		img.Set(i, y, c)
		fmt.Fprintf(os.Stdout, "%d: %f [%f] [min=%f, max=%f]\n", i, scores.Scores[i].Score, scores.Normalized(i), scores.Min, scores.Max)
	}
	saveImage(img, "noveltyScores.png")
}

func TestFindClusters(t *testing.T) {
	inputData := loadEventData(t)
	m := classification.NewDistanceMatrixWithDistanceFunc(classification.TimestampDistance(12 * time.Hour))
	clusters := m.Clusters(inputData)
	for i, c := range clusters {
		fmt.Fprintf(os.Stdout, "Cluster #%d:\n", i)
		for _, d := range c {
			item := d.(data)
			fmt.Fprintf(os.Stdout, "  %d - %s\n", item.line, item.id)
		}
	}
}

func saveImage(img image.Image, name string) {
	out, _ := os.Create(name)
	defer out.Close()
	png.Encode(out, img)
}

type data struct {
	line int
	ts   time.Time
	id   string
}

func (d data) Timestamp() time.Time {
	return d.ts
}

func loadEventData(t *testing.T) []classification.Timestamped {
	in, err := os.Open("testdata/events.csv")
	if err != nil {
		t.Fatalf("Failed to open testdata: %s", err)
		return nil
	}
	defer in.Close()
	r := bufio.NewReader(in)
	var d []classification.Timestamped
	var lineNb int
	for line, err := r.ReadString('\n'); err == nil; line, err = r.ReadString('\n') {
		line = strings.TrimRight(line, "\n")
		if len(line) == 0 {
			continue
		}
		parts := strings.SplitN(line, ",", 2)
		ts, err := time.Parse(time.RFC3339, parts[0])
		if err != nil {
			t.Fatalf("Error while parsing input data '%s' at %d: %s", parts[0], lineNb, err)
		}
		d = append(d, data{line: lineNb, ts: ts, id: parts[1]})
		lineNb++
	}
	return d
}
