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
	"regexp"
	"strings"
	"unicode"

	"bazil.org/fuse"
	"github.com/boltdb/bolt"
	"github.com/golang/glog"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// config stores the general configuration for the store.
// DbPath is the path to the database file.
var config struct {
	DbPath string
}

// ArtistStore is the information for a specific artist
// to be stored in the database.
type ArtistStore struct {
	ArtistName   string
	ArtistPath   string
	ArtistAlbums []string
}

// AlbumStore is the information for a specific album
// to be stored in the database.
type AlbumStore struct {
	AlbumName string
	AlbumPath string
}

// SongStore is the information for a specific song
// to be stored in the database.
type SongStore struct {
	SongName     string
	SongPath     string
	SongFullPath string
}

// InitDB initializes the database with the
// specified configuration and returns nil if
// there was no problem.
func InitDB(path string) error {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("Artists"))
		if err != nil {
			glog.Errorf("Error creating bucket: %s", err)
			return fmt.Errorf("Error creating bucket: %s", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	config.DbPath = path
	return nil
}

// isMn checks if the rune is in the Unicode
// category Mn.
func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}

// getCompatibleString removes all the special characters
// from the string name to create a new string compatible
// with different file names.
func getCompatibleString(name string) string {
	// Replace all the & signs with and text
	name = strings.Replace(name, "&", "and", -1)
	// Change all the characters to ASCII
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, _ := transform.String(t, name)
	// Replace all the spaces with underscore
	s, _ := regexp.Compile(`\s+`)
	result = s.ReplaceAllString(result, "_")
	// Remove all the non alphanumeric characters
	r, _ := regexp.Compile(`\W`)
	result = r.ReplaceAllString(result, "")
	return result
}

// StoreNewSong takes the information received from
// the song file tags and creates the item in the
// database accordingly. It checks the different fields
// and completes the missing information with the default
// data.
func StoreNewSong(song *musicmgr.FileTags, path string) error {
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	var artistStore ArtistStore
	var albumStore AlbumStore
	var songStore SongStore

	err = db.Update(func(tx *bolt.Tx) error {
		// Get the artists bucket
		artistsBucket, updateError := tx.CreateBucketIfNotExists([]byte("Artists"))
		if updateError != nil {
			glog.Errorf("Error creating bucket: %s", updateError)
			return fmt.Errorf("Error creating bucket: %s", updateError)
		}

		// Generate the compatible names for the fields
		artistPath := getCompatibleString(song.Artist)
		albumPath := getCompatibleString(song.Album)
		songPath := getCompatibleString(song.Title)

		// Generate artist bucket
		artistBucket, updateError := artistsBucket.CreateBucketIfNotExists([]byte(artistPath))
		if updateError != nil {
			glog.Errorf("Error creating bucket: %s", updateError)
			return fmt.Errorf("Error creating bucket: %s", updateError)
		}

		// Update the description of the Artist
		descValue := artistBucket.Get([]byte(".description"))
		if descValue == nil {
			artistStore.ArtistName = song.Artist
			artistStore.ArtistPath = artistPath
			artistStore.ArtistAlbums = []string{albumPath}
		} else {
			err := json.Unmarshal(descValue, &artistStore)
			if err != nil {
				artistStore.ArtistName = song.Artist
				artistStore.ArtistPath = artistPath
				artistStore.ArtistAlbums = []string{albumPath}
			}

			var found bool = false
			for _, a := range artistStore.ArtistAlbums {
				if a == albumPath {
					found = true
					break
				}
			}

			if found == false {
				artistStore.ArtistAlbums = append(artistStore.ArtistAlbums, albumPath)
			}
		}
		encoded, err := json.Marshal(artistStore)
		if err != nil {
			return err
		}
		artistBucket.Put([]byte(".description"), encoded)

		// Get the album bucket
		albumBucket, updateError := artistBucket.CreateBucketIfNotExists([]byte(albumPath))
		if updateError != nil {
			glog.Errorf("Error creating bucket: %s", updateError)
			return fmt.Errorf("Error creating bucket: %s", updateError)
		}

		// Update the album description
		albumStore.AlbumName = song.Album
		albumStore.AlbumPath = albumPath
		encoded, err = json.Marshal(albumStore)
		if err != nil {
			return err
		}
		albumBucket.Put([]byte(".description"), encoded)

		_, file := filepath.Split(path)
		extension := filepath.Ext(file)

		// Add the song to the album bucket
		songStore.SongName = song.Title
		songStore.SongPath = songPath + extension
		songStore.SongFullPath = path

		encoded, err = json.Marshal(songStore)
		if err != nil {
			return err
		}

		albumBucket.Put([]byte(songPath+extension), encoded)
		return nil
	})

	return nil
}

