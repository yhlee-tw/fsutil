// Package fsutil provides fs.FS wrapper(s).
package fsutil

import (
	"io"
	"io/fs"
	"time"
)

// FixedModTimeFS converts fsys to an fs.FS implementation that always returns
// a fixed ModTime.
//
// Primary for use with go:embed FS that does not provide ModTime info.
func FixedModTimeFS(fsys fs.FS, fixedModTime time.Time) fs.FS {
	return &fixedModTimeFS{FS: fsys, fixedModTime: fixedModTime}
}

type fixedModTimeFS struct {
	fs.FS
	fixedModTime time.Time
}

func (f *fixedModTimeFS) Open(name string) (fs.File, error) {
	file, err := f.FS.Open(name)
	if err != nil {
		return nil, err
	}
	if fi, err := file.Stat(); err == nil && fi.IsDir() {
		if dirfile, ok := file.(fs.ReadDirFile); ok {
			return &fixedModTimeDirFile{ReadDirFile: dirfile, fixedModTime: f.fixedModTime}, err
		}
	}
	wrapped := fixedModTimeFile{File: file, fixedModTime: f.fixedModTime}
	if s, ok := file.(io.Seeker); ok {
		return &fixedModTimeSeekableFile{fixedModTimeFile: wrapped, seek: s.Seek}, err
	}
	return &wrapped, err
}

type fixedModTimeFile struct {
	fs.File
	fixedModTime time.Time
}

func (f *fixedModTimeFile) Stat() (fs.FileInfo, error) {
	fileinfo, err := f.File.Stat()
	return &fixedModTimeFileInfo{FileInfo: fileinfo, fixedModTime: f.fixedModTime}, err
}

type fixedModTimeSeekableFile struct {
	fixedModTimeFile
	seek func(offset int64, whence int) (int64, error)
}

func (f *fixedModTimeSeekableFile) Seek(offset int64, whence int) (int64, error) {
	return f.seek(offset, whence)
}

type fixedModTimeDirFile struct {
	fs.ReadDirFile
	fixedModTime time.Time
}

func (f *fixedModTimeDirFile) Stat() (fs.FileInfo, error) {
	fileinfo, err := f.ReadDirFile.Stat()
	return &fixedModTimeFileInfo{FileInfo: fileinfo, fixedModTime: f.fixedModTime}, err
}

type fixedModTimeFileInfo struct {
	fs.FileInfo
	fixedModTime time.Time
}

func (f *fixedModTimeFileInfo) ModTime() time.Time {
	return f.fixedModTime
}
