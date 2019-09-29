package opengl

import (
	"fmt"
	"time"
)

type Callback func(fps float32)

type Clock interface {
	Now() time.Time
	Sleep(time.Duration)
}

type Fps struct {
	clock    Clock
	frames   int
	period   time.Duration
	refTime  time.Time
	next     time.Time
	callback Callback
}

type systemClock struct {
}

func (systemClock) Now() time.Time {
	return time.Now()
}

func (systemClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func NewSystemClock() Clock {
	return systemClock{}
}

func debugFps(frames float32) {
	fmt.Printf("%f fps\n", frames)
}

func NewFps(targetFps int) *Fps {
	return newFps(targetFps, systemClock{}, debugFps)
}

func NewFpsWithClock(targetFps int, clock Clock) *Fps {
	return newFps(targetFps, clock, debugFps)
}

func NewFpsWithCallback(targetFps int, callback Callback) *Fps {
	return newFps(targetFps, systemClock{}, callback)
}

func newFps(targetFps int, clock Clock, callback Callback) *Fps {
	return &Fps{
		clock:    clock,
		refTime:  clock.Now(),
		callback: callback,
		period:   time.Second / time.Duration(targetFps),
	}
}

func (fps *Fps) BeginFrame() {
	now := fps.clock.Now()
	fps.next = now.Add(fps.period)
}

func (fps *Fps) EndFrame() {
	now := fps.clock.Now()
	delta := now.Sub(fps.refTime)
	fps.frames = fps.frames + 1
	if delta > time.Second {
		var intervalMs float32
		intervalMs = float32(delta.Nanoseconds()) / float32(time.Millisecond.Nanoseconds())
		fmt.Printf("interval=%fms frames=%d\n", intervalMs, fps.frames)
		frameCount := float32(fps.frames) * 1000. / intervalMs
		fps.callback(frameCount)
		fps.refTime = now
		fps.frames = 0
	}
	if fps.next.After(now) {
		fps.clock.Sleep(fps.next.Sub(now))
	}
}
