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
	"github.com/dankomiocevic/mulifs/musicmgr"
	"github.com/dankomiocevic/mulifs/store"
	"github.com/golang/glog"
	"os"
	"path/filepath"
	"strings"
)

// visit checks that the specified file is
// a music file and is on the correct path.
// If it is ok, it stores it on the database.
func visit(path string, f os.FileInfo, err error) error {
	if strings.HasSuffix(path, ".mp3") {
		glog.Infof("Reading %s\n", path)
		err, f := musicmgr.GetMp3Tags(path)
		if err != nil {
			glog.Errorf("Error in %s\n", path)
		}
		if f.Artist == "drop" {
			glog.Errorf("Error in %s\n", path)
		}
		if f.Artist == "playlists" {
			glog.Errorf("Error in %s\n", path)
		}
		store.StoreNewSong(&f, path)
	}
	return nil
}

// ScanFolder scans the specified root path
// and SubDirectories searching for music files.
// It uses filepath to walk through the file tree
// and calls visit on every endpoint found.
func ScanFolder(root string) error {
	err := filepath.Walk(root, visit)
	return err
}
