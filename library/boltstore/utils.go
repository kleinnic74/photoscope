package boltstore

import (
	"github.com/boltdb/bolt"
)

// Cursor provides additional convenience functions acround a bolt.Cursor
type Cursor interface {
	Reverse() Cursor
	Skip(count uint) Cursor
	Limit(count uint) Cursor
	First() (key, value []byte)
	Next() (key, value []byte)
}

type baseCursor struct {
	delegate *bolt.Cursor
	limit    int
	skip     int
}

type forwardCursor struct {
	baseCursor
}

type reverseCursor struct {
	baseCursor
}

func newForwardCursor(delegate *bolt.Cursor) Cursor {
	return &forwardCursor{baseCursor: baseCursor{delegate: delegate, limit: -1}}
}

func (c *forwardCursor) Reverse() Cursor {
	return newReverseCursor(c.delegate)
}

func (c *forwardCursor) Skip(count uint) Cursor {
	c.skip = int(count)
	return c
}

func (c *forwardCursor) Limit(count uint) Cursor {
	c.limit = int(count)
	return c
}

func (c *forwardCursor) First() (key []byte, value []byte) {
	var k, v []byte
	for k, v = c.delegate.First(); c.skip > 0 && k != nil; k, v = c.delegate.Next() {
		c.skip--
	}
	c.limit--
	return k, v
}

func (c *forwardCursor) Next() (key []byte, value []byte) {
	if c.limit == 0 {
		return nil, nil
	}
	k, v := c.delegate.Next()
	for ; c.skip > 0 && k != nil; k, v = c.delegate.Next() {
		c.skip--
	}
	c.limit--
	return k, v
}

//------------------------------------------------------------------------------

func newReverseCursor(delegate *bolt.Cursor) Cursor {
	return &reverseCursor{baseCursor: baseCursor{delegate: delegate}}
}

func (c *reverseCursor) Reverse() Cursor {
	return newForwardCursor(c.delegate)
}

func (c *reverseCursor) First() (key, value []byte) {
	var k, v []byte
	for k, v = c.delegate.Last(); c.skip > 0 && k != nil; k, v = c.delegate.Prev() {
		c.skip--
	}
	c.limit--
	return k, v
}

func (c *reverseCursor) Limit(count uint) Cursor {
	c.limit = int(count)
	return c
}

func (c *reverseCursor) Skip(count uint) Cursor {
	c.skip = int(count)
	return c
}

func (c *reverseCursor) Next() (key []byte, value []byte) {
	if c.limit == 0 {
		return nil, nil
	}
	k, v := c.delegate.Prev()
	for ; c.skip > 0 && k != nil; k, v = c.delegate.Prev() {
		c.skip--
	}
	c.limit--
	return k, v
}
