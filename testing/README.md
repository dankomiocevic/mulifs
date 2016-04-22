MuLiFS testing module
=====================

The idea behind this testing module is to force MuLiFS to perform all the tasks it would do in a typical environment.

As it is a FileSystem using unit testing was more like forcing a testing tool to something that is not right. To test MuLi it is necessary to initialize many parts and everything works as a group.

I thought (and help me here if you have a better idea) that MuLi should be tested as a FileSystem and the best way to test it is creating a Bash script since it is the tool that will most probably manage MuLi.

So I created a really long script that does all the things that MuLi would do in a real environment, those are listed in the following sections.

Please feel free to add more testing code to the script and help me to improve this FileSystem, just send me a Pull Request.


How it works
------------
There is an MP3 file that I created (some time ago for a background music in a game) and that file will be modified (change the Tags) in order to generate more files.

Then the file will be copied around and modified doing all the necessary testing and checks.


Testing modules
---------------
The script will test the following:

- Create lots of mp3 files and add different Tags in order to force MuLi to create the directory structure. 
- Check the .description files. (WIP)
- Test the Copy command (Artists, Albums and songs). (WIP)
- Test the Rename command (Artists, Albums and songs). (WIP)
- Test the Delete command (Artists, Albums and songs). (WIP)
- Test the MkDir comand (Artists, Albums and songs). (WIP)
- Test the Drop directory (throw new files and existing files). (WIP)
- Test the Playlist Rename command (Artists, Albums and songs). (WIP)
- Test the Playlist Copy command (Artists, Albums and songs). (WIP)
- Test the Playlist Delete command (Artists, Albums and songs). (WIP)
- Test the Playlist MkDir command (Artists, Albums and songs). (WIP)


Requirements
------------
The requirements to make this tool work are the following:

- Have a MuLiFS compiled binary.
- Have a ID3Tag tool installed.

I am using MAC default id3 tool and the script works with that, if you are using a different one, there are two funcions in the script (set_tags and check_tags) that need to be modified in order to use the new tool.
