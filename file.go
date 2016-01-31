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
	"os"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
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
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	glog.Infof("Entering file Attr with name: %s.\n", f.name)
	if f.name == ".description" {
		descriptionJson, err := store.GetDescription(f.artist, f.album, f.name)
		if err != nil {
			return err
		}

		a.Size = uint64(len(descriptionJson))
		a.Mode = 0444
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

	songPath, err := store.GetFilePath(f.artist, f.album, f.name)
	if err != nil {
		return nil, err
	}

	r, err := os.Open(songPath)
	if err != nil {
		return nil, err
	}
	resp.Flags |= fuse.OpenNonSeekable
	return &FileHandle{r: r, f: nil}, nil
}

type FileHandle struct {
	r *os.File
	f *File
}

var _ fs.Handle = (*FileHandle)(nil)

var _ fs.HandleReleaser = (*FileHandle)(nil)

func (fh *FileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	glog.Infof("Entered Release: Artist: %s, Album: %s, Song: %s\n", fh.f.artist, fh.f.album, fh.f.name)
	if fh.r == nil {
		if fh.f.name == ".description" {
			return nil
		}
	}

	if req.Flags.IsReadOnly() {
		// we don't need to track read-only handles
		return nil
	}
	// This is not an music file or this is a strange situation.
	if len(fh.f.artist) < 1 || len(fh.f.album) < 1 {
		return fh.r.Close()
	}

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
		if fh.f.name == ".description" {
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
		return nil
	}
	buf := make([]byte, req.Size)
	n, err := fh.r.Read(buf)
	resp.Data = buf[:n]
	return err
}

var _ = fs.HandleWriter(&FileHandle{})

func (fh *FileHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	glog.Infof("Entered Write\n")
	if fh.r == nil {
		glog.Errorf("There is no file handler.\n")
		if fh.f.name == ".description" {
			//TODO: Allow to write description
			return fuse.EPERM
		}
		return nil
	}

	glog.Infof("Writing data: %s\n", string(req.Data))
	n, err := fh.r.Write(req.Data)
	resp.Size = n
	return err
}

var _ = fs.HandleFlusher(&FileHandle{})

func (fh *FileHandle) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	glog.Infof("Entered Flush\n")
	if fh.r == nil {
		glog.Infof("There is no file handler.\n")
	}
	fh.r.Sync()
	return nil
}

var _ = fs.NodeSetattrer(&File{})

func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	glog.Infof("Entered SetAttr\n")

	if req.Valid.Size() {
		glog.Infof("New size: %d\n", int(req.Size))
	}
	return nil
}