// ListArtists returns all the Dirent corresponding
// to Artists in the database.
// This is used to generate the Artist listing on the
// generated filesystem.
// It returns nil in the second return value if there
// was no error and nil if the Artists were
// obtained correctly.
func ListArtists() ([]fuse.Dirent, error) {
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var a []fuse.Dirent
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Artists"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				var node fuse.Dirent
				node.Name = string(k)
				node.Type = fuse.DT_Dir
				a = append(a, node)
			} else {
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

	return a, nil
}

// ListAlbums returns all the Dirent corresponding
// to Albums for a specified Artist in the database.
// This is used to generate the Album listing on the
// generated filesystem.
// It returns nil in the second return value if there
// was no error and nil if the Albums were
// obtained correctly.
func ListAlbums(artist string) ([]fuse.Dirent, error) {
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var a []fuse.Dirent
	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Artists"))
		b := root.Bucket([]byte(artist))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				var node fuse.Dirent
				node.Name = string(k)
				node.Type = fuse.DT_Dir
				a = append(a, node)
			} else {
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

	return a, nil
}

// ListSongs returns all the Dirent corresponding
// to Songs for a specified Artist and Album
// in the database.
// This is used to generate the Song listing on the
// generated filesystem.
// It returns nil in the second return value if there
// was no error and nil if the Songs were
// obtained correctly.
func ListSongs(artist string, album string) ([]fuse.Dirent, error) {
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var a []fuse.Dirent
	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Artists"))
		artistBucket := root.Bucket([]byte(artist))
		b := artistBucket.Bucket([]byte(album))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var song SongStore
			if k[0] != '.' || string(k) == ".description" {
				err := json.Unmarshal(v, &song)
				if err != nil {
					continue
				}
			}
			var node fuse.Dirent
			node.Name = string(k)
			node.Type = fuse.DT_File
			a = append(a, node)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return a, nil
}

// GetArtistPath checks that a specified Artist
// exists on the database and returns a fuse
// error if it does not.
// It also returns the Artist name as string.
func GetArtistPath(artist string) (string, error) {
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return "", err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Artists"))
		artistBucket := root.Bucket([]byte(artist))

		if artistBucket == nil {
			return fuse.ENOENT
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	return artist, nil
}

// GetAlbumPath checks that a specified Artist
// and Album exists on the database and returns
// a fuse error if it does not.
// It also returns the Album name as string.
func GetAlbumPath(artist string, album string) (string, error) {
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return "", err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Artists"))
		artistBucket := root.Bucket([]byte(artist))

		if artistBucket == nil {
			return fuse.ENOENT
		}

		albumBucket := artistBucket.Bucket([]byte(album))

		if albumBucket == nil {
			return fuse.ENOENT
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return artist, nil
}

// GetFilePath checks that a specified Song
// Album exists on the database and returns
// the full path to the Song file.
// If there is an error obtaining the Song
// an error will be returned.
func GetFilePath(artist string, album string, song string) (string, error) {
	glog.Infof("Getting file path for song: %s Artist: %s Album: %s\n", song, artist, album)
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var returnValue string

	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Artists"))
		artistBucket := root.Bucket([]byte(artist))
		b := artistBucket.Bucket([]byte(album))
		songJson := b.Get([]byte(song))
		if songJson == nil {
			glog.Info("Song not found.")
			return fuse.ENOENT
		}

		var songStore SongStore
		err := json.Unmarshal(songJson, &songStore)
		if err != nil {
			glog.Error("Cannot open song.")
			return errors.New("Cannot open song.")
		}
		returnValue = songStore.SongFullPath
		return nil
	})

	if err != nil {
		return "", err
	}
	return returnValue, nil
}

