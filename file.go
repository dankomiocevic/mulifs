// Copyright 2016 Danko Miocevic. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Author: Danko Miocevic

package main

import (
	"github.com/dankomiocevic/mulifs/store"
	"github.com/dankomiocevic/mulifs/tools"
	"github.com/golang/glog"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"bazil.org/fuse/fuseutil"
	"golang.org/x/net/context"
)

// File defines a file in the filesystem structure.
// files can be Songs or .description files.
// Songs are actual songs in the Music Library and
// .description files detail more information about the
// Directory they are located in.
type File struct {
	artist string
	album  string
	song   string
	name   string
	mPoint string

	// To store special files in the DB.
	mu      sync.Mutex
	writers uint
	data    []byte
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	glog.Infof("Entering file Attr with name: %s.\n", f.name)
	if f.name[0] == '.' {
		if f.name == ".description" {
			descriptionJson, err := store.GetDescription(f.artist, f.album, f.name)
			if err != nil {
				return err
			}

			a.Size = uint64(len(descriptionJson))
			a.Mode = 0444
		} else {
			f.mu.Lock()
			defer f.mu.Unlock()

			a.Mode = 0666
			a.Size = uint64(len(f.data))
			if f.writers == 0 {
				// This is read only.
				// Get the real size from the DB
				b, err := store.GetSpecialFile(f.artist, f.album, f.name)
				if err == nil {
					a.Size = uint64(len(b))
				}
				return err
			}
			return nil
		}
	} else {
		songPath, err := store.GetFilePath(f.artist, f.album, f.name)
		if err != nil {
			return err
		}

		r, err := os.Open(songPath)
		if err != nil {
			return err
		}
		defer r.Close()

		fi, err := r.Stat()
		if err != nil {
			return err
		}

		a.Size = uint64(fi.Size())
		a.Mode = 0777
	}
	return nil
}

var _ = fs.NodeOpener(&File{})

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	glog.Infof("Entered Open with file name: %s.\n", f.name)

	if f.name == ".description" {
		return &FileHandle{r: nil, f: f}, nil
	}

	if runtime.GOOS == "darwin" {
		resp.Flags |= fuse.OpenDirectIO
	}

	if req.Flags.IsReadOnly() {
		glog.Info("Open: File requested is read only.\n")
		if runtime.GOOS == "darwin" {
			// If we are dealing with MAC Special files.
			if len(f.name) > 1 && f.name[0] == '.' && (f.name[1] == '_' || f.name == ".DS_Store") {
				return &FileHandle{r: nil, f: f}, nil
			}
		}
	}
	if req.Flags.IsReadWrite() {
		glog.Info("Open: File requested is read write.\n")
	}
	if req.Flags.IsWriteOnly() {
		glog.Info("Open: File requested is write only.\n")
	}

	if runtime.GOOS == "darwin" {
		// If we are dealing with MAC Special files.
		if len(f.name) > 1 && f.name[0] == '.' && (f.name[1] == '_' || f.name == ".DS_Store") {
			f.mu.Lock()
			defer f.mu.Unlock()

			if f.writers == 0 {
				data, err := store.GetSpecialFile(f.artist, f.album, f.name)
				if err != nil {
					return nil, err
				}
				f.data = append([]byte(nil), data...)
				f.writers++
				return f, nil
			}
		}
	}

	songPath, err := store.GetFilePath(f.artist, f.album, f.name)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	r, err := os.Open(songPath)
	if err != nil {
		return nil, err
	}
	return &FileHandle{r: r, f: f}, nil
}

type FileHandle struct {
	r *os.File
	f *File
}

var _ fs.Handle = (*FileHandle)(nil)

var _ fs.HandleReleaser = (*FileHandle)(nil)

func (fh *FileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	if fh.r == nil {
		if fh.f.name == ".description" {
			glog.Infof("Entered Release: .description file\n")
			return nil
		}

		if len(fh.f.name) > 1 && fh.f.name[0] == '.' && (fh.f.name[1] == '_' || fh.f.name == ".DS_Store") {
			glog.Infof("Entered Release: Mac special file: %s\n", fh.f.name)
			if req.Flags.IsReadOnly() {
				// We are not tracking read only special files.
				glog.Info("File is Read only\n")
				return nil
			}

			fh.f.mu.Lock()
			defer fh.f.mu.Unlock()

			fh.f.writers--
			if fh.f.writers == 0 {
				fh.f.data = nil
			}
			return nil
		}
		return nil
	}

	if fh.r == nil {
		glog.Info("Release: There is no file handler.\n")
		return fuse.EIO
	}
	glog.Infof("Releasing the file: %s\n", fh.r.Name())

	// This is not an music file or this is a strange situation.
	if fh.f == nil || len(fh.f.artist) < 1 || len(fh.f.album) < 1 {
		glog.Info("Entered Release: Artist or Album not set.\n")
		return fh.r.Close()
	}

	glog.Infof("Entered Release: Artist: %s, Album: %s, Song: %s\n", fh.f.artist, fh.f.album, fh.f.name)
	ret_val := fh.r.Close()
	extension := filepath.Ext(fh.f.name)
	songPath, err := store.GetFilePath(fh.f.artist, fh.f.album, fh.f.name)
	if err != nil {
		return err
	}

	if fh.f.artist == "drop" {
		return ret_val
	}

	if fh.f.artist == "playlist" {
		return ret_val
	}

	if extension == ".mp3" {
		//TODO: Use the correct artist and album
		tools.SetMp3Tags(fh.f.artist, fh.f.album, fh.f.song, songPath)
	}
	return ret_val
}

