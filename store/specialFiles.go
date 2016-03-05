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

// This file manages all the files related to OSX special files
// that start with a dot.

import (
	"bazil.org/fuse"
	"github.com/boltdb/bolt"
	"github.com/golang/glog"
)

// GetSpecialFile gets a value from the database.
func GetSpecialFile(artist string, album string, name string) (returnValue []byte, err error) {
	glog.Infof("Getting special file: %s Artist: %s Album: %s\n", name, artist, album)
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		buck := tx.Bucket([]byte("Artists"))
		if len(artist) > 1 {
			buck := buck.Bucket([]byte(artist))
			if buck == nil {
				glog.Errorf("Artist not found.")
				return fuse.ENOENT
			}
			if len(album) > 1 {
				buck := buck.Bucket([]byte(album))
				if buck == nil {
					glog.Errorf("Album not found.")
					return fuse.ENOENT
				}
			}
		}
		b := buck.Get([]byte(name))
		if b == nil {
			glog.Info("Special file not found. \n")
			return fuse.ENOENT
		}

		returnValue = append([]byte(nil), b...)
		return nil
	})

	return
}

// PutSpecialFile gets a value from the database.
func PutSpecialFile(artist string, album string, name string, data []byte) error {
	glog.Infof("Updating special file: %s Artist: %s Album: %s\n", name, artist, album)
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return fuse.EIO
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		buck := tx.Bucket([]byte("Artists"))
		if len(artist) > 1 {
			buck := buck.Bucket([]byte(artist))
			if buck == nil {
				glog.Errorf("Artist not found.")
				return fuse.ENOENT
			}
			if len(album) > 1 {
				buck := buck.Bucket([]byte(album))
				if buck == nil {
					glog.Errorf("Album not found.")
					return fuse.ENOENT
				}
			}
		}
		b := buck.Put([]byte(name), data)
		if b != nil {
			glog.Errorf("Cannot write special file. \n")
			return fuse.EIO
		}

		return nil
	})

	return err
}

// MoveSpecialFile moves a value from one key to another in the BoltDB
func MoveSpecialFile(oldArtist, oldAlbum, oldName, newArtist, newAlbum, newName string) error {
	glog.Infof("Moving special file: %s Artist: %s Album: %s to new name: %s, new Artist: %s, new Album: %s\n", oldName, oldArtist, oldAlbum, newName, newArtist, newAlbum)

	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return fuse.EIO
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		oldBuck := tx.Bucket([]byte("Artists"))
		if len(oldArtist) > 1 {
			oldBuck = oldBuck.Bucket([]byte(oldArtist))
			if oldBuck == nil {
				glog.Info("Artist not found.")
				return fuse.ENOENT
			}
			if len(oldAlbum) > 1 {
				oldBuck = oldBuck.Bucket([]byte(oldAlbum))
				if oldBuck == nil {
					glog.Info("Album not found.")
					return fuse.ENOENT
				}
			}
		}

		data := oldBuck.Get([]byte(oldName))

		buck := tx.Bucket([]byte("Artists"))
		if len(newArtist) > 1 {
			buck = buck.Bucket([]byte(newArtist))
			if buck == nil {
				glog.Info("Artist not found.")
				return fuse.ENOENT
			}
			if len(newAlbum) > 1 {
				buck := buck.Bucket([]byte(newAlbum))
				if buck == nil {
					glog.Errorf("Album not found.")
					return fuse.ENOENT
				}
			}
		}
		b := buck.Put([]byte(newName), data)
		if b != nil {
			glog.Errorf("Cannot write special file. \n")
			return fuse.EIO
		}
		oldBuck.Delete([]byte(oldName))
		return nil
	})

	return err
}

// DeleteSpecialFile deletes a key for a special file in the BoltDB
func DeleteSpecialFile(artist string, album string, name string) error {
	glog.Infof("Deleting special file: %s Artist: %s Album: %s\n", name, artist, album)
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return fuse.EIO
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		buck := tx.Bucket([]byte("Artists"))
		if len(artist) > 1 {
			buck := buck.Bucket([]byte(artist))
			if buck == nil {
				glog.Errorf("Artist not found.")
				return fuse.ENOENT
			}
			if len(album) > 1 {
				buck := buck.Bucket([]byte(album))
				if buck == nil {
					glog.Errorf("Album not found.")
					return fuse.ENOENT
				}
			}
		}
		buck.Delete([]byte(name))

		return nil
	})

	return err
}
