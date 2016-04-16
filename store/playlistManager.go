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

package store

import (
	"bazil.org/fuse"
	"encoding/json"
	"errors"
	"github.com/boltdb/bolt"
	"github.com/dankomiocevic/mulifs/playlistmgr"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
)

// GetPlaylistPath checks that a specified playlist
// exists on the database and returns an
// error if it does not.
// It also returns the playlist name as string.
func GetPlaylistPath(playlist string) (string, error) {
	glog.Infof("Entered Playlist path with playlist: %s\n", playlist)
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return "", err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Playlists"))
		if root == nil {
			return errors.New("No playlists.")
		}

		playlistBucket := root.Bucket([]byte(playlist))
		if playlistBucket == nil {
			return errors.New("Playlist not exists.")
		}

		return nil
	})

	return playlist, err
}

// GetPlaylistFilePath function should return the path for a specific
// file in a specific playlist.
// The file could be on two places, first option is that the file is
// stored in the database. In that case, the file will be stored somewhere
// else in the MuLi filesystem but that will be specified on the
// item in the database.
// On the other hand, the file could be just dropped inside the playlist
// and it will be temporary stored in a directory inside the playlists
// directory.
// The playlist name is specified on the first argument and the song
// name on the second.
// The mount path is also needed and should be specified on the third
// argument.
// This function returns a string containing the file path and an error
// that will be nil if everything is ok.
func GetPlaylistFilePath(playlist, song, mPoint string) (string, error) {
	glog.Infof("Entered Playlist file path with song: %s, and playlist: %s\n", song, playlist)
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var returnValue string
	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Playlists"))
		if root == nil {
			return errors.New("No playlists.")
		}
		playlistBucket := root.Bucket([]byte(playlist))
		if playlistBucket == nil {
			return errors.New("Playlist not exists.")
		}

		songJson := playlistBucket.Get([]byte(song))
		if songJson == nil {
			return errors.New("Song not found.")
		}

		var file playlistmgr.PlaylistFile
		err := json.Unmarshal(songJson, &file)
		if err != nil {
			return errors.New("Cannot open song.")
		}
		returnValue = file.Path
		return nil
	})

	if err == nil {
		return returnValue, nil
	}

	if mPoint[len(mPoint)-1] != '/' {
		mPoint = mPoint + "/"
	}

	fullPath := mPoint + "playlists/" + playlist + "/" + song
	// Check if the file exists
	src, err := os.Stat(fullPath)
	if err != nil || src.IsDir() {
		return "", errors.New("File not exists.")
	}

	return fullPath, nil
}

// ListPlaylists function returns all the names of the playlists available
// in the MuLi system.
// It receives no arguments and returns a slice of Dir objects to list
// all the available playlists and the error if there is any.
func ListPlaylists() ([]fuse.Dirent, error) {
	glog.Info("Entered list playlists.")
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var a []fuse.Dirent
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Playlists"))
		if b == nil {
			glog.Infof("There is no Playlists bucket.")
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				var node fuse.Dirent
				node.Name = string(k)
				node.Type = fuse.DT_Dir
				a = append(a, node)
			}
		}
		return nil
	})
	return a, nil
}

// ListPlaylistSongs function returns all the songs inside a playlist.
// The available songs are loaded from the database and also from the
// temporary drop directory named after the playlist.
// It receives a playlist name and returns a slice with all the
// files.
func ListPlaylistSongs(playlist, mPoint string) ([]fuse.Dirent, error) {
	glog.Infof("Listing contents of playlist %s.\n", playlist)
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var a []fuse.Dirent
	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Playlists"))
		if root == nil {
			return nil
		}

		b := root.Bucket([]byte(playlist))
		if b == nil {
			return nil
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v != nil {
				var node fuse.Dirent
				node.Name = string(k)
				node.Type = fuse.DT_File
				a = append(a, node)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if mPoint[len(mPoint)-1] != '/' {
		mPoint = mPoint + "/"
	}

	fullPath := mPoint + "playlists/" + playlist + "/"

	files, _ := ioutil.ReadDir(fullPath)
	for _, f := range files {
		if !f.IsDir() {
			var node fuse.Dirent
			node.Name = string(f.Name())
			node.Type = fuse.DT_File
			a = append(a, node)
		}
	}
	return a, nil

	return nil, nil
}

// CreatePlaylist function creates a playlist item in the database and
// also creates it in the filesystem.
// It receives the playlist name and returns the modified name and an
// error if something went wrong.
func CreatePlaylist(name, mPoint string) (string, error) {
	glog.Infof("Creating Playlist with name: %s\n", name)
	name = GetCompatibleString(name)
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return "", err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("Playlists"))
		if err != nil {
			glog.Errorf("Error creating Playlists bucket: %s\n", err)
			return err
		}

		_, err = root.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			glog.Errorf("Error creating %s bucket: %s\n", name, err)
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return name, err
}

// RegeneratePlaylistFile creates the playlist file from the
// information in the database.
func RegeneratePlaylistFile(name, mPoint string) error {
	glog.Infof("Regenerating playlist for name: %s\n", name)
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	var a []playlistmgr.PlaylistFile
	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Playlists"))
		if root == nil {
			glog.Info("Cannot open Playlists bucket.")
			return errors.New("Cannot open Playlists bucket.")
		}

		b := root.Bucket([]byte(name))
		if b == nil {
			glog.Infof("Playlist %s not exists", name)
			return errors.New("Playlist not exists.")
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v != nil {
				var file playlistmgr.PlaylistFile
				err := json.Unmarshal(v, &file)
				if err == nil {
					a = append(a, file)
				} else {
					glog.Errorf("Cannot unmarshal Playlist File %s: %s\n", k, err)
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return playlistmgr.RegeneratePlaylistFile(a, name, mPoint)
}

// AddFileToPlaylist function adds a file to a specific playlist.
// The function also checks that the file exist in the MuLi database.
func AddFileToPlaylist(file playlistmgr.PlaylistFile, playlistName string) error {
	path, err := GetFilePath(file.Artist, file.Album, file.Title)
	if err != nil {
		return errors.New("Playlist item not found in MuLi.")
	}

	file.Path = path
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Playlists"))
		if root == nil {
			glog.Errorf("Error opening Playlists bucket: %s\n", err)
			return err
		}

		playlistBucket := root.Bucket([]byte(playlistName))
		if playlistBucket == nil {
			glog.Errorf("Error opening %s playlist bucket: %s\n", playlistName, err)
			return err
		}

		encoded, err := json.Marshal(file)
		if err != nil {
			glog.Errorf("Cannot encode PlaylistFile.")
			return err
		}
		playlistBucket.Put([]byte(file.Title), encoded)
		return nil
	})

	return err
}

// DeletePlaylist function deletes a playlist from the database
// and also deletes all the entries in the specific files and
// deletes it from the filesystem.
func DeletePlaylist(name, mPoint string) error {
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Playlists"))
		if root == nil {
			glog.Errorf("Error opening Playlists bucket: %s\n", err)
			return err
		}
		//TODO: Scan all the files in the playlist and remove the
		// connections with songs in MuLi.

		err := root.DeleteBucket([]byte(name))
		return err
	})

	return playlistmgr.DeletePlaylist(name, mPoint)
}
