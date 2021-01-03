package classification

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/kleinnic74/photos/consts"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/library/boltstore"
	"bitbucket.org/kleinnic74/photos/tasks"
)

type SplitEventTask struct {
	index     *boltstore.EventIndex
	threshold time.Duration
}

func NewSplitEventsTask(eventIndex *boltstore.EventIndex) tasks.Task {
	return &SplitEventTask{index: eventIndex, threshold: 7 * 24 * time.Hour}
}

func (t *SplitEventTask) Describe() string {
	return "Split photo timeline into events"
}

func (t *SplitEventTask) Execute(ctx context.Context, executor tasks.TaskExecutor, lib library.PhotoLibrary) error {
	photos, err := lib.FindAll(ctx, consts.Ascending)
	if err != nil {
		return err
	}
	var group []*library.Photo
	var last time.Time
	for _, p := range photos {
		if p.DateTaken.Sub(last) > t.threshold && len(group) > 1 {
			executor.Submit(ctx, NewIdentifyEventsTask(t.index, group))
			group = []*library.Photo{}
		}
	}
	return nil
}

type IdentifyEventsTask struct {
	index  *boltstore.EventIndex
	photos []*library.Photo
}

func NewIdentifyEventsTask(eventIndex *boltstore.EventIndex, photos []*library.Photo) tasks.Task {
	return &IdentifyEventsTask{index: eventIndex, photos: photos}
}

func (t *IdentifyEventsTask) Describe() string {
	return fmt.Sprintf("Splitting events out of %d photos", len(t.photos))
}

type sortedPhotos []*library.Photo

func (s sortedPhotos) Len() int            { return len(s) }
func (s sortedPhotos) Get(i int) time.Time { return s[i].DateTaken }

func (t *IdentifyEventsTask) Execute(ctx context.Context, _ tasks.TaskExecutor, _ library.PhotoLibrary) error {
	c := NewDistanceClassifier(TimestampDistance(12 * time.Hour))

	clusters := c.Clusters(sortedPhotos(t.photos))
	for _, cluster := range clusters {
		start, end := t.photos[cluster.First].DateTaken, t.photos[cluster.First+cluster.Count-1].DateTaken
		e := boltstore.Event{
			ID:   start.Format(time.RFC3339),
			From: start,
			To:   end,
		}
		ids := make([]library.PhotoID, cluster.Count)
		for i := 0; i < cluster.Count; i++ {
			ids[i] = t.photos[cluster.First+1].ID
		}
		t.index.AddPhotosToEvent(ctx, e, ids)
	}
	return nil
}
