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
MULI_X='../mulifs -alsologtostderr'
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
  id3tag --artist="$2" --album="$3" --song="$4" $1 &> /dev/null
  return $?
}

# Tag stripping function
# This function should strip all the tags from the MP3 file.
# The fist argument is the file.
function strip_tags {
  id3convert -s $1 &> /dev/null
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

# Create empty MP3 function
function create_empty {
  cd $PWD_DIR
  echo -n "Creating empty MP3 files..."
  local HAS_ERROR=0

  cp test.mp3 $SRC_DIR/empty.mp3 &> /dev/null
  strip_tags $SRC_DIR/empty.mp3
  #TODO: Maybe test here what happens if we only add some tags
  #       like Artist or Album only.
  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Check that empty MP3 exists function
function check_empty {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Checking empty MP3 files..."
  local HAS_ERROR=0

  if [ ! -d "unknown" ] ; then
    echo "ERROR"
    echo "The unknown Artist directory does not exist."
    return 1
  fi
  cd unknown

  if [ ! -d "unknown" ] ; then
    echo "ERROR"
    echo "The unknown Album directory does not exist."
    return 2
  fi
  cd unknown

  if [ -f "empty.mp3" ] ; then
    check_tags empty.mp3 unknown unknown empty
    if [ $? -ne 0 ] ; then
      if [ $HAS_ERROR -eq 0 ] ; then
        echo "ERROR"
      fi
      HAS_ERROR=1
      echo "ERROR: File unknown/unknown/empty.mp3 tags not match"
    fi
  else
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "The empty.mp3 file cannot be found."
  fi

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Create MP3 with special characters function
function create_special {
  cd $PWD_DIR
  echo -n "Creating special MP3 files..."
  local HAS_ERROR=0

  cp test.mp3 $SRC_DIR/special.mp3 &> /dev/null
  set_tags $SRC_DIR/special.mp3 'Increíble Artista' 'Suco de Aça' 'Canció'
  #TODO: Maybe test here what happens if we only add some tags
  #       like Artist or Album only.
  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Check that MP3 with special characters
# exists function
function check_special {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Checking special MP3 files..."
  local HAS_ERROR=0

  if [ ! -d "Increible_Artista" ] ; then
    echo "ERROR"
    echo "The special Artist directory does not exist."
    return 1
  fi
  cd Increible_Artista 

  if [ ! -d "Suco_de_Acai" ] ; then
    echo "ERROR"
    echo "The special Album directory does not exist."
    return 2
  fi
  cd Suco_de_Acai 

  if [ -f "Cancion.mp3" ] ; then
    check_tags $SRC_DIR/special.mp3 "Increíble Artista!" "Suco de Açaí" "Canción"
    if [ $? -ne 0 ] ; then
      if [ $HAS_ERROR -eq 0 ] ; then
        echo "ERROR"
      fi
      HAS_ERROR=1
      echo "ERROR: File Increible_Artista/Suco_de_Acai/Cancion.mp3 tags not match"
    fi
  else
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "The Cancion.mp3 file cannot be found."
  fi

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Create fake MP3s function
function create_fake {
  cd $PWD_DIR
  echo -n "Creating lots of fake files..."
  local ARTIST_COUNT=$TEST_SIZE
  local HAS_ERROR=0

  while [ $ARTIST_COUNT -gt 0 ]; do
    local ALBUM_COUNT=$TEST_SIZE
    while [ $ALBUM_COUNT -gt 0 ]; do
      local SONG_COUNT=$TEST_SIZE
      while [ $SONG_COUNT -gt 0 ]; do
        cp "test.mp3" "$SRC_DIR/testAr${ARTIST_COUNT}Al${ALBUM_COUNT}Sn${SONG_COUNT}.mp3" &> /dev/null
        set_tags "$SRC_DIR/testAr${ARTIST_COUNT}Al${ALBUM_COUNT}Sn${SONG_COUNT}.mp3" "GreatArtist$ARTIST_COUNT" "GreatAlbum$ALBUM_COUNT" "Song$SONG_COUNT"
        if [ $? -ne 0 ]; then
          if [ $HAS_ERROR -eq 0 ] ; then
            echo "ERROR"
          fi
          HAS_ERROR=1
          echo "ERROR in file $SRC_DIR/testAr${ARTIST_COUNT}Al${ALBUM_COUNT}Sn${SONG_COUNT}.mp3"
        fi
        let SONG_COUNT=SONG_COUNT-1
      done
      let ALBUM_COUNT=ALBUM_COUNT-1
    done
    let ARTIST_COUNT=ARTIST_COUNT-1
  done

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Create fake MP3s function
function drop_files {
  cd $PWD_DIR
  echo -n "Droping files..."
  local ARTIST_COUNT=$TEST_SIZE
  local HAS_ERROR=0

  while [ $ARTIST_COUNT -gt 0 ]; do
    local ALBUM_COUNT=$TEST_SIZE
    while [ $ALBUM_COUNT -gt 0 ]; do
      local SONG_COUNT=$TEST_SIZE
      while [ $SONG_COUNT -gt 0 ]; do
        cp "test.mp3" "$DST_DIR/drop/Artist${ARTIST_COUNT}Album${ALBUM_COUNT}Song${SONG_COUNT}.mp3" &> /dev/null
        set_tags "$DST_DIR/drop/Artist${ARTIST_COUNT}Album${ALBUM_COUNT}Song${SONG_COUNT}.mp3" "GreatArtist$ARTIST_COUNT" "GreatAlbum$ALBUM_COUNT" "Song$SONG_COUNT"
        if [ $? -ne 0 ]; then
          if [ $HAS_ERROR -eq 0 ] ; then
            echo "ERROR"
          fi
          HAS_ERROR=1
          echo "ERROR in file $DST_DIR/drop/Artist${ARTIST_COUNT}Album${ALBUM_COUNT}Song${SONG_COUNT}.mp3"
        fi
        let SONG_COUNT=SONG_COUNT-1
      done
      let ALBUM_COUNT=ALBUM_COUNT-1
    done
    let ARTIST_COUNT=ARTIST_COUNT-1
  done

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
  sleep 5
}

# Check that all the files were in the right place.
function check_fake {
  cd $PWD_DIR
  echo -n "Checking the fake files..."
  local ARTIST_COUNT=$TEST_SIZE
  local HAS_ERROR=0

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
                if [ $HAS_ERROR -eq 0 ] ; then
                  echo "ERROR"
                fi
                HAS_ERROR=1
                echo "ERROR: File ${DST_DIR}/GreatArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT}/Song${SONG_COUNT}.mp3 has 0 size"
              fi
            else
              if [ $HAS_ERROR -eq 0 ] ; then
                echo "ERROR"
              fi
              HAS_ERROR=1
              echo "ERROR: File ${DST_DIR}/GreatArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT}/Song${SONG_COUNT}.mp3 not exists"
            fi
            let SONG_COUNT=SONG_COUNT-1
          done
          cd ..
        else 
          if [ $HAS_ERROR -eq 0 ] ; then
            echo "ERROR"
          fi
          HAS_ERROR=1
          echo "ERROR: Directory ${DST_DIR}/GreatArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT} not exists"
        fi
        let ALBUM_COUNT=ALBUM_COUNT-1
      done
    else
      if [ $HAS_ERROR -eq 0 ] ; then
        echo "ERROR"
      fi
      HAS_ERROR=1
      echo "ERROR: Directory ${DST_DIR}/GreatArtist${ARTIST_COUNT} not exists"
    fi
    let ARTIST_COUNT=ARTIST_COUNT-1
  done

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# function to create directories and fill them with empty files
function mkdirs {
  cd $PWD_DIR
  echo -n "Creating dirs..."
  local ARTIST_COUNT=$TEST_SIZE
  local HAS_ERROR=0

  while [ $ARTIST_COUNT -gt 0 ]; do
    cd $PWD_DIR
    mkdir "$DST_DIR/NewArtist$ARTIST_COUNT" &> /dev/null
    if [ $? -eq 0 ]; then
      cd "$DST_DIR/NewArtist$ARTIST_COUNT"
      local ALBUM_COUNT=$TEST_SIZE
      while [ $ALBUM_COUNT -gt 0 ]; do
        mkdir "NewAlbum$ALBUM_COUNT" &> /dev/null
        if [ $? -eq 0 ]; then
          cd "NewAlbum$ALBUM_COUNT"
          local SONG_COUNT=$TEST_SIZE
          while [ $SONG_COUNT -gt 0 ]; do
            cp $PWD_DIR/test.mp3 NewSong${SONG_COUNT}.mp3 &> /dev/null
            strip_tags NewSong${SONG_COUNT}.mp3 
            let SONG_COUNT=SONG_COUNT-1
          done
          cd ..
        else 
          if [ $HAS_ERROR -eq 0 ] ; then
            echo "ERROR"
          fi
          HAS_ERROR=1
          echo "ERROR: Directory ${DST_DIR}/NewArtist${ARTIST_COUNT}/NewAlbum${ALBUM_COUNT} cannot be created"
        fi
        let ALBUM_COUNT=ALBUM_COUNT-1
      done
    else
      if [ $HAS_ERROR -eq 0 ] ; then
        echo "ERROR"
      fi
      HAS_ERROR=1
      echo "ERROR: Directory ${DST_DIR}/NewArtist${ARTIST_COUNT} cannot be created"
    fi
    let ARTIST_COUNT=ARTIST_COUNT-1
  done

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}



# Check that all the created files are in the right place.
function check_mkdirs {
  cd $PWD_DIR
  echo -n "Checking the created files..."
  local ARTIST_COUNT=$TEST_SIZE
  local HAS_ERROR=0

  while [ $ARTIST_COUNT -gt 0 ]; do
    cd $PWD_DIR
    if [ -d "$DST_DIR/NewArtist$ARTIST_COUNT" ]; then
      cd "$DST_DIR/NewArtist$ARTIST_COUNT"
      local ALBUM_COUNT=$TEST_SIZE
      while [ $ALBUM_COUNT -gt 0 ]; do
        if [ -d "NewAlbum$ALBUM_COUNT" ]; then
          cd "NewAlbum$ALBUM_COUNT"
          local SONG_COUNT=$TEST_SIZE
          while [ $SONG_COUNT -gt 0 ]; do
            if [ -f "NewSong${SONG_COUNT}.mp3" ]; then
              if [ ! -s "NewSong${SONG_COUNT}.mp3" ]; then
                if [ $HAS_ERROR -eq 0 ] ; then
                  echo "ERROR"
                fi
                HAS_ERROR=1
                echo "ERROR: File ${DST_DIR}/NewArtist${ARTIST_COUNT}/NewAlbum${ALBUM_COUNT}/NewSong${SONG_COUNT}.mp3 has 0 size"
              else
                check_tags NewSong${SONG_COUNT}.mp3 NewArtist${ARTIST_COUNT} NewAlbum${ALBUM_COUNT} NewSong${SONG_COUNT}
                if [ $? -ne 0 ] ; then
                  if [ $HAS_ERROR -eq 0 ] ; then
                    echo "ERROR"
                  fi
                  HAS_ERROR=1
                  echo "ERROR: File ${DST_DIR}/NewArtist${ARTIST_COUNT}/NewAlbum${ALBUM_COUNT}/NewSong${SONG_COUNT}.mp3 tags are wrong."
                fi
              fi
            else
              if [ $HAS_ERROR -eq 0 ] ; then
                echo "ERROR"
              fi
              HAS_ERROR=1
              echo "ERROR: File ${DST_DIR}/NewArtist${ARTIST_COUNT}/NewAlbum${ALBUM_COUNT}/NewSong${SONG_COUNT}.mp3 not exists"
            fi
            let SONG_COUNT=SONG_COUNT-1
          done
          cd ..
        else 
          if [ $HAS_ERROR -eq 0 ] ; then
            echo "ERROR"
          fi
          HAS_ERROR=1
          echo "ERROR: Directory ${DST_DIR}/NewArtist${ARTIST_COUNT}/NewAlbum${ALBUM_COUNT} not exists"
        fi
        let ALBUM_COUNT=ALBUM_COUNT-1
      done
    else
      if [ $HAS_ERROR -eq 0 ] ; then
        echo "ERROR"
      fi
      HAS_ERROR=1
      echo "ERROR: Directory ${DST_DIR}/NewArtist${ARTIST_COUNT} not exists"
    fi
    let ARTIST_COUNT=ARTIST_COUNT-1
  done

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Move artists 
function move_artists {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Moving Artists..."
  local ARTIST_COUNT=$TEST_SIZE

  while [ $ARTIST_COUNT -gt 0 ] ; do 
    if [ -d "GreatArtist$ARTIST_COUNT" ]; then
      mv "GreatArtist$ARTIST_COUNT" "DifferentArtist$ARTIST_COUNT" &> /dev/null
    fi
    let ARTIST_COUNT=ARTIST_COUNT-1
  done

  echo "OK!"
}

# Check moved Artists
function check_moved_artists {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Checking moved Artists..."

  local ARTIST_COUNT=$TEST_SIZE
  local HAS_ERROR=0
  while [ $ARTIST_COUNT -gt 0 ]; do
    cd $PWD_DIR
    if [ -d "$DST_DIR/DifferentArtist$ARTIST_COUNT" ]; then
      cd "$DST_DIR/DifferentArtist$ARTIST_COUNT"
      local ALBUM_COUNT=$TEST_SIZE
      while [ $ALBUM_COUNT -gt 0 ]; do
        if [ -d "GreatAlbum$ALBUM_COUNT" ]; then
          cd "GreatAlbum$ALBUM_COUNT"
          local SONG_COUNT=$TEST_SIZE
          while [ $SONG_COUNT -gt 0 ]; do
            if [ -f "Song${SONG_COUNT}.mp3" ]; then
              if [ ! -s "Song${SONG_COUNT}.mp3" ]; then
                if [ $HAS_ERROR -eq 0 ] ; then
                  echo "ERROR"
                fi
                HAS_ERROR=1
                echo "ERROR: File ${DST_DIR}/DifferentArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT}/Song${SONG_COUNT}.mp3 has 0 size"
              else
                check_tags Song${SONG_COUNT}.mp3 DifferentArtist$ARTIST_COUNT GreatAlbum$ALBUM_COUNT Song${SONG_COUNT}
                if [ $? -ne 0 ] ; then
                  if [ $HAS_ERROR -eq 0 ] ; then
                    echo "ERROR"
                  fi
                  HAS_ERROR=1
                  echo "ERROR: File ${DST_DIR}/DifferentArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT}/Song${SONG_COUNT}.mp3 tags not match"
                fi
              fi
            else
              if [ $HAS_ERROR -eq 0 ] ; then
                echo "ERROR"
              fi
              HAS_ERROR=1
              echo "ERROR: File ${DST_DIR}/DifferentArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT}/Song${SONG_COUNT}.mp3 not exists"
            fi
            let SONG_COUNT=SONG_COUNT-1
          done
          cd ..
        else 
          if [ $HAS_ERROR -eq 0 ] ; then
            echo "ERROR"
          fi
          HAS_ERROR=1
          echo "ERROR: Directory ${DST_DIR}/DifferentArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT} not exists"
        fi
        let ALBUM_COUNT=ALBUM_COUNT-1
      done
    else
      if [ $HAS_ERROR -eq 0 ] ; then
        echo "ERROR"
      fi
      HAS_ERROR=1
      echo "ERROR: Directory ${DST_DIR}/DifferentArtist${ARTIST_COUNT} not exists"
    fi
    let ARTIST_COUNT=ARTIST_COUNT-1
  done

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Copy artists around
function copy_artists {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Copying Artists around..."
  local ARTIST_COUNT=$TEST_SIZE

  while [ $ARTIST_COUNT -gt 0 ] ; do 
    if [ -d "GreatArtist$ARTIST_COUNT" ]; then
      cp -r "GreatArtist$ARTIST_COUNT" "OtherArtist$ARTIST_COUNT" &> /dev/null
    fi
    let ARTIST_COUNT=ARTIST_COUNT-1
  done

  echo "OK!"
}

# Check copied Artists
function check_copied_artists {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Checking copied Artists..."

  local ARTIST_COUNT=$TEST_SIZE
  local HAS_ERROR=0
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
                if [ $HAS_ERROR -eq 0 ] ; then
                  echo "ERROR"
                fi
                HAS_ERROR=1
                echo "ERROR: File ${DST_DIR}/OtherArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT}/Song${SONG_COUNT}.mp3 has 0 size"
              else
                check_tags Song${SONG_COUNT}.mp3 OtherArtist$ARTIST_COUNT GreatAlbum$ALBUM_COUNT Song${SONG_COUNT}
                if [ $? -ne 0 ] ; then
                  if [ $HAS_ERROR -eq 0 ] ; then
                    echo "ERROR"
                  fi
                  HAS_ERROR=1
                  echo "ERROR: File ${DST_DIR}/OtherArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT}/Song${SONG_COUNT}.mp3 tags not match"
                fi
              fi
            else
              if [ $HAS_ERROR -eq 0 ] ; then
                echo "ERROR"
              fi
              HAS_ERROR=1
              echo "ERROR: File ${DST_DIR}/OtherArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT}/Song${SONG_COUNT}.mp3 not exists"
            fi
            let SONG_COUNT=SONG_COUNT-1
          done
          cd ..
        else 
          if [ $HAS_ERROR -eq 0 ] ; then
            echo "ERROR"
          fi
          HAS_ERROR=1
          echo "ERROR: Directory ${DST_DIR}/OtherArtist${ARTIST_COUNT}/GreatAlbum${ALBUM_COUNT} not exists"
        fi
        let ALBUM_COUNT=ALBUM_COUNT-1
      done
    else
      if [ $HAS_ERROR -eq 0 ] ; then
        echo "ERROR"
      fi
      HAS_ERROR=1
      echo "ERROR: Directory ${DST_DIR}/OtherArtist${ARTIST_COUNT} not exists"
    fi
    let ARTIST_COUNT=ARTIST_COUNT-1
  done

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Delete copied artists
function delete_artists {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Deleting copied Artists..."

  local ARTIST_COUNT=$TEST_SIZE
  local HAS_ERROR=0
  while [ $ARTIST_COUNT -gt 0 ]; do
    cd $PWD_DIR
    if [ -d "$DST_DIR/OtherArtist$ARTIST_COUNT" ]; then
      rm -rf "$DST_DIR/OtherArtist$ARTIST_COUNT" &> /dev/null
      if [ $? -eq 0 ] ; then
        if [ -d "$DST_DIR/OtherArtist$ARTIST_COUNT" ]; then
          if [ $HAS_ERROR -eq 0 ] ; then
            echo "ERROR"
          fi
          HAS_ERROR=1
          echo "ERROR: Directory ${DST_DIR}/OtherArtist${ARTIST_COUNT} has not been deleted."
        fi 
      else
        if [ $HAS_ERROR -eq 0 ] ; then
          echo "ERROR"
        fi
        HAS_ERROR=1
        echo "ERROR: Directory ${DST_DIR}/OtherArtist${ARTIST_COUNT} cannot be deleted."
      fi
    else
      if [ $HAS_ERROR -eq 0 ] ; then
        echo "ERROR"
      fi
      HAS_ERROR=1
      echo "ERROR: Directory ${DST_DIR}/OtherArtist${ARTIST_COUNT} not exists"
    fi
    let ARTIST_COUNT=ARTIST_COUNT-1
  done

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Move albums 
function move_albums {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Moving Albums..."

  local HAS_ERROR=0
  if [ ! -d "DifferentArtist1" ]; then
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "ERROR: Cannot find DifferentArtist1"
    return
  fi

  cd DifferentArtist1

  local ALBUM_COUNT=$TEST_SIZE
  while [ $ALBUM_COUNT -gt 0 ] ; do 
    if [ -d "GreatAlbum$ALBUM_COUNT" ]; then
      mv "GreatAlbum$ALBUM_COUNT" "DifferentAlbum$ALBUM_COUNT" &> /dev/null
    fi
    let ALBUM_COUNT=ALBUM_COUNT-1
  done
  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Check moved albums
function check_moved_albums {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Checking moved Albums..."
  local HAS_ERROR=0
  if [ ! -d "DifferentArtist1" ]; then
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "ERROR: Cannot find DifferentArtist1"
    return
  fi
  cd DifferentArtist1

  local ALBUM_COUNT=$TEST_SIZE
  while [ $ALBUM_COUNT -gt 0 ]; do
    if [ -d "DifferentAlbum$ALBUM_COUNT" ]; then
      cd "DifferentAlbum$ALBUM_COUNT"
      local SONG_COUNT=$TEST_SIZE

      while [ $SONG_COUNT -gt 0 ]; do
        if [ -f "Song${SONG_COUNT}.mp3" ]; then
          if [ ! -s "Song${SONG_COUNT}.mp3" ]; then
            if [ $HAS_ERROR -eq 0 ] ; then
              echo "ERROR"
            fi
            HAS_ERROR=1
            echo "ERROR: File ${DST_DIR}/DifferentArtist1/DifferentAlbum${ALBUM_COUNT}/Song${SONG_COUNT}.mp3 has 0 size"
          else
            check_tags Song${SONG_COUNT}.mp3 DifferentArtist1 DifferentAlbum$ALBUM_COUNT Song${SONG_COUNT}
            if [ $? -ne 0 ] ; then
              if [ $HAS_ERROR -eq 0 ] ; then
                echo "ERROR"
              fi
              HAS_ERROR=1
              echo "ERROR: File ${DST_DIR}/DifferentArtist1/DifferentAlbum${ALBUM_COUNT}/Song${SONG_COUNT}.mp3 tags not match"
            fi
          fi
        else
          if [ $HAS_ERROR -eq 0 ] ; then
            echo "ERROR"
          fi
          HAS_ERROR=1
          echo "ERROR: File ${DST_DIR}/DifferentArtist1/DifferentAlbum${ALBUM_COUNT}/Song${SONG_COUNT}.mp3 not exists"
        fi
        let SONG_COUNT=SONG_COUNT-1
      done
      cd ..
    else 
      if [ $HAS_ERROR -eq 0 ] ; then
        echo "ERROR"
      fi
      HAS_ERROR=1
      echo "ERROR: Directory ${DST_DIR}/DifferentArtist1/DifferentAlbum${ALBUM_COUNT} not exists"
    fi
    let ALBUM_COUNT=ALBUM_COUNT-1
  done

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Copy albums around
function copy_albums {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Copying Albums around..."

  local HAS_ERROR=0
  if [ ! -d "GreatArtist1" ]; then
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "ERROR: Cannot find GreatArtist1"
    return
  fi

  cd GreatArtist1

  local ALBUM_COUNT=$TEST_SIZE
  while [ $ALBUM_COUNT -gt 0 ] ; do 
    if [ -d "GreatAlbum$ALBUM_COUNT" ]; then
      cp -r "GreatAlbum$ALBUM_COUNT" "OtherAlbum$ALBUM_COUNT" &> /dev/null
    fi
    let ALBUM_COUNT=ALBUM_COUNT-1
  done
  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Check copied albums
function check_copied_albums {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Checking copied Albums..."
  local HAS_ERROR=0
  if [ ! -d "GreatArtist1" ]; then
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "ERROR: Cannot find GreatArtist1"
    return
  fi
  cd GreatArtist1

  local ALBUM_COUNT=$TEST_SIZE
  while [ $ALBUM_COUNT -gt 0 ]; do
    if [ -d "OtherAlbum$ALBUM_COUNT" ]; then
      cd "OtherAlbum$ALBUM_COUNT"
      local SONG_COUNT=$TEST_SIZE

      while [ $SONG_COUNT -gt 0 ]; do
        if [ -f "Song${SONG_COUNT}.mp3" ]; then
          if [ ! -s "Song${SONG_COUNT}.mp3" ]; then
            if [ $HAS_ERROR -eq 0 ] ; then
              echo "ERROR"
            fi
            HAS_ERROR=1
            echo "ERROR: File ${DST_DIR}/GreatArtist1/OtherAlbum${ALBUM_COUNT}/Song${SONG_COUNT}.mp3 has 0 size"
          else
            check_tags Song${SONG_COUNT}.mp3 GreatArtist1 OtherAlbum$ALBUM_COUNT Song${SONG_COUNT}
            if [ $? -ne 0 ] ; then
              if [ $HAS_ERROR -eq 0 ] ; then
                echo "ERROR"
              fi
              HAS_ERROR=1
              echo "ERROR: File ${DST_DIR}/GreatArtist1/OtherAlbum${ALBUM_COUNT}/Song${SONG_COUNT}.mp3 tags not match"
            fi
          fi
        else
          if [ $HAS_ERROR -eq 0 ] ; then
            echo "ERROR"
          fi
          HAS_ERROR=1
          echo "ERROR: File ${DST_DIR}/GreatArtist1/OtherAlbum${ALBUM_COUNT}/Song${SONG_COUNT}.mp3 not exists"
        fi
        let SONG_COUNT=SONG_COUNT-1
      done
      cd ..
    else 
      if [ $HAS_ERROR -eq 0 ] ; then
        echo "ERROR"
      fi
      HAS_ERROR=1
      echo "ERROR: Directory ${DST_DIR}/GreatArtist1/OtherAlbum${ALBUM_COUNT} not exists"
    fi
    let ALBUM_COUNT=ALBUM_COUNT-1
  done

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Delete copied albums
function delete_albums {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Deleting copied Albums..."
  local HAS_ERROR=0
  if [ ! -d "GreatArtist1" ]; then
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "ERROR: Cannot find GreatArtist1"
    return
  fi
  cd GreatArtist1

  local ALBUM_COUNT=$TEST_SIZE
  while [ $ALBUM_COUNT -gt 0 ]; do
    if [ -d "OtherAlbum$ALBUM_COUNT" ]; then
      cd "OtherAlbum$ALBUM_COUNT"

      rm -rf "OtherAlbum$ALBUM_COUNT" &> /dev/null
      if [ $? -eq 0 ] ; then
        if [ -d "OtherAlbum$ALBUM_COUNT" ]; then
          if [ $HAS_ERROR -eq 0 ] ; then
            echo "ERROR"
          fi
          HAS_ERROR=1
          echo "ERROR: Directory ${DST_DIR}/GreatArtist1/OtherAlbum${ALBUM_COUNT} has not been deleted."
        fi
      else
        if [ $HAS_ERROR -eq 0 ] ; then
          echo "ERROR"
        fi
        HAS_ERROR=1
        echo "ERROR: Directory ${DST_DIR}/GreatArtist1/OtherAlbum${ALBUM_COUNT} cannot be deleted."
      fi
      cd ..
    else 
      if [ $HAS_ERROR -eq 0 ] ; then
        echo "ERROR"
      fi
      HAS_ERROR=1
      echo "ERROR: Directory ${DST_DIR}/GreatArtist1/OtherAlbum${ALBUM_COUNT} not exists"
    fi
    let ALBUM_COUNT=ALBUM_COUNT-1
  done

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Move songs
function move_songs {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Moving Songs..."

  local HAS_ERROR=0
  if [ ! -d "DifferentArtist1" ]; then
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "ERROR: Cannot find DifferentArtist1"
    return
  fi

  cd DifferentArtist1

  if [ ! -d "DifferentAlbum1" ]; then
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "ERROR: Cannot find DifferentAlbum1"
    return
  fi

  cd DifferentAlbum1
  local SONG_COUNT=$TEST_SIZE
  while [ $SONG_COUNT -gt 0 ] ; do 
    if [ -f "Song$SONG_COUNT.mp3" ]; then
      cp "Song$SONG_COUNT.mp3" "DifferentSong$SONG_COUNT.mp3" &> /dev/null
    fi
    let SONG_COUNT=SONG_COUNT-1
  done
  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Check moved songs 
function check_moved_songs {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Checking moved Songs..."
  local HAS_ERROR=0
  if [ ! -d "DifferentArtist1" ]; then
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "ERROR: Cannot find DifferentArtist1"
    return
  fi
  cd DifferentArtist1

  if [ ! -d "DifferentAlbum1" ]; then
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "ERROR: Cannot find DifferentAlbum1"
    return
  fi
  cd DifferentAlbum1

  local SONG_COUNT=$TEST_SIZE

  while [ $SONG_COUNT -gt 0 ]; do
    if [ -f "DifferentSong${SONG_COUNT}.mp3" ]; then
      if [ ! -s "DifferentSong${SONG_COUNT}.mp3" ]; then
        if [ $HAS_ERROR -eq 0 ] ; then
          echo "ERROR"
        fi
        HAS_ERROR=1
        echo "ERROR: File ${DST_DIR}/DifferentArtist1/DifferentAlbum1/DifferentSong${SONG_COUNT}.mp3 has 0 size"
      else
        check_tags DifferentSong${SONG_COUNT}.mp3 DifferentArtist1 DifferentAlbum1 DifferentSong${SONG_COUNT}
        if [ $? -ne 0 ] ; then
          if [ $HAS_ERROR -eq 0 ] ; then
            echo "ERROR"
          fi
          HAS_ERROR=1
          echo "ERROR: File ${DST_DIR}/DifferentArtist1/DifferentAlbum1/DifferentSong${SONG_COUNT}.mp3 tags not match"
        fi
      fi
    else
      if [ $HAS_ERROR -eq 0 ] ; then
        echo "ERROR"
      fi
      HAS_ERROR=1
      echo "ERROR: File ${DST_DIR}/DifferentArtist1/DifferentAlbum1/DifferentSong${SONG_COUNT}.mp3 not exists"
    fi
    let SONG_COUNT=SONG_COUNT-1
  done

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Copy songs around
function copy_songs {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Copying Songs around..."

  local HAS_ERROR=0
  if [ ! -d "GreatArtist1" ]; then
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "ERROR: Cannot find GreatArtist1"
    return
  fi

  cd GreatArtist1

  if [ ! -d "GreatAlbum1" ]; then
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "ERROR: Cannot find GreatAlbum1"
    return
  fi

  cd GreatAlbum1
  local SONG_COUNT=$TEST_SIZE
  while [ $SONG_COUNT -gt 0 ] ; do 
    if [ -f "Song$SONG_COUNT.mp3" ]; then
      cp "Song$SONG_COUNT.mp3" "OtherSong$SONG_COUNT.mp3" &> /dev/null
    fi
    let SONG_COUNT=SONG_COUNT-1
  done
  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Check copied songs 
function check_copied_songs {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Checking copied Songs..."
  local HAS_ERROR=0
  if [ ! -d "GreatArtist1" ]; then
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "ERROR: Cannot find GreatArtist1"
    return
  fi
  cd GreatArtist1

  if [ ! -d "GreatAlbum1" ]; then
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "ERROR: Cannot find GreatAlbum1"
    return
  fi
  cd GreatAlbum1

  local SONG_COUNT=$TEST_SIZE

  while [ $SONG_COUNT -gt 0 ]; do
    if [ -f "OtherSong${SONG_COUNT}.mp3" ]; then
      if [ ! -s "OtherSong${SONG_COUNT}.mp3" ]; then
        if [ $HAS_ERROR -eq 0 ] ; then
          echo "ERROR"
        fi
        HAS_ERROR=1
        echo "ERROR: File ${DST_DIR}/GreatArtist1/GreatAlbum1/OtherSong${SONG_COUNT}.mp3 has 0 size"
      else
        check_tags OtherSong${SONG_COUNT}.mp3 GreatArtist1 GreatAlbum1 OtherSong${SONG_COUNT}
        if [ $? -ne 0 ] ; then
          if [ $HAS_ERROR -eq 0 ] ; then
            echo "ERROR"
          fi
          HAS_ERROR=1
          echo "ERROR: File ${DST_DIR}/GreatArtist1/GreatAlbum1/OtherSong${SONG_COUNT}.mp3 tags not match"
        fi
      fi
    else
      if [ $HAS_ERROR -eq 0 ] ; then
        echo "ERROR"
      fi
      HAS_ERROR=1
      echo "ERROR: File ${DST_DIR}/GreatArtist1/GreatAlbum1/OtherSong${SONG_COUNT}.mp3 not exists"
    fi
    let SONG_COUNT=SONG_COUNT-1
  done

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Delete songs
function delete_songs {
  cd $PWD_DIR
  cd $DST_DIR
  echo -n "Deleting copied Songs..."
  local HAS_ERROR=0
  if [ ! -d "GreatArtist1" ]; then
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "ERROR: Cannot find GreatArtist1"
    return
  fi
  cd GreatArtist1

  if [ ! -d "GreatAlbum1" ]; then
    if [ $HAS_ERROR -eq 0 ] ; then
      echo "ERROR"
    fi
    HAS_ERROR=1
    echo "ERROR: Cannot find GreatAlbum1"
    return
  fi
  cd GreatAlbum1

  local SONG_COUNT=$TEST_SIZE

  while [ $SONG_COUNT -gt 0 ]; do
    if [ -f "OtherSong${SONG_COUNT}.mp3" ]; then
      rm -f OtherSong${SONG_COUNT}.mp3 &> /dev/null
      if [ $? -ne 0 ] ; then
        if [ $HAS_ERROR -eq 0 ] ; then
          echo "ERROR"
        fi
        HAS_ERROR=1
        echo "ERROR: File ${DST_DIR}/GreatArtist1/GreatAlbum1/OtherSong${SONG_COUNT}.mp3 cannot be deleted."
      else
        if [ -f "OtherSong${SONG_COUNT}.mp3" ]; then
          if [ $HAS_ERROR -eq 0 ] ; then
            echo "ERROR"
          fi
          HAS_ERROR=1
          echo "ERROR: File ${DST_DIR}/GreatArtist1/GreatAlbum1/OtherSong${SONG_COUNT}.mp3 has not been deleted."
        fi
      fi
    else
      if [ $HAS_ERROR -eq 0 ] ; then
        echo "ERROR"
      fi
      HAS_ERROR=1
      echo "ERROR: File ${DST_DIR}/GreatArtist1/GreatAlbum1/OtherSong${SONG_COUNT}.mp3 not exists"
    fi
    let SONG_COUNT=SONG_COUNT-1
  done

  if [ $HAS_ERROR -eq 0 ] ; then
    echo "OK!"
  fi
}

# Pre-Mount function
function create_dirs {
  cd $PWD_DIR
  echo -n "Creating dirs..."
  mkdir $SRC_DIR $DST_DIR &> /dev/null
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
create_empty
create_special
mount_muli
check_fake
check_empty
check_special
copy_artists
check_copied_artists
copy_albums
check_copied_albums
copy_songs
check_copied_songs
delete_songs
delete_albums
delete_artists
move_artists
check_moved_artists
move_albums
check_moved_albums
move_songs
check_moved_songs
mkdirs
check_mkdirs
drop_files
check_fake
umount_muli
clean_up


