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
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

// Dir struct specifies a Directory in the
// filesystem, the Directories can be Artist or Albums.
// The root Directory that contains all the Artists
// is also a Directory.
type Dir struct {
	fs     *FS
	artist string
	album  string
	mPoint string
}

var _ = fs.Node(&Dir{})

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	glog.Infof("Entered Attr dir: Artist: %s, Album: %s\n", d.artist, d.album)
	a.Mode = os.ModeDir | 0777
	return nil
}

var dirDirs = []fuse.Dirent{
	{Name: "drop", Type: fuse.DT_Dir},
	{Name: "playlists", Type: fuse.DT_Dir},
}

var _ = fs.NodeStringLookuper(&Dir{})

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	glog.Infof("Entering Lookup with artist: %s, album: %s and name: %s.\n", d.artist, d.album, name)
	if name == ".description" {
		return &File{artist: d.artist, album: d.album, song: name, name: name, mPoint: d.mPoint}, nil
	}

	if name[0] == '.' {
		return nil, fuse.EIO
	}

	if len(d.artist) < 1 {
		if name == "drop" {
			return &Dir{fs: d.fs, artist: "drop", album: "", mPoint: d.mPoint}, nil
		}
		if name == "playlists" {
			return &Dir{fs: d.fs, artist: "playlists", album: "", mPoint: d.mPoint}, nil
		}

		_, err := store.GetArtistPath(name)
		if err != nil {
			glog.Info(err)
			return nil, err
		}
		return &Dir{fs: d.fs, artist: name, album: "", mPoint: d.mPoint}, nil
	}

	if len(d.album) < 1 && d.artist != "drop" && d.artist != "playlists" {
		_, err := store.GetAlbumPath(d.artist, name)
		if err != nil {
			glog.Info(err)
			return nil, err
		}
		return &Dir{fs: d.fs, artist: d.artist, album: name, mPoint: d.mPoint}, nil
	}

	var err error
	if d.artist == "drop" {
		_, err = store.GetDropFilePath(name, d.mPoint)
		if err != nil {
			glog.Info(err)
			return nil, fuse.ENOENT
		}
	} else if d.artist == "playlists" {
		if len(d.album) < 1 {
			_, err = store.GetPlaylistPath(name)
			if err != nil {
				glog.Info(err)
				return nil, fuse.ENOENT
			}
			return &Dir{fs: d.fs, artist: d.artist, album: name, mPoint: d.mPoint}, nil
		} else {
			_, err = store.GetPlaylistFilePath(d.album, name, d.mPoint)
			if err != nil {
				glog.Info(err)
				return nil, fuse.ENOENT
			}
		}
	} else {
		_, err = store.GetFilePath(d.artist, d.album, name)
		if err != nil {
			glog.Info(err)
			return nil, err
		}
	}
	extension := filepath.Ext(name)
	songName := name[:len(name)-len(extension)]
	return &File{artist: d.artist, album: d.album, song: songName, name: name, mPoint: d.mPoint}, nil
}

var _ = fs.HandleReadDirAller(&Dir{})

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	glog.Infof("Entering ReadDirAll\n")
	if len(d.artist) < 1 {
		a, err := store.ListArtists()
		if err != nil {
			return nil, fuse.ENOENT
		}
		for _, v := range dirDirs {
			a = append(a, v)
		}
		return a, nil
	}

	if d.artist == "drop" {
		if len(d.album) > 0 {
			return nil, fuse.ENOENT
		}

		rootPoint := d.mPoint
		if rootPoint[len(rootPoint)-1] != '/' {
			rootPoint = rootPoint + "/"
		}

		path := rootPoint + "drop"
		// Check if the drop directory exists
		src, err := os.Stat(path)
		if err != nil {
			return nil, nil
		}

		// Check if it is a directory
		if !src.IsDir() {
			return nil, nil
		}

		var a []fuse.Dirent
		files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			var node fuse.Dirent
			node.Name = f.Name()
			node.Type = fuse.DT_File
			a = append(a, node)
		}
		return a, nil
	}

	if d.artist == "playlists" {
		if len(d.album) < 1 {
			a, err := store.ListPlaylists()
			if err != nil {
				return nil, fuse.ENOENT
			}
			return a, nil
		}

		a, err := store.ListPlaylistSongs(d.album, d.mPoint)
		if err != nil {
			return nil, fuse.ENOENT
		}

		return a, nil
	}

	if len(d.album) < 1 {
		a, err := store.ListAlbums(d.artist)
		if err != nil {
			return nil, fuse.ENOENT
		}
		return a, nil
	}

	a, err := store.ListSongs(d.artist, d.album)
	if err != nil {
		return nil, fuse.ENOENT
	}

	return a, nil
}

var _ = fs.NodeMkdirer(&Dir{})

