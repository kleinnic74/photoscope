package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"bitbucket.org/kleinnic74/photos/consts"
	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/embed"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/logging"
	"bitbucket.org/kleinnic74/photos/swarm"
	"go.uber.org/zap"
)

const (
	instanceBucket = "_instance"
	idKey          = "id"
)

type PropertyProvider func(context.Context) string
type PropertyDefinition func() (string, PropertyProvider)

func WithProperty(name string, f PropertyProvider) PropertyDefinition {
	return func() (string, PropertyProvider) {
		return name, f
	}
}

func WithPropertyValue(name string, value string) PropertyDefinition {
	return func() (string, PropertyProvider) {
		return name, func(context.Context) string { return value }
	}
}

func NewInstance(ctx context.Context, id library.LibraryID, p ...PropertyDefinition) (*swarm.Instance, error) {
	hostnameFQ, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	hostname := strings.Split(hostnameFQ, ".")[0]
	i := &swarm.Instance{
		ID:         string(id),
		Properties: make(map[string]string),
	}
	i.Properties["id"] = i.ID
	for _, pd := range p {
		name, f := pd()
		i.Properties[name] = f(ctx)
	}
	i.Name = fmt.Sprintf("Photoscope on %s", hostname)
	return i, err
}

func DefaultInstanceProperties() []PropertyDefinition {
	return []PropertyDefinition{
		WithProperty("ts", thumbCreationSpeed),
		WithPropertyValue("gc", consts.GitCommit),
		WithPropertyValue("gr", consts.GitRepo),
	}
}

func thumbCreationSpeed(ctx context.Context) string {
	log := logging.From(ctx)

	log.Info("Benchmarking local host...")
	refImg, err := embed.Get("/jpg/reference.jpg")
	if err != nil {
		// Cannot open reference image, assume high costs
		log.Warn("Cannot load benchmark image", zap.Error(err))
		return fmt.Sprintf("%d", int64(math.MaxInt64))
	}

	var total int64
	for i := 0; i < 3; i++ {
		cost, err := benchmarkThumb(ctx, bytes.NewReader(refImg))
		if err != nil {
			log.Warn("Cannot create thumb out of reference image", zap.Error(err))
			return fmt.Sprintf("%d", int64(math.MaxInt64))
		}
		total += cost
	}
	cost := total / 3
	log.Info("Benchmark results", zap.Int64("thumbDuration", cost))
	return fmt.Sprintf("%d", cost)
}

func benchmarkThumb(ctx context.Context, img io.Reader) (cost int64, err error) {
	var t domain.LocalThumber
	cost = math.MaxInt64

	start := time.Now()
	if _, err = t.CreateThumb(img, domain.JPEG, domain.NormalOrientation, domain.Small); err != nil {
		return
	}
	cost = time.Since(start).Microseconds()
	return
}
