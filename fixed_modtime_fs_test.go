package fsutil

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/yhlee-tw/fsutil/internal/fsutiltest"
)

type mockFS struct {
	seek func(offset int64, whence int) (int64, error)
}

func (fsys *mockFS) Open(name string) (fs.File, error) {
	f := mockFile{name, []byte("test\n"), 0}
	if fsys.seek != nil {
		return &mockSeekableFile{f, fsys.seek}, nil
	}
	return &f, nil
}

type mockFile struct {
	name   string
	b      []byte
	offset int64
}

func (f *mockFile) Close() error       { return nil }
func (f *mockFile) Name() string       { return "test_data.txt" }
func (f *mockFile) Size() int64        { return int64(len(f.b)) }
func (f *mockFile) Mode() fs.FileMode  { return 0444 }
func (f *mockFile) ModTime() time.Time { return time.Time{} }
func (f *mockFile) IsDir() bool        { return false }
func (f *mockFile) Sys() interface{}   { return nil }

func (f *mockFile) Stat() (fs.FileInfo, error) { return f, nil }

func (f *mockFile) Read(b []byte) (int, error) {
	if f.offset >= int64(len(f.b)) {
		return 0, io.EOF
	}
	n := copy(b, f.b[f.offset:])
	f.offset += int64(n)
	return n, nil
}

type mockSeekableFile struct {
	mockFile
	seek func(offset int64, whence int) (int64, error)
}

func (f *mockSeekableFile) Seek(offset int64, whence int) (int64, error) {
	return f.seek(offset, whence)
}

func TestModTime(t *testing.T) {
	// given
	tests := []struct {
		name     string
		fsys     fs.FS
		filename string
		open_ok  bool
		seekable bool
	}{
		{"mockFS", &mockFS{}, "test_data.txt", true, false},
		{"fstest.MapFS", fstest.MapFS{"test.txt": &fstest.MapFile{Data: []byte("test\n")}}, "test.txt", true, true},
		{"os.DirFS", os.DirFS("internal/fsutiltest"), "test_data.txt", true, true},
		{"os.DirFS/Dir", os.DirFS("internal"), "fsutiltest", true, false},
		{"embed.FS/bad_file", fsutiltest.TestEmbeddedFS, "404", false, false},
		{"embed.FS", fsutiltest.TestEmbeddedFS, "test_data.txt", true, true},
	}
	mt := time.Date(2021, time.October, 10, 23, 0, 0, 0, time.UTC)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := FixedModTimeFS(tt.fsys, mt)

			f, err := fsys.Open(tt.filename)
			if tt.open_ok {
				if err != nil {
					t.Fatalf("Open: unexpected error %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("Open: expect error but did not get it")
				}
				return
			}
			if _, ok := f.(io.Seeker); ok != tt.seekable {
				t.Logf("File: expected seekable %v got %v", tt.seekable, ok)
			}
			fi, err := f.Stat()
			if err != nil {
				t.Fatalf("Stat: unexpected error %v", err)
			}
			actual := fi.ModTime()
			if actual != mt {
				t.Errorf("ModTime: expected %v got %v", mt, actual)
			}
		})
	}
}

func TestSeeker(t *testing.T) {
	mt := time.Date(2021, time.October, 10, 23, 0, 0, 0, time.UTC)
	called := false
	mock_seek := func(offset int64, whence int) (int64, error) {
		called = true
		return 0, nil
	}
	fsys := FixedModTimeFS(&mockFS{mock_seek}, mt)

	f, err := fsys.Open("test_data.txt")
	if err != nil {
		t.Fatalf("Open: unexpected error %v", err)
	}
	sk, ok := f.(io.Seeker)
	if !ok {
		t.Fatal("File: expected seekable but is not")
	}
	actual, err := sk.Seek(0, 0)
	if err != nil {
		t.Fatalf("Seek: unexpected error %v", err)
	}
	if actual != 0 {
		t.Errorf("Seek: expected return 0 got %v", actual)
	}
	if !called {
		t.Errorf("Seek: mock_seek was called but was not")
	}
}

func ExampleFixedModTimeFS() {
	// embed.FS returns zero ModTime
	var fsys fs.FS = fsutiltest.TestEmbeddedFS
	f, _ := fsys.Open("test_data.txt")
	fi, _ := f.Stat()
	fmt.Printf("embed.FS: %v\n", fi.ModTime())

	// FixedModTimeFS returns fixed ModTime
	mt := time.Date(2021, time.October, 10, 23, 0, 0, 0, time.UTC)
	fsys = FixedModTimeFS(fsys, mt)
	f, _ = fsys.Open("test_data.txt")
	fi, _ = f.Stat()
	fmt.Printf("FixedModTimeFS: %v\n", fi.ModTime())
	// Output:
	// embed.FS: 0001-01-01 00:00:00 +0000 UTC
	// FixedModTimeFS: 2021-10-10 23:00:00 +0000 UTC
}