func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	name := req.Name
	glog.Infof("Entering mkdir with name: %s.\n", name)
	// Do not allow creating directories starting with dot
	if name[0] == '.' {
		glog.Info("Names starting with dot are not allowed.")
		return nil, fuse.EPERM
	}

	if d.mPoint[len(d.mPoint)-1] != '/' {
		d.mPoint = d.mPoint + "/"
	}
	if len(d.artist) < 1 {
		glog.Info("Creating an Artist.")
		ret, err := store.CreateArtist(name)
		if err != nil {
			glog.Infof("Error creating artist: %s\n", err)
			return nil, err
		}

		path := d.mPoint + ret
		err = os.MkdirAll(path, 0777)
		if err != nil {
			glog.Infof("Error creating artist folder: %s\n", err)
			return nil, fuse.EIO
		}
		return &Dir{fs: d.fs, artist: ret, album: "", mPoint: d.mPoint}, nil
	}

	if d.artist == "drop" {
		return nil, fuse.EIO
	}

	if d.artist == "playlists" {
		if len(d.album) < 1 {
			ret, err := store.CreatePlaylist(name, d.mPoint)
			if err != nil {
				glog.Infof("Error creating playlist: %s\n", err)
				return nil, err
			}

			err = store.RegeneratePlaylistFile(ret, d.mPoint)
			if err != nil {
				glog.Infof("Error regenerating playlist: %s\n", err)
				return nil, err
			}
			return &Dir{fs: d.fs, artist: "playlists", album: ret, mPoint: d.mPoint}, nil
		}
		return nil, fuse.EPERM
	}

	if len(d.album) < 1 {
		glog.Infof("Creating album: %s in artist: %s.\n", d.artist, name)
		ret, err := store.CreateAlbum(d.artist, name)
		if err != nil {
			return nil, err
		}
		path := d.mPoint + d.artist + "/" + ret
		err = os.MkdirAll(path, 0777)
		if err != nil {
			glog.Infof("Error creating artist folder: %s\n", err)
			return nil, fuse.EIO
		}
		return &Dir{fs: d.fs, artist: d.artist, album: ret, mPoint: d.mPoint}, nil
	}

	return nil, fuse.EIO
}

var _ = fs.NodeCreater(&Dir{})

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	glog.Infof("Entered Create Dir\n")

	if req.Flags.IsReadOnly() {
		glog.Info("Create: File requested is read only.\n")
	}
	if req.Flags.IsReadWrite() {
		glog.Info("Create: File requested is read write.\n")
	}
	if req.Flags.IsWriteOnly() {
		glog.Info("Create: File requested is write only.\n")
	}

	if runtime.GOOS == "darwin" {
		resp.Flags |= fuse.OpenDirectIO
	}

	if d.artist == "drop" {
		if len(d.album) > 0 {
			glog.Info("Subdirectories are not allowed in drop folder.")
			return nil, nil, fuse.EIO
		}

		rootPoint := d.mPoint
		if rootPoint[len(rootPoint)-1] != '/' {
			rootPoint = rootPoint + "/"
		}

		name := req.Name
		path := rootPoint + "drop/"
		extension := filepath.Ext(name)

		if extension != ".mp3" {
			glog.Info("Only mp3 files are allowed.")
			return nil, nil, fuse.EIO
		}

		// Check if the drop directory exists
		src, err := os.Stat(path)
		if err != nil || !src.IsDir() {
			err = os.MkdirAll(path, 0777)
			if err != nil {
				glog.Infof("Cannot create dir: %s\n", err)
				return nil, nil, err
			}
		}

		fi, err := os.Create(path + name)
		if err != nil {
			glog.Infof("Cannot create file: %s\n", err)
			return nil, nil, err
		}

		keyName := name[:len(name)-len(extension)]
		f := &File{
			artist: d.artist,
			album:  d.album,
			song:   keyName,
			name:   name,
			mPoint: d.mPoint,
		}

		if fi != nil {
			glog.Infof("Returning file handle for: %s.\n", fi.Name())
		}
		return f, &FileHandle{r: fi, f: f}, nil
	}

	if d.artist == "playlists" {
		if len(d.album) < 1 {
			glog.Info("Files are not allowed outside playlists.")
			return nil, nil, fuse.EIO
		}

		rootPoint := d.mPoint
		if rootPoint[len(rootPoint)-1] != '/' {
			rootPoint = rootPoint + "/"
		}

		name := req.Name
		path := rootPoint + "playlists/" + d.album
		extension := filepath.Ext(name)

		if extension != ".mp3" {
			glog.Info("Only mp3 files are allowed.")
			return nil, nil, fuse.EIO
		}

		// Check if the playlist drop directory exists
		src, err := os.Stat(path)
		if err != nil || !src.IsDir() {
			err = os.MkdirAll(path, 0777)
			if err != nil {
				glog.Infof("Cannot create dir: %s\n", err)
				return nil, nil, err
			}
		}

		fi, err := os.Create(path + "/" + name)
		if err != nil {
			glog.Infof("Cannot create file: %s\n", err)
			return nil, nil, err
		}

		keyName := name[:len(name)-len(extension)]
		f := &File{
			artist: d.artist,
			album:  d.album,
			song:   keyName,
			name:   name,
			mPoint: d.mPoint,
		}

		if fi != nil {
			glog.Infof("Returning file handle for: %s.\n", fi.Name())
		}
		return f, &FileHandle{r: fi, f: f}, nil
	}

	if len(d.artist) < 1 || len(d.album) < 1 {
		return nil, nil, fuse.EPERM
	}

	nameRaw := req.Name
	if nameRaw[0] == '.' {
		glog.Info("Cannot create files starting with dot.")
		return nil, nil, fuse.EPERM
	}

	rootPoint := d.mPoint
	if rootPoint[len(rootPoint)-1] != '/' {
		rootPoint = rootPoint + "/"
	}

	path := rootPoint + d.artist + "/" + d.album + "/"
	name, err := store.CreateSong(d.artist, d.album, nameRaw, path)
	if err != nil {
		glog.Info("Error creating song.")
		return nil, nil, fuse.EPERM
	}

	err = os.MkdirAll(path, 0777)
	if err != nil {
		glog.Info("Cannot create folder.")
		return nil, nil, err
	}

	fi, err := os.Create(path + name)
	if err != nil {
		glog.Infof("Cannot create file: %s\n", err)
		return nil, nil, err
	}

	extension := filepath.Ext(name)
	keyName := name[:len(name)-len(extension)]
	f := &File{
		artist: d.artist,
		album:  d.album,
		song:   keyName,
		name:   name,
		mPoint: d.mPoint,
	}

	if fi != nil {
		glog.Infof("Returning file handle for: %s.\n", fi.Name())
	}
	return f, &FileHandle{r: fi, f: f}, nil
}

