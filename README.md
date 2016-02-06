MuLiFS : Music Library Filesystem
=================================

[![GoDoc](https://godoc.org/github.com/dankomiocevic/mulifs?status.svg)](https://godoc.org/github.com/dankomiocevic/mulifs)

MuLi (pronounced Moo-Lee) is a filesystem written in Go to mount music
libraries and organize the music based on the music file tags.

It scans a Directory tree and reads all the Tags in the music files and
generates a Directory structure organizing the songs by Artist and
Album.

Quick Start
-----------

For the anxious that don't like to read, here is the command to make 
this work:

```
mulifs MUSIC_SOURCE MOUNTPOINT 
```

Where the MUSIC_SOURCE is the path where the music is stored and
MOUNTPOINT is the path where MuLi sould be mounted.


Project status
--------------

This project is currently under development and it is not ready to use yet.
The basic functionality is ready but some work needs to be done, including:

* Finish the playlists management.
* Finish the drop Directory.
* Add testing code.
* Test and test!


How it works
------------

Organizing a Music library is always a tedious task and there is always
lots of different information that does not match.

MuLi reads a Directory tree (Directories and Subdirectories of a specific
path) and scans for all the music files (it actually supports only MP3).
Every time it finds a music file it reads the ID Tags that specify the 
Artist, Album and Song name.
If any of these parameters are missing it completes the information with
default values (unknown Artist or Album and tries to read the song name 
from the path) and updates the Tags for future scans.
It stores all the gathered information into a BoltDB that is an object 
store that is fast, simple and completely written in Go, that makes 
MuLi portable!

Once the Directory is completely scanned and all the information is
in the Database, MuLi creates a directory structure as following:

```
mounted_path
│
├── Some_Artist
│    │
│    ├── Some_Album
│    │     └── Some_song.mp3
│    │ 
│    └── Other_Album
│          ├── More_songs.mp3
│          ├── ...
│          └── Other_song.mp3
│
├── Other_Artist
│    │
│    └── Some_Album
│          ├── Great_Song.mp3
│          ├── ...
│          └── AwesomeSong.mp3
│
├── drop
│ 
└── playlists
     │
     └── Music_I_Like
           ├── Other_Artist
           │     └── Some_Album
           │           ├── Great_Song.mp3
           │           └── AwesomeSong.mp3
           ├── ...
           └── Some_Artist
                 └── Some_Album
                       ├── Great_Song.mp3
                       └── AwesomeSong.mp3

```
Lets take a look at this Directory structure! 

The first thing to notice is that in the root directory of the mounted 
path (the path where the filesystem is mounted), there are folders 
with the Artists names, one folder per Artist.

When MuLi scans the music files to get the Tags information it changes
the names to make them compatible with every operative system and filesystem.
It removes the special characters and replaces the spaces with underscores,
but only in the Directory and Files names. It does not modify the 
real names stored in the music files!

Inside every Artist song there are Directories that match every Album
in the Music Library, as it happens in the Artists Directory names, the 
names in the Albums are also modified.

Finally, inside every Album are the Songs! The songs can be read, moved,
modified and deleted without any problem. But be careful! When the Song
is deleted, it is deleted from the origin path too!!

When a Song is moved from one path to another inside the MuLi filesystem,
the Tags inside the Song file are also updated. This makes the Music 
Library consistent and keeps every Song updated!
If you create or copy a new Song file inside any folder, the Tags inside the
file will be modified accordingly.

Directoriee and Songs can be created and moved and it modifies the Tags
on the Songs and creates or modifies Artists and Albums.
Again, be careful! If you delete a Directory it will be PERMANENT for the
Songs inside it!

There are two special directories in the filesystem:

1. drop: Every file that is stored here will be scanned and moved to the 
correct location depending on the Tags it contains. If you have a new file
that you want to add to the Music Library and you don't want to create
the parent Directories, just drop it here!

2. playlists: This Directory manages the playlists, for every playlist
in the Source Directory, all the files inside it are analyzed and 
the same Directory structure will be created. Then a playlist will
contain Artists and Albums folders for every Song inside it. The format
used in playlists is M3U.


Description files
-----------------

As MuLi modifies the names for the Tags to be compatible with different
filesystems, it also generates a special file inside every Directory
called .description.

The description files are the only files allowed to start with a dot in
the MuLi filesystem, it contains a JSON with the information of the
containing Album or Artist.

For example, in the previous file structure, the Some_Artist directory
would contain a Description file as this one:

```json
{
  "ArtistName":"Some Artist",
  "ArtistPath":"Some_Artist",
  "ArtistAlbums":
    [
      "Some_Album", 
      "Other_Album"
    ]
}
```
Here all the information for the Artist can be processed and read.
On the other hand, for the Album Directories the example would be
like this one in the Other_Album folder:

```json
{
  "AlbumName":"Other Album",
  "AlbumPath":"OtherAlbum"
}
```

Every special character will be removed, also the dots and the spaces
are replaced with underscores.


Information Storage
-------------------

All the information about the Music Library that MuLi uses and gathers is
stored in a [Bolt](https://github.com/boltdb/bolt) database (or Object store
if you prefer).
More information about it [here](https://github.com/dankomiocevic/mulifs/tree/master/store)


Requirements
------------

* A computer running a flavour of *nix


Dependencies
------------

* [github.com/bazil/fuse](https://github.com/bazil/fuse)
* [github.com/boltdb/bolt](https://github.com/boltdb/bolt)
* [github.com/golang/glog](https://github.com/golang/glog)

MuLi is based on the awesome [Bazil's](https://github.com/bazil) 
implementation of [FUSE](https://github.com/bazil/fuse) purely in Go.
It uses FUSE to generate the filesystem in userspace.

It also uses the great and simple [BoltDB](https://github.com/boltdb/bolt)
to store all the information of Songs, Artists and Albums.

To manage the logs it uses the Glog library for Go.

If you don't know these projects take a look at them!


Installation (if you are not familiar with Go)
----------------------------------------------

1. Follow this link to install Go and set up your environment:
   
  [https://golang.org/doc/install](https://golang.org/doc/install)
  (don't forget to set up your GOPATH)

2. Download, compile and install MuLi by running the following command:
  
```
go get github.com/dankomiocevic/mulifs
```


Running MuLi
------------

To start the Music Library Filesystem run:

```
mulifs [global_options] MUSIC_SOURCE MOUNTPOINT 
```


### Params ###
* MUSIC_SOURCE: The path of the folder containing the music files.
* MOUNTPOINT: The path where MuLi should be mounted.

### Global Options ###
* alsologtostderr: log to standard error as well as files
* db_path string: Database path. (default "muli.db")
* log_backtrace_at value: when logging hits line file:N, emit a stack trace (default :0)
* log_dir string: If non-empty, write log files in this directory
* logtostderr: log to standard error instead of files
* stderrthreshold value: logs at or above this threshold go to stderr
* v value: log level for V logs
* vmodule value: comma-separated list of pattern=N settings for file-filtered logging


License
-------

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