// GetDescription obtains a Song description from the
// database as a JSON object.
// If the description is obtained correctly a string with
// the JSON is returned and nil.
func GetDescription(artist string, album string, name string) (string, error) {
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var returnValue string

	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Artists"))
		artistBucket := root.Bucket([]byte(artist))
		var descJson []byte
		if len(album) < 1 {
			descJson = artistBucket.Get([]byte(name))
		} else {
			b := artistBucket.Bucket([]byte(album))
			descJson = b.Get([]byte(name))
		}

		if descJson == nil {
			return fuse.ENOENT
		}

		returnValue = string(descJson) + "\n"
		return nil
	})

	if err != nil {
		return "", err
	}
	return returnValue, nil
}

// CreateArtist creates a new artist from a Raw
// name. It generates the compatible string to
// use as Directory name and stores the information
// in the database. It also generates the description
// file.
// If there is an error it will be specified in the
// error return value, nil otherwise.
func CreateArtist(nameRaw string) (string, error) {
	name := getCompatibleString(nameRaw)
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return name, err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Artists"))
		artistBucket, createError := root.CreateBucket([]byte(name))
		if createError != nil {
			if createError == bolt.ErrBucketExists {
				return fuse.EEXIST
			} else {
				return fuse.EIO
			}
		}

		var artistStore ArtistStore
		artistStore.ArtistName = nameRaw
		artistStore.ArtistPath = name
		artistStore.ArtistAlbums = []string{}

		encoded, err := json.Marshal(artistStore)
		if err != nil {
			return err
		}
		artistBucket.Put([]byte(".description"), encoded)
		return nil
	})

	return name, err
}

// CreateAlbum creates an album for a specific Artist
// from a Raw name. The name is returned as a compatible
// string to use as a Directory name and the description
// file is created.
// The description file for the Artist will be also updated
// in the process.
// If the Album is created correctly the string return value
// will contain the compatible string to use as Directory
// name and the second value will contain nil.
func CreateAlbum(artist string, nameRaw string) (string, error) {
	name := getCompatibleString(nameRaw)
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return name, err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Artists"))
		artistBucket := root.Bucket([]byte(artist))
		if artistBucket == nil {
			return fuse.ENOENT
		}

		var artistStore ArtistStore
		// Create the album bucket
		albumBucket, createError := artistBucket.CreateBucket([]byte(name))
		if createError != nil {
			if createError == bolt.ErrBucketExists {
				return fuse.EEXIST
			} else {
				return fuse.EIO
			}
		}

		// Update the description of the Artist
		descValue := artistBucket.Get([]byte(".description"))
		if descValue == nil {
			artistStore.ArtistName = artist
			artistStore.ArtistPath = artist
			artistStore.ArtistAlbums = []string{name}
		} else {
			err := json.Unmarshal(descValue, &artistStore)
			if err != nil {
				artistStore.ArtistName = artist
				artistStore.ArtistPath = artist
				artistStore.ArtistAlbums = []string{name}
			}

			var found bool = false
			for _, a := range artistStore.ArtistAlbums {
				if a == name {
					found = true
					break
				}
			}

			if found == false {
				artistStore.ArtistAlbums = append(artistStore.ArtistAlbums, name)
			}
		}
		encoded, err := json.Marshal(artistStore)
		if err != nil {
			return err
		}
		artistBucket.Put([]byte(".description"), encoded)

		// Update the album description
		var albumStore AlbumStore
		albumStore.AlbumName = nameRaw
		albumStore.AlbumPath = name
		encoded, err = json.Marshal(albumStore)
		if err != nil {
			return err
		}
		albumBucket.Put([]byte(".description"), encoded)
		return nil
	})

	return name, err
}

