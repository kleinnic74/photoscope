package domain

import (
	"image"
	"io"
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

func NewParallelThumber(delegate Thumber, n int) Thumber {
	thumber := &parallelThumber{
		delegate: delegate,
		requests: make(chan thumbRequest),
	}
	for i := 0; i < n; i++ {
		go thumber.loop()
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

func (t *parallelThumber) loop() {
	for req := range t.requests {
		req.resp <- t.makeThumb(req)
	}
}

func (t *parallelThumber) makeThumb(r thumbRequest) thumbResult {
	img, err := t.delegate.CreateThumb(r.in, r.f, r.o, r.s)
	return thumbResult{img: img, err: err}
}
