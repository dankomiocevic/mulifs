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
	"fmt"
	"time"
)

/** FileItem struct contains three elements
 *  the File object that needs to be processed,
 *  the last time it was modified and the last
 *  action performed over it.
 */
type FileItem struct {
	Fn         func(File) error
	FileObject File
	Touched    time.Time
}

var fileItems []FileItem
var fChannel chan FileItem

/** InitDispatcher initializes the
 *  lists and channels to connect to the
 *  dispatcher. It also inits the main loop.
 */
func InitDispatcher() {
	fileItems = make([]FileItem, 0, 20)
	fChannel = make(chan FileItem, 10)

	go processMsgs()
}

/** compareFiles compares two File structs
 *  and returns true if the structs are the equal.
 */
func compareFiles(f1, f2 File) bool {
	return f1.artist == f2.artist && f1.album == f2.album && f1.song == f2.song && f1.name == f2.name
}

/** addFile adds a new FileItem to the list of
 *  items that need to be processed once timed out.
 *  If the element already exists on the list, it gets
 *  updated.
 */
func addFile(f FileItem) {
	for count, item := range fileItems {
		if compareFiles(item.FileObject, f.FileObject) {
			fileItems[count].Touched = f.Touched
			if f.Fn != nil {
				fileItems[count].Fn = f.Fn
			}
			return
		}
	}
	fileItems = append(fileItems, f)
}

/** cleanLists checks that any of the file
 *  elements has been timed out and runs
 *  the correct action over it.
 *  It also deletes the timed out elements
 *  from the list.
 */
func cleanLists() {
	timeout := time.Now().Add(time.Second * -3)
	for i := len(fileItems) - 1; i >= 0; i-- {
		item := fileItems[i]
		if item.Touched.Before(timeout) {
			if item.Fn != nil {
				item.Fn(item.FileObject)
			}
			fileItems = append(fileItems[:i], fileItems[i+1:]...)
		}
	}
}

/** processMsgs receives all the messages
 *  from the channels and process them.
 *  This is the main loop of the dispatcher.
 */
func processMsgs() {
	for {
		select {
		case res := <-fChannel:
			addFile(res)
		case <-time.After(time.Second * 3):
			cleanLists()
		}
	}
}

/** PushFileItem receives a new File
 *  to be processed in the near future.
 *  The fn parameter is the function that
 *  is going to been executed on the file.
 */
func PushFileItem(f File, fn func(File) error) {
	fmt.Printf("Push event for file\n")
	fileItem := FileItem{
		Fn:         fn,
		FileObject: f,
		Touched:    time.Now(),
	}
	fChannel <- fileItem
}
