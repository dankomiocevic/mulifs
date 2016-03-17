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
	"errors"
	"github.com/dankomiocevic/mulifs/musicmgr"
	"github.com/golang/glog"
	"os"
	"path/filepath"

	"bazil.org/fuse"
)

/** Deletes a file in the drop folder.
 */
func deleteDrop(path string) {
	os.Remove(path)
}

/** This function manages the Drop directory.
 *  The user can copy/create files into this directory and
 *  the files will be organized to the correct directory
 *  based on the file tags.
 */
func HandleDrop(path, rootPoint string) error {
	glog.Infof("Handle drop with path: %s\n", path)
	err, fileTags := musicmgr.GetMp3Tags(path)
	if err != nil {
		deleteDrop(path)
		return fuse.EIO
	}

	extension := filepath.Ext(path)

	artist, err := CreateArtist(fileTags.Artist)
	if err != nil && err != fuse.EEXIST {
		glog.Infof("Error creating Artist: %s\n", err)
		return err
	}

	album, err := CreateAlbum(artist, fileTags.Album)
	if err != nil && err != fuse.EEXIST {
		glog.Infof("Error creating Album: %s\n", err)
		return err
	}

	//_, file := filepath.Split(path)
	newPath := rootPoint + artist + "/" + album + "/"
	os.MkdirAll(newPath, 0777)

	file := getCompatibleString(fileTags.Title) + extension
	err = os.Rename(path, newPath+file)
	if err != nil {
		glog.Infof("Error renaming song: %s\n", err)
		return fuse.EIO
	}

	_, err = CreateSong(artist, album, fileTags.Title+extension, newPath)
	deleteDrop(path)
	if err != nil {
		glog.Infof("Error creating song in the DB: %s\n", err)
	}
	return err
}

/** Returns the path of a file in the drop directory.
 */
func GetDropFilePath(name, mPoint string) (string, error) {
	rootPoint := mPoint
	if rootPoint[len(rootPoint)-1] != '/' {
		rootPoint = rootPoint + "/"
	}

	path := mPoint + "drop/" + name
	glog.Infof("Getting drop file path for: %s\n", path)
	// Check if the file exists
	src, err := os.Stat(path)
	if err == nil && src.IsDir() {
		glog.Info("File not found.")
		return "", errors.New("File not found.")
	}
	return path, err
}
