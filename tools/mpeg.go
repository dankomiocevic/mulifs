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

// Package tools contains different kind of tools to
// manage the files in the filesystem.
// The tools include tools to read the different Tags in the
// music files and to scan the Directories and SubDirectories in
// the target path.
package tools

import (
	"github.com/dankomiocevic/mulifs/store"
	"path/filepath"

	id3 "github.com/mikkyang/id3-go"
)

// GetMp3Tags returns a store.FileTags struct with
// all the information obtained from the tags in the
// MP3 file.
// Includes the Artist, Album and Song and defines
// default values if the values are missing.
// If the tags are missing, the default values will
// be stored on the file.
// If the tags are obtained correctly the first
// return value will be nil.
func GetMp3Tags(path string) (error, store.FileTags) {
	mp3File, err := id3.Open(path)
	if err != nil {
		_, file := filepath.Split(path)
		extension := filepath.Ext(file)
		songTitle := file[0 : len(file)-len(extension)]
		return err, store.FileTags{songTitle, "unknown", "unknown"}
	}

	defer mp3File.Close()

	title := mp3File.Title()
	if title == "" || title == "unknown" {
		_, file := filepath.Split(path)
		extension := filepath.Ext(file)
		title = file[0 : len(file)-len(extension)]
		mp3File.SetTitle(title)
	}

	artist := mp3File.Artist()
	if artist == "" {
		artist = "unknown"
		mp3File.SetArtist(artist)
	}

	album := mp3File.Album()
	if album == "" {
		album = "unknown"
		mp3File.SetAlbum(album)
	}

	ft := store.FileTags{title, artist, album}
	return nil, ft
}

// SetMp3Tags updates the Artist, Album and Title
// tags with new values in the song MP3 file.
func SetMp3Tags(artist string, album string, title string, songPath string) error {
	mp3File, err := id3.Open(songPath)
	if err != nil {
		return err
	}
	defer mp3File.Close()

	mp3File.SetTitle(title)
	mp3File.SetArtist(artist)
	mp3File.SetAlbum(album)

	return nil
}
