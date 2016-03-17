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
	"flag"
	"fmt"
	"github.com/dankomiocevic/mulifs/store"
	"github.com/dankomiocevic/mulifs/tools"
	"log"
	"os"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var progName = filepath.Base(os.Args[0])

const progVer = "0.1"

// usage specifies how the command should be called
// showing a message on the standard output.
func usage() {
	fmt.Fprintf(os.Stderr, "Name:\n")
	fmt.Fprintf(os.Stderr, "  %s %s\n", progName, progVer)
	fmt.Fprintf(os.Stderr, "\nSynopsis:\n")
	fmt.Fprintf(os.Stderr, "  %s [global_options] MUSIC_SOURCE MOUNTPOINT \n", progName)
	fmt.Fprintf(os.Stderr, "\nDescription:\n")
	fmt.Fprintf(os.Stderr, "  Mounts a filesystem in MOUNTPOINT with the music files obtained\n")
	fmt.Fprintf(os.Stderr, "  from MUSIC_SOURCE ordered in folders by Artist and Album.\n")
	fmt.Fprintf(os.Stderr, "\n  For more information please visit:\n")
	fmt.Fprintf(os.Stderr, "    <http://github.com/dankomiocevic/mulifs>\n")
	fmt.Fprintf(os.Stderr, "\nParams:\n")
	fmt.Fprintf(os.Stderr, "  MUSIC_SOURCE: The path of the folder containing the music files.\n")
	fmt.Fprintf(os.Stderr, "  MOUNTPOINT: The path where MuLi should be mounted.\n")
	fmt.Fprintf(os.Stderr, "\nGlobal Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n")
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(progName + ": ")

	flag.Usage = usage
	var db_path string
	flag.StringVar(&db_path, "db_path", "muli.db", "Database path.")
	flag.Parse()

	if flag.NArg() < 2 {
		usage()
		os.Exit(2)
	}
	path := flag.Arg(0)
	mountpoint := flag.Arg(1)

	if path[0] == '-' {
		usage()
		os.Exit(3)
	}

	if mountpoint[0] == '-' {
		usage()
		os.Exit(4)
	}

	err := store.InitDB(db_path)
	if err != nil {
		log.Fatal(err)
		os.Exit(5)
	}

	err = tools.ScanFolder(path)
	if err != nil {
		log.Fatal(err)
		os.Exit(6)
	}

	// Init the dispatcher system to process
	// delayed events.
	InitDispatcher()

	if err = mount(path, mountpoint); err != nil {
		log.Fatal(err)
		os.Exit(7)
	}
}

// mount calls the fuse library to specify
// the details of the mounted filesystem.
func mount(path, mountpoint string) error {
	// TODO: Check that there is no folder named
	// playlist or drop in the path.
	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("MuLi"),
		fuse.Subtype("MuLiFS"),
		fuse.LocalVolume(),
		fuse.VolumeName("Music Library"),
	)

	if err != nil {
		return err
	}
	defer c.Close()

	filesys := &FS{
		mPoint: path,
	}

	if err := fs.Serve(c, filesys); err != nil {
		return err
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		return err
	}

	return nil
}
