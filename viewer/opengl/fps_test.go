package opengl

import (
	"log"
	"testing"
	"time"
)

type MockClock struct {
	time      time.Time
	sleptTime time.Duration
}

func (c *MockClock) Now() time.Time {
	return c.time
}

func (c *MockClock) Sleep(d time.Duration) {
	c.sleptTime = d
}

func (c *MockClock) Advance(d time.Duration) {
	log.Printf("Moving clock forward by %s", d)
	c.time = c.time.Add(d)
}
func TestFpsCount(t *testing.T) {
	clock := &MockClock{time: time.Now()}
	fps := NewFpsWithClock(10, clock)
	fps.BeginFrame()
	clock.Advance(40 * time.Millisecond)
	fps.EndFrame()
	if fps.frames != 1 {
		t.Errorf("Expected frame count to be %d, was %d", 1, fps.frames)
	}
	clock.expectSlept(t, 60*time.Millisecond)
}

func TestFpsCountPerSecond(t *testing.T) {
	clock := &MockClock{time: time.Now()}
	var frames float32
	fps := newFps(10, clock, func(framesPerSecond float32) {
		log.Printf("FPS callback: %d", framesPerSecond)
		frames = framesPerSecond
	})
	fps.BeginFrame()
	clock.Advance(2 * time.Second)
	fps.EndFrame()
	clock.expectSlept(t, 0*time.Millisecond)
	if frames != 0.5 {
		t.Errorf("Expected frame count to be %f, was %f", 0.5, frames)
	}
}

func (c MockClock) expectSlept(t *testing.T, expected time.Duration) {
	if c.sleptTime != expected {
		t.Errorf("Expected clock to sleep %s, but only did %s", expected, c.sleptTime)
	}
}
