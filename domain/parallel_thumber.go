package domain

import (
	"context"
	"image"
	"io"
	"runtime"
)

type thumbRequest struct {
	in io.Reader
	f  Format
	o  Orientation
	s  ThumbSize

	resp chan<- thumbResult
}

type thumbResult struct {
	img image.Image
	err error
}

type parallelThumber struct {
	delegate Thumber

	requests chan thumbRequest
}

func CalculateOptimumParallelism() int {
	return runtime.NumCPU()
}

func NewParallelThumber(ctx context.Context, delegate Thumber, n int) Thumber {
	thumber := &parallelThumber{
		delegate: delegate,
		requests: make(chan thumbRequest),
	}
	for i := 0; i < n; i++ {
		go thumber.loop(ctx)
	}
	return thumber
}

func (t *parallelThumber) CreateThumb(in io.Reader, f Format, o Orientation, size ThumbSize) (image.Image, error) {
	res := make(chan thumbResult)
	req := thumbRequest{in, f, o, size, res}

	t.requests <- req
	result := <-res
	return result.img, result.err
}

func (t *parallelThumber) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-t.requests:
			req.resp <- t.makeThumb(req)
		}
	}
}

func (t *parallelThumber) makeThumb(r thumbRequest) thumbResult {
	img, err := t.delegate.CreateThumb(r.in, r.f, r.o, r.s)
	return thumbResult{img: img, err: err}
}
