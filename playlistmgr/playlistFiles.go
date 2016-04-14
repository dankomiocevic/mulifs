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

// Package playlistmgr contains all the tools to read and modify
// playlists files.
package playlistmgr

import (
	"bufio"
	"errors"
	"github.com/golang/glog"
	"os"
	"strings"
)

// FileTags defines the tags found in a specific music file.
type PlaylistFile struct {
	Title  string
	Artist string
	Album  string
	Path   string
}

// CheckPlaylistFile opens a Playlist file and checks that
// it really is a valid Playlist.
func CheckPlaylistFile(path string) error {
	f, err := os.Open(path)

	if err != nil {
		return err
	}
	defer f.Close()

	firstLine := make([]byte, len("#EXTM3U"))
	_, err = f.Read(firstLine)
	if err != nil {
		return err
	}

	if string(firstLine) != "#EXTM3U" {
		return errors.New("Not a playlist!")
	}

	return nil
}

// ProcessPlaylist function receives the path of a playlist
// and adds all the information into the database.
// It process every line in the file and reads all the
// songs in it.
func ProcessPlaylist(path string) ([]PlaylistFile, error) {
	var a []PlaylistFile
	src, err := os.Stat(path)
	if err != nil || src.IsDir() {
		return a, errors.New("Playlist not found.")
	}

	file, err := os.Open(path)
	if err != nil {
		return a, errors.New("Cannot open playlist file.")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#MULI ") {
			line = line[len("#MULI "):]
			items := strings.Split(line, " - ")
			if len(items) != 3 {
				var playlistFile PlaylistFile
				playlistFile.Artist = items[0]
				playlistFile.Album = items[1]
				playlistFile.Title = items[2]
				a = append(a, playlistFile)
			}
		}
	}

	err = scanner.Err()
	return a, err
}

// DeletePlaylist deletes a playlist from the filesystem.
func DeletePlaylist(playlist, mPoint string) error {
	if mPoint[len(mPoint)-1] != '/' {
		mPoint = mPoint + "/"
	}

	path := mPoint + playlist
	src, err := os.Stat(path)
	if err == nil && src.IsDir() {
		os.Remove(path)
	}

	path = path + ".m3u"
	_, err = os.Stat(path)
	if err == nil {
		os.Remove(path)
	}

	return nil
}

// RegeneratePlaylistFile creates the playlist file from the
// information in the database.
func RegeneratePlaylistFile(songs []PlaylistFile, playlist, mPoint string) error {
	glog.Infof("Regenerating playlist file for playlist: %s\n", playlist)
	if mPoint[len(mPoint)-1] != '/' {
		mPoint = mPoint + "/"
	}

	path := mPoint + playlist + ".m3u"

	_, err := os.Stat(path)
	if err == nil {
		os.Remove(path)
	}

	f, err := os.Create(path)
	if err != nil {
		glog.Errorf("Error creating playlist file: %s\n", path)
		return err
	}
	defer f.Close()

	_, err = f.WriteString("#EXTM3U\n")
	if err != nil {
		glog.Infof("Cannot write on file.")
	}
	for _, s := range songs {
		_, err = f.WriteString("#MULI ")
		if err != nil {
			glog.Infof("Cannot write on file.")
		}
		_, err = f.WriteString(s.Artist)
		if err != nil {
			glog.Infof("Cannot write on file.")
		}
		_, err = f.WriteString(" - ")
		if err != nil {
			glog.Infof("Cannot write on file.")
		}
		_, err = f.WriteString(s.Album)
		if err != nil {
			glog.Infof("Cannot write on file.")
		}
		_, err = f.WriteString(" - ")
		if err != nil {
			glog.Infof("Cannot write on file.")
		}
		_, err = f.WriteString(s.Title)
		if err != nil {
			glog.Infof("Cannot write on file.")
		}
		_, err = f.WriteString("\n")
		if err != nil {
			glog.Infof("Cannot write on file.")
		}
		_, err = f.WriteString(s.Path)
		if err != nil {
			glog.Infof("Cannot write on file.")
		}
		_, err = f.WriteString("\n\n")
		if err != nil {
			glog.Infof("Cannot write on file.")
		}
	}

	f.Sync()
	return nil
}
