# Copyright 2016 Danko Miocevic. All Rights Reserved.
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
# http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Author: Danko Miocevic

# This script tests the MuLi filesystem.
# For more information please read the README.md contained on 
# this folder.

# Config options
MULI_X='../mulifs'
SRC_DIR='./testSrc'
DST_DIR='./testDst'
PWD_DIR=$(pwd)
TEST_SIZE=2

# Tag setting function
# This function is used to set the tags to the MP3 files.
# The first argument is the file.
# The second argument is the Artist.
# The third argument is the Album.
# The fourth argument is the Title.
# This function can be modified in order to use another
# tag editor command.
# Returns the id3 command return value.
function set_tags {
  id3tag --artist=$2 --album=$3 --song=$4 $1 &> /dev/null
  return $?
}

# Tag checking function
# This function is used to check that the tags are correct
# in a specific MP3 file.
# The first argument is the file.
# The second argument is the Artist.
# The third argument is the Album.
# The fourth argument is the Title.
# This function can be modified in order to use another
# tag editor command.
# Returns 0 if the tags are correct.
function check_tags {
  local ARTIST=$(id3info $1 | grep TPE1)
  local ALBUM=$(id3info $1 | grep TALB)
  local TITLE=$(id3info $1 | grep TIT2)
  if [ ${ARTIST#*:} != $2 ]; then
    return 1
  fi

  if [ ${ALBUM#*:} != $3 ]; then
    return 2
  fi

  if [ ${TITLE#*:} != $4 ]; then
    return 3
  fi
  return 0
}

# Create fake MP3s function
function create_fake {
  cd $PWD_DIR
  echo "Creating lots of fake files!"
  local ARTIST_COUNT=$TEST_SIZE

  while [ $ARTIST_COUNT -gt 0 ]; do
    echo "Creating ARTIST GreatArtist$ARTIST_COUNT"
    local ALBUM_COUNT=$TEST_SIZE
    while [ $ALBUM_COUNT -gt 0 ]; do
      local SONG_COUNT=$TEST_SIZE
      while [ $SONG_COUNT -gt 0 ]; do
        cp "test.mp3" "$SRC_DIR/testAr${ARTIST_COUNT}Al${ALBUM_COUNT}Sn${SONG_COUNT}.mp3" &> /dev/null
        set_tags "$SRC_DIR/testAr${ARTIST_COUNT}Al${ALBUM_COUNT}Sn${SONG_COUNT}.mp3" "GreatArtist$ARTIST_COUNT" "GreatAlbum$ALBUM_COUNT" "Song$SONG_COUNT"
        if [ $? -ne 0 ]; then
          echo "ERROR in file $SRC_DIR/testAr${ARTIST_COUNT}Al${ALBUM_COUNT}Sn${SONG_COUNT}.mp3"
        fi
        let SONG_COUNT=SONG_COUNT-1
      done
      let ALBUM_COUNT=ALBUM_COUNT-1
    done
    let ARTIST_COUNT=ARTIST_COUNT-1
  done
}

# Check that all the files were in the right place.
function check_fake {
  cd $PWD_DIR
  echo "Checking the fake files."
  local ARTIST_COUNT=$TEST_SIZE

  while [ $ARTIST_COUNT -gt 0 ]; do
    cd $PWD_DIR
    if [ -d "$DST_DIR/GreatArtist$ARTIST_COUNT" ]; then
      cd "$DST_DIR/GreatArtist$ARTIST_COUNT"
      local ALBUM_COUNT=$TEST_SIZE
      while [ $ALBUM_COUNT -gt 0 ]; do
        if [ -d "GreatAlbum$ALBUM_COUNT" ]; then
          cd "GreatAlbum$ALBUM_COUNT"
          local SONG_COUNT=$TEST_SIZE
          while [ $SONG_COUNT -gt 0 ]; do
            if [ -f "Song${SONG_COUNT}.mp3" ]; then
              if [ ! -s "Song${SONG_COUNT}.mp3" ]; then
                echo "ERROR: File ${DST_DIR}/GreatArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT}/Song${SONG_COUNT} has 0 size"
              fi
            else
              echo "ERROR: File ${DST_DIR}/GreatArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT}/Song${SONG_COUNT} not exists"
            fi
            let SONG_COUNT=SONG_COUNT-1
          done
          cd ..
        else 
          echo "ERROR: Directory ${DST_DIR}/GreatArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT} not exists"
        fi
        let ALBUM_COUNT=ALBUM_COUNT-1
      done
    else
      echo "ERROR: Directory ${DST_DIR}/GreatArtist${ARTIST_COUNT} not exists"
    fi
    let ARTIST_COUNT=ARTIST_COUNT-1
  done
}

# Copy artists around
function copy_artists {
  cd $PWD_DIR
  cd $DST_DIR
  echo "Copying Artists around."
  local ARTIST_COUNT=$TEST_SIZE
  while [ $ARTIST_COUNT -gt 0 ] ; do 
    if [ -d "GreatArtist$ARTIST_COUNT" ]; then
      cp -r "GreatArtist$ARTIST_COUNT" "OtherArtist$ARTIST_COUNT" &> /dev/null
    fi
    let ARTIST_COUNT=ARTIST_COUNT-1
  done
}

# Check copied artists
function check_copied_artists {
  cd $PWD_DIR
  cd $DST_DIR
  echo "Checking copied Artists"

  local ARTIST_COUNT=$TEST_SIZE
  while [ $ARTIST_COUNT -gt 0 ]; do
    cd $PWD_DIR
    if [ -d "$DST_DIR/OtherArtist$ARTIST_COUNT" ]; then
      cd "$DST_DIR/OtherArtist$ARTIST_COUNT"
      local ALBUM_COUNT=$TEST_SIZE
      while [ $ALBUM_COUNT -gt 0 ]; do
        if [ -d "GreatAlbum$ALBUM_COUNT" ]; then
          cd "GreatAlbum$ALBUM_COUNT"
          local SONG_COUNT=$TEST_SIZE
          while [ $SONG_COUNT -gt 0 ]; do
            if [ -f "Song${SONG_COUNT}.mp3" ]; then
              if [ ! -s "Song${SONG_COUNT}.mp3" ]; then
                echo "ERROR: File ${DST_DIR}/OtherArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT}/Song${SONG_COUNT} has 0 size"
              else
                check_tags Song${SONG_COUNT}.mp3 OtherArtist$ARTIST_COUNT GreatAlbum$ALBUM_COUNT Song${SONG_COUNT}
                if [ $? -ne 0 ] ; then
                  echo "ERROR: File ${DST_DIR}/OtherArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT}/Song${SONG_COUNT} tags not match"
                fi
              fi
            else
              echo "ERROR: File ${DST_DIR}/OtherArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT}/Song${SONG_COUNT} not exists"
            fi
            let SONG_COUNT=SONG_COUNT-1
          done
          cd ..
        else 
          echo "ERROR: Directory ${DST_DIR}/GreatArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT} not exists"
        fi
        let ALBUM_COUNT=ALBUM_COUNT-1
      done
    else
      echo "ERROR: Directory ${DST_DIR}/GreatArtist${ARTIST_COUNT} not exists"
    fi
    let ARTIST_COUNT=ARTIST_COUNT-1
  done
}

# Pre-Mount function
function create_dirs {
  cd $PWD_DIR
  echo -n "Creating dirs..."
  mkdir $SRC_DIR $DST_DIR
  if [ $? -eq 0 ] ; then
      echo "OK!"
  else
    echo "ERROR"
  fi
}

# Mount FS function
function mount_muli {
  cd $PWD_DIR
  echo -n "Mounting MuLi filesystem..."
  # Run MuLi in the background
  $MULI_X $SRC_DIR $DST_DIR &> muli.log &
  while [ ! -d "$DST_DIR/drop" ]; do
    sleep 1
  done
  echo "OK!"
}

# Umount FS function
function umount_muli {
  cd $PWD_DIR
  echo -n "Umounting MuLi filesystem..."
  # Umount the destination directory
  umount $DST_DIR
  if [ $? -eq 0 ] ; then
    echo "OK!"
  else
    echo "ERROR"
  fi
}

# Clean everything up
function clean_up {
  cd $PWD_DIR
  echo -n "Cleaning everything up..."
  # Delete the SRC and DST folders
  rm -rf $SRC_DIR $DST_DIR muli.db
  if [ $? -eq 0 ] ; then
    echo "OK!"
  else
    echo "ERROR"
  fi
}

# Perform tests
create_dirs
create_fake
mount_muli
check_fake
copy_artists
check_copied_artists
umount_muli
clean_up