var _ = fs.NodeRemover(&Dir{})

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	//TODO: Correct this function to work with drop folder.
	name := req.Name
	glog.Infof("Entered Remove function with Artist: %s, Album: %s and Name: %s.\n", d.artist, d.album, name)

	if name == ".description" {
		return nil
	}

	if name[0] == '.' {
		return fuse.EIO
	}

	if req.Dir {
		if len(name) < 1 {
			return fuse.EIO
		}

		if len(d.artist) < 1 {
			err := store.DeleteArtist(name)
			if err != nil {
				return fuse.EIO
			}

			return nil
		}

		if d.artist == "playlists" {
			store.DeletePlaylist(name, d.mPoint)
			return nil
		}

		err := store.DeleteAlbum(d.artist, name)
		if err != nil {
			return fuse.EIO
		}

		return nil
	} else {
		if len(d.artist) < 1 || len(d.album) < 1 {
			return fuse.EIO
		}

		fullPath, err := store.GetFilePath(d.artist, d.album, name)
		if err != nil {
			return fuse.EIO
		}

		err = store.DeleteSong(d.artist, d.album, name)
		if err != nil {
			return fuse.EIO
		}

		//TODO: Check if there are no more files in the folder
		//      and delete the folder.

		err = os.Remove(fullPath)
		if err != nil {
			return err
		}

		return nil
	}
}

var _ = fs.NodeRenamer(&Dir{})

func (d *Dir) Rename(ctx context.Context, r *fuse.RenameRequest, newDir fs.Node) error {
	var newD *Dir

	newD = newDir.(*Dir)
	glog.Infof("Renaming: OldName: %s, NewName: %s, newDir: %s/%s\n", r.OldName, r.NewName, newD.artist, newD.album)

	if d.mPoint[len(d.mPoint)-1] != '/' {
		d.mPoint = d.mPoint + "/"
	}
	path := d.mPoint + d.artist + "/" + d.album + "/" + r.OldName

	//if newDir != d {
	//	return fuse.Errno(syscall.EXDEV)
	//}

	if r.OldName == ".description" || r.NewName == ".description" {
		return fuse.EPERM
	}

	if r.NewName[0] == '.' || r.OldName[0] == '.' {
		glog.Info("Names starting with dot are not allowed.")
		return fuse.EPERM
	}

	if len(d.artist) < 1 {
		glog.Info("Changing artist name.")
		if len(newD.artist) > 0 {
			return fuse.EPERM
		}

		err := store.MoveArtist(r.OldName, r.NewName, d.mPoint)
		return err
	}

	if d.artist == "drop" {
		glog.Info("Cannot rename inside drop folder.")
		return fuse.EPERM
	}

	if d.artist == "playlists" {
		glog.Info("Rename inside playlists folder.")
		//TODO: Rename inside playlists.
		return nil
	}

	if len(d.album) < 1 {
		glog.Info("Moving album")
		if len(newD.album) > 0 {
			return fuse.EPERM
		}

		err := store.MoveAlbum(d.artist, r.OldName, newD.artist, r.NewName, d.mPoint)
		return err
	}

	err := store.MoveSongs(d.artist, d.album, r.OldName, newD.artist, newD.album, r.NewName, path, d.mPoint)
	if err != nil {
		return fuse.EIO
	}
	return nil
}
