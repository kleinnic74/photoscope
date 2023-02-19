package library

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromSeekableReader(t *testing.T) {
	data := []struct {
		Name string
		Load func(in io.Reader) io.Reader
		Dir  func(*testing.T, []fs.DirEntry)
	}{
		{
			Name: "Reader is an *os.File",
			Load: func(in io.Reader) io.Reader { return in },
			Dir: func(t *testing.T, entries []fs.DirEntry) {
				if len(entries) != 0 {
					t.Errorf("Too many entries in temporary directory, should be empty. Found: %v", entries)
				}
			},
		},
		{
			Name: "Reader is not seekable",
			Load: func(in io.Reader) io.Reader { return wrappedReader{in} },
			Dir: func(t *testing.T, entries []fs.DirEntry) {
				if len(entries) != 1 {
					t.Errorf("Too many entries in temporary directory, should contain only one. Found: %v", entries)
				}
				if entries[0].Name() != TestImageName {
					t.Errorf("Bad file name in directory, expected XXx, got %s", entries[0].Name())
				}
			},
		},
	}
	for i, d := range data {
		t.Run(fmt.Sprintf("#%d %s", i, d.Name), func(t *testing.T) {
			tmpdir, err := os.MkdirTemp(".", "test-photoscope-*")
			if err != nil {
				t.Fatalf("Failed to initialize loader: %s", err)
			}
			defer func() { os.Remove(tmpdir) }()

			// Verify that media objects are cleaned up after processing
			defer func() {
				assertFilesInDir(t, tmpdir, func(t *testing.T, de []fs.DirEntry) {
					if len(de) != 0 {
						t.Errorf("After cleaning up the media object, no files shall be left in tmp dir but there were: %v", de)
					}
				})
			}()

			loader := NewLoader(tmpdir)

			// Load the expected photo content
			in, expectedContent, err := loadTestPhoto(TestImageName)
			if err != nil {
				t.Fatalf("Failed to load test data: %s", err)
			}
			defer in.Close()

			// Modify the reader as per the test
			testedReader := d.Load(in)
			media, err := loader.LoadMediaObject(TestImageName, testedReader)
			if err != nil {
				t.Fatalf("Loader failed to load file: %s", err)
			}
			defer media.Cleanup()

			// Check contents of temporary directory
			assertFilesInDir(t, tmpdir, d.Dir)

			// Ensure processed media content is the same as of the loaded file
			var actualContent []byte
			if err := media.ProcessContent(func(r io.Reader) error {
				actualContent, err = ioutil.ReadAll(r)
				return err
			}); err != nil {
				t.Fatalf("Failed to process media object content: %s", err)
			}
			if c := bytes.Compare(expectedContent, actualContent); c != 0 {
				t.Errorf("Content of media object does not match expected content (%d)", c)
			}
		})
	}
}

func assertFilesInDir(t *testing.T, dir string, assertion func(*testing.T, []fs.DirEntry)) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read temporary directory %s: %s", dir, err)
		return
	}
	assertion(t, entries)
}

func loadTestPhoto(name string) (io.ReadCloser, []byte, error) {
	basedir, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}
	path := filepath.Join(filepath.Clean(filepath.Join(basedir, "..")), "domain", "testdata", name)
	in, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	content, err := ioutil.ReadAll(in)
	if err != nil {
		in.Close()
		return in, nil, err
	}
	if _, err := in.Seek(0, io.SeekStart); err != nil {
		in.Close()
		return in, nil, err
	}
	return in, content, err
}

type wrappedReader struct {
	r io.Reader
}

func (w wrappedReader) Read(p []byte) (n int, err error) {
	return w.r.Read(p)
}

const TestImageName = "Canon_40D.jpg"