// CreateSong creates a song for a specific Artist and Album
// from a Raw name and a path. The name is returned as a compatible
// string to use as a Directory name and the description
// file is created.
// The path parameter is used to identify the file being
// added to the filesystem.
// If the Song is created correctly the string return value
// will contain the compatible string to use as File
// name and the second value will contain nil.
func CreateSong(artist string, album string, nameRaw string, path string) (string, error) {
	glog.Infof("Adding song to the DB: %s with Artist: %s and Album: %s\n", nameRaw, artist, album)
	extension := filepath.Ext(nameRaw)
	if extension != ".mp3" {
		return "", errors.New("Wrong file format.")
	}

	nameRaw = nameRaw[:len(nameRaw)-len(extension)]
	name := getCompatibleString(nameRaw)

	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return name, err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Artists"))
		artistBucket := root.Bucket([]byte(artist))
		if artistBucket == nil {
			return errors.New("Artist not found.")
		}
		albumBucket := artistBucket.Bucket([]byte(album))
		if albumBucket == nil {
			return errors.New("Album not found.")
		}

		var songStore SongStore
		songStore.SongName = nameRaw
		songStore.SongPath = name + extension
		songStore.SongFullPath = path + name + extension

		encoded, err := json.Marshal(songStore)
		if err != nil {
			return err
		}

		albumBucket.Put([]byte(name+extension), encoded)
		glog.Infof("Created with name: %s\n", name+extension)
		return nil
	})

	return name + extension, err
}

// DeleteArtist deletes the specified Artist only
// in the database and returns nil if there was no error.
func DeleteArtist(artist string) error {
	glog.Infof("Deleting Artist: %s\n", artist)
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Artists"))
		buck := root.Bucket([]byte(artist))

		c := buck.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if k[0] == '.' {
				continue
			}
			album := buck.Bucket([]byte(artist))
			d := album.Cursor()
			for i, _ := d.First(); i != nil; i, _ = d.Next() {
				if i[0] == '.' {
					continue
				}
				songJson := album.Get([]byte(i))
				if songJson == nil {
					continue
				}

				var song SongStore
				err := json.Unmarshal(songJson, &song)
				if err != nil {
					continue
				}
				os.Remove(song.SongFullPath)
			}
		}
		root.DeleteBucket([]byte(artist))
		return nil
	})
	return err
}

// DeleteAlbum deletes the specified Album for
// the specified Artist only in the database and
// returns nil if there was no error.
func DeleteAlbum(artistName string, albumName string) error {
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Artists"))
		artistBucket := root.Bucket([]byte(artistName))
		if artistBucket == nil {
			return errors.New("Artist not found.")
		}

		album := artistBucket.Bucket([]byte(artistName))
		if album == nil {
			return nil
		}
		d := album.Cursor()
		for i, _ := d.First(); i != nil; i, _ = d.Next() {
			if i[0] == '.' {
				continue
			}
			songJson := album.Get([]byte(i))
			if songJson == nil {
				continue
			}

			var song SongStore
			err := json.Unmarshal(songJson, &song)
			if err != nil {
				continue
			}
			os.Remove(song.SongFullPath)
		}
		artistBucket.DeleteBucket([]byte(albumName))
		return nil
	})
	return err
}

// DeleteSong deletes the specified Song in the
// specified Album and Artist only in the database
// and returns nil if there was no error.
func DeleteSong(artist string, album string, song string) error {
	glog.Infof("Deleting song: %s with Artist: %s and Album: %s\n", song, artist, album)
	if song[0] == '.' {
		return nil
	}

	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		return fuse.EIO
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("Artists"))
		artistBucket := root.Bucket([]byte(artist))
		if artistBucket == nil {
			return errors.New("Artist not found.")
		}

		albumBucket := artistBucket.Bucket([]byte(album))
		if albumBucket == nil {
			return errors.New("Album not found.")
		}

		songJson := albumBucket.Get([]byte(song))
		if songJson != nil {
			var songData SongStore
			err := json.Unmarshal(songJson, &songData)
			if err != nil {
				os.Remove(songData.SongFullPath)
			}
		}
		albumBucket.Delete([]byte(song))
		return nil
	})
	return err
}
