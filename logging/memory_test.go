package logging

import (
	"testing"
)

func TestReadAfterWrite(t *testing.T) {
	data := []string{"first", "second", "third"}
	m := NewMemoryLogger(10).(*memoryLogs)
	for i, d := range data {
		t.Logf("#%d before: Write index: %d, start index: %d\n", i, m.w, m.start)
		_, err := m.Write([]byte(d))
		if err != nil {
			t.Fatalf("Failed to write string %d: %s", i, err)
		}
		t.Logf("#%d after: Write index: %d (%t), start index: %d (%t)\n", i, m.w, m.wCycle, m.start, m.startCycle)
	}
	// Read back
	var c collector
	if err := m.Export(&c, false); err != nil {
		t.Fatalf("Failed to export logs from memory: %s", err)
	}
	if len(c) != len(data) {
		t.Errorf("Data missing: %d written, %d exported", len(data), len(c))
	}
	for i, a := range c {
		if a != data[i] {
			t.Errorf("Bad value for entry %d: got '%s' instead of '%s'", i, a, data[i])
		}
	}
}

func TestReadAfterWrappedWrite(t *testing.T) {
	data := []string{"first", "second", "third"}
	m := NewMemoryLogger(2).(*memoryLogs)
	for i, d := range data {
		t.Logf("#%d before: Write index: %d, read index: %d\n", i, m.w, m.start)
		_, err := m.Write([]byte(d))
		if err != nil {
			t.Fatalf("Failed to write string %d: %s", i, err)
		}
		t.Logf("#%d after: Write index: %d (%t), read index: %d (%t)\n", i, m.w, m.wCycle, m.start, m.startCycle)
	}
	// Read back
	var c collector
	if err := m.Export(&c, false); err != nil {
		t.Fatalf("Failed to export logs from memory: %s", err)
	}
	expected := data[1:]
	if len(c) != len(expected) {
		t.Errorf("Data missing: %d written, %d exported", len(expected), len(c))
	}
	for i, a := range c {
		if a != expected[i] {
			t.Errorf("Bad value for entry %d: got '%s' instead of '%s'", i, a, expected[i])
		}
	}
}

func TestReadReverseAfterWrappedWrite(t *testing.T) {
	data := []string{"first", "second", "third"}
	m := NewMemoryLogger(2).(*memoryLogs)
	for i, d := range data {
		t.Logf("#%d before: Write index: %d, read index: %d\n", i, m.w, m.start)
		_, err := m.Write([]byte(d))
		if err != nil {
			t.Fatalf("Failed to write string %d: %s", i, err)
		}
		t.Logf("#%d after: Write index: %d (%t), read index: %d (%t)\n", i, m.w, m.wCycle, m.start, m.startCycle)
	}
	// Read back
	var c collector
	if err := m.Export(&c, true); err != nil {
		t.Fatalf("Failed to export logs from memory: %s", err)
	}
	expected := data[1:]
	if len(c) != len(expected) {
		t.Errorf("Data missing: %d written, %d exported", len(expected), len(c))
	}
	for i, a := range c {
		e := expected[len(expected)-i-1]
		if a != e {
			t.Errorf("Bad value for entry %d: got '%s' instead of '%s'", i, a, e)
		}
	}
}

type collector []string

func (c *collector) Write(data []byte) (int, error) {
	*c = append(*c, string(data))
	return len(data), nil
}
