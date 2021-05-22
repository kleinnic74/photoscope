package logging

import (
	"io"

	"go.uber.org/zap"
)

type LogsExporter interface {
	Export(io.Writer, bool) error
}

type logLine []byte

type memoryLogs struct {
	lines      []logLine
	start      int
	startCycle bool
	w          int
	wCycle     bool
}

func NewMemoryLogger(size int) zap.Sink {
	return &memoryLogs{
		lines: make([]logLine, size),
	}
}

func (m *memoryLogs) Write(p []byte) (n int, err error) {
	l := make(logLine, len(p))
	copy(l, p)
	m.lines[m.w] = l
	m.w = m.w + 1
	if m.w >= len(m.lines) {
		if m.start <= m.w {
			m.start = m.start + 1
			if m.start >= len(m.lines) {
				m.start = 0
				m.startCycle = !m.startCycle
			}
		}
		m.w = m.w % len(m.lines)
		m.wCycle = !m.wCycle
	}
	return len(p), nil
}

func (m *memoryLogs) Sync() error {
	return nil
}

func (m *memoryLogs) Close() error {
	return nil
}

func (m *memoryLogs) Export(w io.Writer, revert bool) error {
	if revert {
		return m.dumpBackward(w)
	}
	return m.dumpForward(w)
}

func (m *memoryLogs) dumpForward(w io.Writer) error {
	i := m.start
	cycle := m.startCycle
	for m.wCycle != cycle || i < m.w {
		if _, err := w.Write(m.lines[i]); err != nil {
			return err
		}
		i = i + 1
		if i >= len(m.lines) {
			i = 0
			cycle = !m.startCycle
		}
	}
	return nil
}

func (m *memoryLogs) dumpBackward(w io.Writer) error {
	i := m.w
	cycle := m.wCycle
	for m.startCycle != cycle || i > m.start {
		i = i - 1
		if i < 0 {
			i = len(m.lines) - 1
			cycle = !cycle
		}
		if _, err := w.Write(m.lines[i]); err != nil {
			return err
		}
	}
	return nil
}
