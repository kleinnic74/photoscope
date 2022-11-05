package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"time"

	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/embed"
	"bitbucket.org/kleinnic74/photos/logging"
	"bitbucket.org/kleinnic74/photos/swarm"
	"go.uber.org/zap"
)

func addRemoteThumber(self swarm.InstanceID, thumbers *domain.Thumbers) swarm.PeerHandler {
	return func(ctx context.Context, peer swarm.Peer) {
		if self == peer.ID {
			// Do not add ourselves as remote thumber
			return
		}
		thumber, err := swarm.NewRemoteThumber(fmt.Sprintf("%s/thumb", peer.URL), domain.JPEG)
		if err != nil {
			logging.From(ctx).Warn("Failed to create remote thumber", zap.Error(err))
			return
		}
		logging.From(ctx).Info("Remote thumber added", zap.String("peer.url", peer.URL))
		thumbers.Add(thumber, 1)
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
