How MuLi manages the Playlists?
===============================

There is a special directory in MuLi structure called *playlists*, this directory manages music playlists in m3u format. 
The playlists are managed in directories like the rest of the filesystem, but it has some restrictions related to the files that the m3u file can contain, the main restriction is that it can only handle files that are already stored in the MuLi filesystem.

Let's take a look at an example of a managed m3u file:

```ini
#EXTM3U

#MULI Some_Artist - Some_Album - Some_song
/path/to/file/Some_song.mp3
#MULI Some_Artist - Some_Album - Other_song
/path/to/file/Other_song.mp3
#MULI Some_Artist - Other_Album - Great_song 
/path/to/file/Great_song.mp3
```

Before each line there is a #MULI tag that defines where is the file located in the MuLi structure. This information is generated in order to maintain the data after the filesystem is unmounted.

It is not advisable to modify the playlist from the external folder instead of using MuLi since it only updates the files when the filesystem is loaded. If there is a line that does not have a #MULI tag before the song path, that line will be ignored and will be deleted when the playlist is generated again.

MuLi regenerates the playlists every time there is a change in one of the songs or there is a change in the playlist structure.
