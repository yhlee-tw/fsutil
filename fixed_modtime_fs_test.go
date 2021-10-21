package fsutil

import (
	"fmt"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/yhlee-tw/fsutil/internal/fsutiltest"
)

func TestOsDirFS(t *testing.T) {
	// given
	mt := time.Date(2021, time.October, 10, 23, 0, 0, 0, time.UTC)
	fsys := FixedModTimeFS(os.DirFS("internal/fsutiltest"), mt)

	// when
	f, err := fsys.Open("test_data.txt")
	if err != nil {
		t.Fatalf("Open: unexpected error %v", err)
	}
	fi, err := f.Stat()
	if err != nil {
		t.Fatalf("Stat: unexpected error %v", err)
	}
	actual := fi.ModTime()

	// then
	if actual != mt {
		t.Errorf("ModTime: expected %v got %v", mt, actual)
	}
}

func TestEmbedFS(t *testing.T) {
	// given
	mt := time.Date(2021, time.October, 10, 23, 0, 0, 0, time.UTC)
	fsys := FixedModTimeFS(fsutiltest.TestEmbeddedFS, mt)

	// when
	f, err := fsys.Open("test_data.txt")
	if err != nil {
		t.Fatalf("Open: unexpected error %v", err)
	}
	fi, err := f.Stat()
	if err != nil {
		t.Fatalf("Stat: unexpected error %v", err)
	}
	actual := fi.ModTime()

	// then
	if actual != mt {
		t.Errorf("ModTime: expected %v got %v", mt, actual)
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
