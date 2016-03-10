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

// Package store is package for managing the database that stores
// the information about the songs, artists and albums.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dankomiocevic/mulifs/musicmgr"
	"os"
	"path/filepath"

	"bazil.org/fuse"
	"github.com/boltdb/bolt"
	"github.com/golang/glog"
)

// MoveSongs changes the Songs path.
// It modifies the information in the database
// and updates the tags to match the new location.
// It also moves the actual file into the new
// location.
func MoveSongs(oldArtist, oldAlbum, oldName, newArtist, newAlbum, newName, path, mPoint string) error {
	glog.Infof("Moving song from Artist: %s, Album: %s, name: %s and path: %s to Artist: %s, Album: %s, name: %s\n", oldArtist, oldAlbum, oldName, path, newArtist, newAlbum, newName)

	extension := filepath.Ext(path)
	if extension != ".mp3" {
		glog.Info("Wrong file format.")
		return errors.New("Wrong file format.")
	}
	rootPoint := mPoint
	if rootPoint[len(rootPoint)-1] != '/' {
		rootPoint = rootPoint + "/"
	}

	newPath := rootPoint + newArtist + "/" + newAlbum + "/"
	newFullPath := newPath + getCompatibleString(newName[:len(newName)-len(extension)]) + extension

	err := os.Rename(path, newFullPath)
	if err != nil {
		glog.Infof("Cannot rename the file: %s\n", err)
		return err
	}

	err = DeleteSong(oldArtist, oldAlbum, oldName)
	if err != nil {
		glog.Infof("Cannot delete song: %s\n", err)
		return err
	}

	musicmgr.SetMp3Tags(newArtist, newAlbum, newName, newFullPath)
	_, err = CreateSong(newArtist, newAlbum, newName, newPath)
	if err != nil {
		glog.Infof("Cannot create son in the db: %s\n", err)
	}
	return err
}

func processNewAlbum(newArtist, newAlbum, oldArtist, oldAlbum string) ([][]byte, error) {
	var songs [][]byte

	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		glog.Error("Error opening the database.")
		return nil, err
	}
	defer db.Close()

	newAlbumRaw := newAlbum
	newAlbum = getCompatibleString(newAlbum)

	err = db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Artists"))
		artistBucket := root.Bucket([]byte(newArtist))
		if artistBucket == nil {
			glog.Info("Destination Artist not found.")
			return errors.New("Artist not found.")
		}

		// Create the bucket
		albumBucket, err := artistBucket.CreateBucketIfNotExists([]byte(newAlbum))
		if err != nil {
			glog.Info("Cannot create Album bucket.")
			return fuse.EIO
		}

		// Get the description file or create it if it does not exist
		description := albumBucket.Get([]byte(".description"))
		if description == nil {
			albumStore := &AlbumStore{
				AlbumName: newAlbumRaw,
				AlbumPath: newAlbum,
			}

			encoded, err := json.Marshal(albumStore)
			if err != nil {
				glog.Info("Cannot encode description JSON.")
				return fuse.EIO
			}
			albumBucket.Put([]byte(".description"), encoded)
		}

		oldArtistBucket := root.Bucket([]byte(oldArtist))
		if oldArtistBucket == nil {
			glog.Info("Source Artist not found.")
			return errors.New("Artist not found")
		}

		oldAlbumBucket := oldArtistBucket.Bucket([]byte(oldAlbum))
		if oldAlbumBucket == nil {
			glog.Info("Source Album not found.")
			return errors.New("Album not found")
		}

		// Get all the songs and store it in a temporary slice
		c := oldAlbumBucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var temp []byte
			if k[0] == '.' {
				continue
			}
			temp = make([]byte, len(v))
			copy(temp, v)
			songs = append(songs, temp)
		}
		return nil
	})
	return songs, err
}

// MoveAlbum changes the album path.
// It modifies the information in the database
// and updates the tags to match the new location
// on every song inside the album.
// It also moves the actual files into the new location.
func MoveAlbum(oldArtist, oldAlbum, newArtist, newAlbum, mPoint string) error {
	glog.Infof("Moving Album from Artist: %s, Album: %s to Artist: %s, Album: %s\n", oldArtist, oldAlbum, newArtist, newAlbum)

	// Check that the file is being moved in the same level
	// Album -> Album
	if len(oldArtist) < 1 || len(newArtist) < 1 {
		glog.Info("Cannot change Album to Artist.")
		return fuse.EPERM
	}

	if len(oldAlbum) < 1 || len(newAlbum) < 1 {
		glog.Info("Cannot change Album to Artist.")
		return fuse.EPERM
	}

	rootPoint := mPoint
	if rootPoint[len(rootPoint)-1] != '/' {
		rootPoint = rootPoint + "/"
	}
	newPath := rootPoint + newArtist + "/" + newAlbum + "/"

	// Create the directory
	err := os.Mkdir(newPath, 0777)
	if err != nil {
		glog.Info("Cannot create the new directory.")
		return fuse.EIO
	}

	var songs [][]byte
	songs, err = processNewAlbum(newArtist, newAlbum, oldArtist, oldAlbum)
	if err != nil {
		return err
	}

	glog.Infof("Moving %d songs.\n", len(songs))
	// Move all the songs inside the Album
	for _, element := range songs {
		var song SongStore
		fmt.Printf("Element: %s.\n", element)
		err := json.Unmarshal(element, &song)
		if err != nil {
			glog.Info("Cannot unmarshall JSON")
			continue
		}
		MoveSongs(oldArtist, oldAlbum, song.SongPath, newArtist, newAlbum, song.SongPath, song.SongFullPath, mPoint)
	}

	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		glog.Error("Error opening the database.")
		return err
	}
	defer db.Close()

	// Finally delete the old Artist bucket
	err = db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Artists"))
		artistBucket := root.Bucket([]byte(oldArtist))
		if artistBucket == nil {
			return errors.New("Artist not found.")
		}
		err = artistBucket.DeleteBucket([]byte(oldAlbum))
		return err
	})

	if err != nil {
		return fuse.EIO
	}
	return nil
}
