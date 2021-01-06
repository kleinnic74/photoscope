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

// RegisterTasks defines the task to split the photo collection into events at startup
func RegisterTasks(taskRepo *tasks.TaskRepository, index *boltstore.EventIndex) {
	taskRepo.RegisterWithProperties("IdentifyEventsOverLibrary", func() tasks.Task {
		return newSplitEventsTask(index)
	}, tasks.TaskProperties{
		RunOnStart:   true,
		UserRunnable: true,
	})
	taskRepo.RegisterWithProperties("IdentifyEventsInGroup", func() tasks.Task {
		return &IdentifyEventsTask{index: index}
	}, tasks.TaskProperties{
		RunOnStart:   false,
		UserRunnable: true,
	})
}

// SplitEventTask is a task that walks the whole photo library and identifies groups of photos temporally seperated
// by more than a given threshold launches asynchronous event identification tasks for each group
type SplitEventTask struct {
	index     *boltstore.EventIndex
	threshold time.Duration
}

func newSplitEventsTask(eventIndex *boltstore.EventIndex) tasks.Task {
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
	group := make([]library.ExtendedPhotoID, 0)
	var last time.Time
	for _, p := range photos {
		if p.DateTaken.Sub(last) > t.threshold && len(group) > 1 {
			executor.Submit(ctx, newIdentifyEventsTask(t.index, group))
			group = make([]library.ExtendedPhotoID, 0)
		}
		group = append(group, p.ExtendedPhotoID)
		last = p.DateTaken
	}
	if len(group) > 0 {
		executor.Submit(ctx, newIdentifyEventsTask(t.index, group))
	}
	return nil
}

// IdentifyEventsTask is a task that will identify temporal events within a group a photos and store each such event in the event index
type IdentifyEventsTask struct {
	index  *boltstore.EventIndex
	Photos []library.ExtendedPhotoID `json:"photos,omitempty"`
}

func newIdentifyEventsTask(eventIndex *boltstore.EventIndex, photos []library.ExtendedPhotoID) tasks.Task {
	return &IdentifyEventsTask{index: eventIndex, Photos: photos}
}

func (t *IdentifyEventsTask) Describe() string {
	return fmt.Sprintf("Splitting events out of %d photos", len(t.Photos))
}

type sortedPhotos struct {
	ctx context.Context
	lib library.PhotoLibrary
	ids []library.ExtendedPhotoID
}

func (s *sortedPhotos) Len() int { return len(s.ids) }
func (s *sortedPhotos) Get(i int) time.Time {
	p, _ := s.lib.Get(s.ctx, s.ids[i].ID)
	return p.DateTaken
}

func (t *IdentifyEventsTask) Execute(ctx context.Context, _ tasks.TaskExecutor, lib library.PhotoLibrary) error {
	c := NewDistanceClassifier(TimestampDistance(12 * time.Hour))

	photos := &sortedPhotos{ctx, lib, t.Photos}
	clusters := c.Clusters(photos)
	for _, cluster := range clusters {
		start, end := photos.Get(cluster.First), photos.Get(cluster.First+cluster.Count-1)
		e := boltstore.Event{
			ID:   boltstore.EventID(start.Format(time.RFC3339)),
			From: start,
			To:   end,
		}
		ids := t.Photos[cluster.First : cluster.First+cluster.Count]
		t.index.AddPhotosToEvent(ctx, e, ids)
	}
	return nil
}
