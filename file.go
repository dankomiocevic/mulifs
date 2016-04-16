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
	"errors"
	"fmt"
	"github.com/dankomiocevic/mulifs/musicmgr"
	"github.com/dankomiocevic/mulifs/playlistmgr"
	"github.com/dankomiocevic/mulifs/store"
	"github.com/golang/glog"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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

/** This function is used to do nothing to the file
 *	but to update the Touch time.
 */
func DelayedVoid(f File) error {
	return nil
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	glog.Infof("Entering file Attr with name: %s, Artist: %s and Album: %s.\n", f.name, f.artist, f.album)
	if f.name[0] == '.' {
		if f.name == ".description" {
			descriptionJson, err := store.GetDescription(f.artist, f.album, f.name)
			if err != nil {
				return err
			}

			a.Size = uint64(len(descriptionJson))
			a.Mode = 0444
		} else {
			return fuse.EPERM
		}
	} else {
		var songPath string
		var err error
		if f.artist == "drop" {
			songPath, err = store.GetDropFilePath(f.name, f.mPoint)
			PushFileItem(*f, nil)
		} else if f.artist == "playlists" {
			songPath, err = store.GetPlaylistFilePath(f.album, f.name, f.mPoint)
			PushFileItem(*f, nil)
		} else {
			songPath, err = store.GetFilePath(f.artist, f.album, f.name)
		}

		if err != nil {
			glog.Infof("Error getting song path: %s\n", err)
			return err
		}

		r, err := os.Open(songPath)
		if err != nil {
			glog.Infof("Error opening file: %s\n", err)
			return err
		}
		defer r.Close()

		fi, err := r.Stat()
		if err != nil {
			glog.Infof("Error getting file status: %s\n", err)
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

	if f.name[0] == '.' {
		return nil, fuse.EPERM
	}

	if runtime.GOOS == "darwin" {
		resp.Flags |= fuse.OpenDirectIO
	}

	if req.Flags.IsReadOnly() {
		glog.Info("Open: File requested is read only.\n")
	}
	if req.Flags.IsReadWrite() {
		glog.Info("Open: File requested is read write.\n")
	}
	if req.Flags.IsWriteOnly() {
		glog.Info("Open: File requested is write only.\n")
	}

	var err error
	var songPath string
	if f.artist == "drop" {
		songPath, err = store.GetDropFilePath(f.name, f.mPoint)
		PushFileItem(*f, DelayedVoid)
	} else if f.artist == "playlists" {
		songPath, err = store.GetPlaylistFilePath(f.album, f.name, f.mPoint)
		PushFileItem(*f, DelayedVoid)
	} else {
		songPath, err = store.GetFilePath(f.artist, f.album, f.name)
	}
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	r, err := os.OpenFile(songPath, int(req.Flags), 0666)
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

// DelayedHandlePlaylistSong handles a dropped file
// inside a playlist and is called by the background
// dispatcher after some time has passed.
func DelayedHandlePlaylistSong(f File) error {
	rootPoint := f.mPoint
	if rootPoint[len(rootPoint)-1] != '/' {
		rootPoint = rootPoint + "/"
	}

	path := rootPoint + "playlists/" + f.album + "/" + f.name

	extension := filepath.Ext(f.name)
	if extension != ".mp3" {
		os.Remove(path)
		return errors.New("File is not an mp3.")
	}

	src, err := os.Stat(path)
	if err != nil || src.IsDir() {
		return errors.New("File not found.")
	}

	err, tags := musicmgr.GetMp3Tags(path)
	if err != nil {
		os.Remove(path)
		return err
	}

	artist := store.GetCompatibleString(tags.Artist)
	album := store.GetCompatibleString(tags.Album)
	title := tags.Title
	if strings.HasSuffix(title, ".mp3") {
		title = title[:len(title)-len(".mp3")]
	}
	title = store.GetCompatibleString(title) + ".mp3"

	newPath, err := store.GetFilePath(artist, album, title)
	if err == nil {
		var playlistFile playlistmgr.PlaylistFile
		playlistFile.Title = title
		playlistFile.Artist = artist
		playlistFile.Album = album
		playlistFile.Path = newPath
		err = store.AddFileToPlaylist(playlistFile, f.album)
	} else {
		err = store.HandleDrop(path, rootPoint)
		if err == nil {
			newPath, err = store.GetFilePath(artist, album, title)
			if err == nil {
				var playlistFile playlistmgr.PlaylistFile
				playlistFile.Title = title
				playlistFile.Artist = artist
				playlistFile.Album = album
				playlistFile.Path = newPath
				err = store.AddFileToPlaylist(playlistFile, f.album)
			}
		}
	}

	os.Remove(path)
	if err == nil {
		err = store.RegeneratePlaylistFile(f.album, rootPoint)
	}
	return err
}

// DelayedHandleDrop handles a dropped file
// but is called by the background dispatcher
// after some time has passed.
func DelayedHandleDrop(f File) error {
	// Get the dropped file path.
	rootPoint := f.mPoint
	if rootPoint[len(rootPoint)-1] != '/' {
		rootPoint = rootPoint + "/"
	}

	path := rootPoint + "drop/" + f.name
	err := store.HandleDrop(path, rootPoint)
	fmt.Printf("DelayedHandleDrop: %s\n", path)
	if err != nil {
		glog.Error(err)
		return err
	}
	return nil
}

var _ fs.HandleReleaser = (*FileHandle)(nil)

func (fh *FileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	if fh.r == nil {
		if fh.f.name == ".description" {
			glog.Infof("Entered Release: .description file\n")
			return nil
		}

		if fh.f.name[0] == '.' {
			return fuse.EPERM
		}
	}

	if fh.r == nil {
		glog.Info("Release: There is no file handler.\n")
		return fuse.EIO
	}
	glog.Infof("Releasing the file: %s\n", fh.r.Name())

	if fh.f != nil && fh.f.artist == "drop" {
		glog.Infof("Entered Release dropping the song: %s\n", fh.f.name)
		ret_val := fh.r.Close()

		PushFileItem(*fh.f, DelayedHandleDrop)
		return ret_val
	}

	if fh.f != nil && fh.f.artist == "playlists" {
		glog.Infof("Entered Release with playlist song: %s\n", fh.f.name)
		ret_val := fh.r.Close()

		PushFileItem(*fh.f, DelayedHandlePlaylistSong)
		return ret_val
	}

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

	if extension == ".mp3" {
		//TODO: Use the correct artist and album
		musicmgr.SetMp3Tags(fh.f.artist, fh.f.album, fh.f.song, songPath)
	}
	return ret_val
}

var _ = fs.HandleReader(&FileHandle{})

func (fh *FileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	glog.Infof("Entered Read.\n")
	//TODO: Check if we need to add something here for playlists and drop directories.
	if fh.r == nil {
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
	if _, err := fh.r.Seek(req.Offset, 0); err != nil {
		return err
	}
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

func (fh *FileHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	glog.Infof("Entered Write\n")
	//TODO: Check if we need to add something here for playlists and drop directories.
	if fh.r == nil {
		if fh.f.name == ".description" {
			glog.Errorf("Not allowed to write description file.\n")
			//TODO: Allow to write description
			return nil
		}
		return fuse.EIO
	}

	glog.Infof("Writing file: %s.\n", fh.r.Name())
	if _, err := fh.r.Seek(req.Offset, 0); err != nil {
		return err
	}
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

	if req.Valid.Size() {
		glog.Infof("New size: %d\n", int(req.Size))
	}
	return nil
}