var _ = fs.HandleReader(&FileHandle{})

func (fh *FileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	glog.Infof("Entered Read.\n")
	if fh.r == nil {
		if runtime.GOOS == "darwin" {
			// MAC Special files.
			if len(fh.f.name) > 1 && fh.f.name[0] == '.' && (fh.f.name[1] == '_' || fh.f.name == ".DS_Store") {
				glog.Infof("Mac special file: %s\n", fh.f.name)

				fh.f.mu.Lock()
				defer fh.f.mu.Unlock()

				var b []byte
				if fh.f.writers == 0 {
					var err error
					b, err = store.GetSpecialFile(fh.f.artist, fh.f.album, fh.f.name)
					if err != nil {
						return err
					}
				} else {
					b = fh.f.data
				}
				fuseutil.HandleRead(req, resp, b)
				return nil
			}
		}

		if fh.f.name == ".description" {
			glog.Info("Reading description file\n")
			if len(fh.f.artist) < 1 {
				return fuse.ENOENT
			}
			_, err := store.GetArtistPath(fh.f.artist)
			if err != nil {
				return err
			}

			if len(fh.f.album) > 1 {
				_, err = store.GetAlbumPath(fh.f.artist, fh.f.album)
				if err != nil {
					return err
				}
			}
			descBytes, err := store.GetDescription(fh.f.artist, fh.f.album, fh.f.name)
			if err != nil {
				return err
			}
			resp.Data = []byte(descBytes)
			return nil
		}
		glog.Info("There is no file handler.\n")
		return fuse.EIO
	}

	glog.Infof("Reading file: %s.\n", fh.r.Name())
	buf := make([]byte, req.Size)
	n, err := fh.r.Read(buf)
	resp.Data = buf[:n]
	if err != nil && err != io.EOF {
		glog.Error(err)
		return err
	}
	return nil
}

var _ = fs.HandleWriter(&FileHandle{})

const maxInt = int(^uint(0) >> 1)

func (fh *FileHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	glog.Infof("Entered Write\n")
	if fh.r == nil {
		if runtime.GOOS == "darwin" {
			// MAC Special files.
			if len(fh.f.name) > 1 && fh.f.name[0] == '.' && (fh.f.name[1] == '_' || fh.f.name == ".DS_Store") {
				glog.Infof("Mac special file: %s\n", fh.f.name)

				fh.f.mu.Lock()
				defer fh.f.mu.Unlock()

				// expand the buffer if necessary
				newLen := req.Offset + int64(len(req.Data))
				if newLen > int64(maxInt) {
					return fuse.Errno(syscall.EFBIG)
				}

				if newLen := int(newLen); newLen > len(fh.f.data) {
					fh.f.data = append(fh.f.data, make([]byte, newLen-len(fh.f.data))...)
				}

				n := copy(fh.f.data[req.Offset:], req.Data)
				resp.Size = n
				return nil
			}
		}

		if fh.f.name == ".description" {
			glog.Errorf("Not allowed to write description file.\n")
			//TODO: Allow to write description
			return fuse.EPERM
		}
		return fuse.EIO
	}

	glog.Infof("Writing file: %s.\n", fh.r.Name())
	n, err := fh.r.Write(req.Data)
	resp.Size = n
	return err
}

var _ = fs.HandleFlusher(&FileHandle{})

func (fh *FileHandle) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	if fh.f != nil {
		glog.Infof("Entered Flush with Song: %s, Artist: %s and Album: %s\n", fh.f.name, fh.f.artist, fh.f.album)
	}

	if fh.r == nil {
		if runtime.GOOS == "darwin" {
			// MAC Special files.
			if len(fh.f.name) > 1 && fh.f.name[0] == '.' && (fh.f.name[1] == '_' || fh.f.name == ".DS_Store") {
				fh.f.mu.Lock()
				defer fh.f.mu.Unlock()

				glog.Infof("File has %d writers.", fh.f.writers)
				if fh.f.writers == 0 {
					glog.Info("This is a read only special file.")
					return nil
				}
				glog.Info("Writing special file on DB.")
				err := store.PutSpecialFile(fh.f.artist, fh.f.album, fh.f.name, fh.f.data)
				return err
			}
		}

		glog.Infof("There is no file handler.\n")
		return fuse.EIO
	}

	glog.Infof("Entered Flush with path: %s\n", fh.r.Name())

	fh.r.Sync()
	return nil
}

var _ = fs.NodeSetattrer(&File{})

func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	glog.Infof("Entered SetAttr with Song: %s, Artist: %s and Album: %s\n", f.name, f.artist, f.album)

	if runtime.GOOS == "darwin" {
		// MAC Special files.
		if len(f.name) > 1 && f.name[0] == '.' && (f.name[1] == '_' || f.name == ".DS_Store") {
			f.mu.Lock()
			defer f.mu.Unlock()
			if req.Valid.Size() {
				if req.Size > uint64(maxInt) {
					return fuse.Errno(syscall.EFBIG)
				}

				// Resize the data array.
				newLen := int(req.Size)
				switch {
				case newLen > len(f.data):
					f.data = append(f.data, make([]byte, newLen-len(f.data))...)
				case newLen < len(f.data):
					f.data = f.data[:newLen]
				}
				return nil
			}
		}
	}

	if req.Valid.Size() {
		glog.Infof("New size: %d\n", int(req.Size))
	}
	return nil
}
