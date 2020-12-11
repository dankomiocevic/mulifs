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
	"strconv"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type fs_config struct {
	uid         uint
	gid         uint
	allow_users bool
}

var config_params fs_config
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

func newTrue() *bool {
	b := true
	return &b
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(progName + ": ")

	flag.Usage = usage
	var err error
	var db_path string
	var mount_ops string
	flag.StringVar(&db_path, "db_path", "muli.db", "Database path.")
	flag.StringVar(&mount_ops, "o", "", "Default mount options.")
	uid_conf := flag.Uint("uid", 0, "User owner of the files.")
	gid_conf := flag.Uint("gid", 0, "Group owner of the files.")
	allow_other := flag.Bool("allow_other", false, "Allow other users to access the filesystem.")

	flag.Parse()
		
	if len(mount_ops) < 1 && flag.NArg() > 3 {
		for index, marg := range flag.Args() {
			if strings.Compare(marg, "-o") == 0 {
				mount_ops = flag.Arg(index + 1)
				break
			}
		}
	}
	
	
	if len(os.Getenv("PATH")) < 1 {
		os.Setenv("PATH", "/bin:/sbin")
	}

	if len(mount_ops) > 0 {
		opts_tokens := strings.Split(mount_ops, ",")
		for _, token := range opts_tokens {
			if strings.Compare(token, "allow_other") == 0 {
				allow_other = newTrue()
			} else if strings.HasPrefix(token, "uid=") {
				parsed_uid, err := strconv.ParseUint(token[len("uid="):], 10, 32)
				if err != nil {
					log.Fatal(err)
					os.Exit(1)
				} else {
					uint_uid := uint(parsed_uid)
					uid_conf = &uint_uid
				}
			} else if strings.HasPrefix(token, "gid=") {
				parsed_gid, err := strconv.ParseUint(token[len("gid="):], 10, 32)
				if err != nil {
					log.Fatal(err)
					os.Exit(1)
				} else {
					uint_gid := uint(parsed_gid)
					gid_conf = &uint_gid
				}
			} else if strings.HasPrefix(token, "db_path=") {
				db_path = token[len("db_path="):]
				if len(db_path) < 3 {
					log.Fatal("Error in db_path")
					os.Exit(1)
				}
			}
		}
	}

	config_params = fs_config{
		uid: *uid_conf, gid: *gid_conf, allow_users: *allow_other,
	}

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

	err = store.InitDB(db_path)
	if err != nil {
		log.Fatal(err)
		os.Exit(5)
	}

	path, err = filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
		os.Exit(6)
	}

	err = tools.ScanFolder(path)
	if err != nil {
		log.Fatal(err)
		os.Exit(7)
	}

	err = tools.ScanPlaylistFolder(path)
	if err != nil {
		log.Fatal(err)
		os.Exit(8)
	}

	// Init the dispatcher system to process
	// delayed events.
	InitDispatcher()

	if err = mount(path, mountpoint); err != nil {
		log.Fatal(err)
		os.Exit(9)
	}
}

// mount calls the fuse library to specify
// the details of the mounted filesystem.
func mount(path, mountpoint string) error {
	// TODO: Check that there is no folder named

	mountOptions := []fuse.MountOption{
		fuse.FSName("MuLi"),
		fuse.Subtype("MuLiFS"),
		fuse.LocalVolume(),
		fuse.VolumeName("Music Library"),
	}

	if config_params.allow_users {
		mountOptions = append(mountOptions, fuse.AllowOther())
	}
	// playlist or drop in the path.
	c, err := fuse.Mount(
		mountpoint, mountOptions...)

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
