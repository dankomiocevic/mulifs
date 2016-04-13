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
// The tools include tools to scan the Directories and SubDirectories in
// the target path.
package tools

import (
	"github.com/dankomiocevic/mulifs/playlistmgr"
	"github.com/dankomiocevic/mulifs/store"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"strings"
)

// visitPlaylist checks that the specified file is
// a music file and is on the correct path.
// If it is ok, it stores it on the database.
func visitPlaylist(name, path, mPoint string) error {
	if path[len(path)-1] != '/' {
		path = path + "/"
	}

	fullPath := path + name
	if strings.HasSuffix(fullPath, ".m3u") {
		glog.Infof("Reading %s\n", fullPath)
		err := playlistmgr.CheckPlaylistFile(fullPath)
		if err != nil {
			glog.Infof("Error in %s playlist\n", name)
			return err
		}

		files, err := playlistmgr.ProcessPlaylist(fullPath)
		if err != nil {
			glog.Infof("Problem reading playlist %s: %s\n", name, err)
			return err
		}

		playlistName := name[:len(name)-len(".m3u")]
		playlistName, _ = store.CreatePlaylist(playlistName, mPoint)

		for _, f := range files {
			store.AddFileToPlaylist(f, playlistName)
		}

		os.Remove(fullPath)
		store.RegeneratePlaylistFile(playlistName, mPoint)
	}
	return nil
}

// ScanPlaylistFolder scans the specified root path
// and SubDirectories searching for playlist files.
// It uses filepath to walk through the file tree
// and calls visit on every endpoint found.
func ScanPlaylistFolder(root string) error {
	if root[len(root)-1] != '/' {
		root = root + "/"
	}

	fullPath := root + "playlists/"
	files, _ := ioutil.ReadDir(fullPath)
	for _, f := range files {
		if !f.IsDir() {
			visitPlaylist(f.Name(), fullPath, root)
		}
	}
	return nil
}
