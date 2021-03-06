How MuLi stores the information?
================================

MuLi keeps a database with all the information about the Songs, Albums and Artists in a [BoltDB](https://github.com/boltdb/bolt).
BoltDB is a simple yet powerful key/value store written in Go, as MuLi does not perform any query or search function and only keeps
of the organization of the Music Library structure Bolt is the best choice for the task.


Structure
---------

The store is organized in the same way than the file structure. Bolt allow us to write Keys/Values and Buckets:

* Key/Value: Is a simple storage of a specific Value (for example Artist Information) under the track of a specific Key 
              (for example the Artist name). 
* Buckets: The buckets work like directories, they allow to store Key/Values and other Buckets under a specific Key. For
              example we store the Albums inside an Artist bucket.
              
The structure we are using to organize the files is like the following:

```
Artists (Bucket)
│
├── Some_Artist (Bucket)
│    ├── .description (Key/Value)
│    │
│    ├── Some_Album (Bucket)
│    │     ├── .description (Key/Value)
│    │     └── Some_song (Key/Value)
│    │ 
│    └── Other_Album (Bucket)
│          ├── .description (Key/Value)
│          ├── More_songs (Key/Value)
│          ├── ...
│          └── Other_song (Key/Value)
│
└── Other_Artist (Bucket)
     ├── .description (Key/Value)
     │
     └── Some_Album (Bucket)
           ├── .description (Key/Value)
           ├── Great_Song (Key/Value)
           ├── ...
           └── AwesomeSong (Key/Value)

```

The following statements are true:

* All the Buckets are inside a root Bucket called "Artists", this is the main Bucket where all the others are located. 
* Every Key inside "Artists" is an Artist and is a Bucket, not a Key/Value. 
* Every Artist Bucket contains Album Buckets and a ".description" Key/Value with the information of the Artist.
* Every Album Bucket contains Song Key/Values and a ".description" Key/Value with the information of the Album.


Opening the store
-----------------
Here is a short snippet that shows how to open the Bolt store:

```Go
db, err := bolt.Open("db/path/muli.db", 0600, nil)
if err != nil {
   return nil, err
}
defer db.Close()
```


Reading the Artists
-------------------

The following is a snippet that explains how to read all the Artists in the store using Bolt:

```Go
  err := db.View(func(tx *bolt.Tx) error {
    // First, get the root Bucket (Artists)
    b := tx.Bucket([]byte("Artists"))
    // Create a cursor to Iterate the values.
    c := b.Cursor()
    for k, v := c.First(); k != nil; k, v = c.Next() {
      // When the value is nil, it is a Bucket.
      if v == nil {
        fmt.Printf("Artist: %s\n", k)
      }
    }
    return nil
  })
```
This code will print all the Buckets (nil value) inside Artists Bucket.


Reading the Artist information
------------------------------

The following snippet shows how to read the information for an Artist:

```Go
  err = db.View(func(tx *bolt.Tx) error {
    // First, get the root bucket (Artists)
    root := tx.Bucket([]byte("Artists"))
    // Now get the specific Artist Bucket
    // inside the previous one.
    b := root.Bucket([]byte("Some_Artist"))
    if b == nil {
      return errors.New("Artist not found.")
    }
    
    // Now get the description JSON
    artistJson := b.Get([]byte(".description"))
    if artistJson == nil {
      return errors.New("Description not found.")
    }
    
    // Of course, the JSON will need some processing
    // to get the values, here we just print it.
    fmt.Printf("Description: %s\n", artistJson)
    return nil
  })
```

The Artist information is a JSON containing the Real Artist Name (the one with the special characters),
the Directory Artist Name (the modified one that is compatible with most filesystems) and an array with
all the Albums this Artist has.

For example:
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

Reading the Albums
------------------

The following snippet lists the Albums for an Artist:
```Go
  err := db.View(func(tx *bolt.Tx) error {
    // First, get the root Bucket (Artists)
    root := tx.Bucket([]byte("Artists"))
    // Now get the specific Artist Bucket
    // inside the previous one.
    b := root.Bucket([]byte("Some_Artist"))
    if b == nil {
      return errors.New("Artist not found.")
    }
    
    // Create a cursor to Iterate the values.
    c := b.Cursor()
    for k, v := c.First(); k != nil; k, v = c.Next() {
      // When the value is nil, it is a Bucket.
      if v == nil {
        fmt.Printf("Album: %s\n", k)
      }
    }
    return nil
  })
```


Reading an Album description
----------------------------

The following snippet shows how to read the information for an Album:

```Go
  err = db.View(func(tx *bolt.Tx) error {
    // First, get the root bucket (Artists)
    root := tx.Bucket([]byte("Artists"))
    // Now get the specific Artist Bucket
    // inside the previous one.
    b := root.Bucket([]byte("Some_Artist"))
    if b == nil {
      return errors.New("Artist not found.")
    }
    
    // Then, get the specific Album Bucket
    // inside the previous one.
    c := root.Bucket([]byte("Other_Album"))
    if c == nil {
      return errors.New("Album not found.")
    }
    
    // Now get the description JSON
    albumJson := c.Get([]byte(".description"))
    if albumJson == nil {
      return errors.New("Description not found.")
    }
    
    // Of course, the JSON will need some processing
    // to get the values, here we just print it.
    fmt.Printf("Description: %s\n", albumJson)
    return nil
  })
```

The Album information is a JSON containing the Real Album Name (the one with the special characters) and
the Directory Album Name (the modified one that is compatible with most filesystems).

For example:
```json
{
  "AlbumName":"Other Album",
  "AlbumPath":"OtherAlbum"
}
```


Reading the Songs in an Album
-----------------------------

This code iterates through all the songs in an album:
```Go
  err := db.View(func(tx *bolt.Tx) error {
    // First, get the root Bucket (Artists)
    root := tx.Bucket([]byte("Artists"))
    // Now get the specific Artist Bucket
    // inside the previous one.
    b := root.Bucket([]byte("Some_Artist"))
    if b == nil {
      return errors.New("Artist not found.")
    }
    
    // Then, get the specific Album Bucket
    // inside the previous one.
    c := root.Bucket([]byte("Other_Album"))
    if c == nil {
      return errors.New("Album not found.")
    }
    
    // Create a cursor to Iterate the values.
    d := c.Cursor()
    for k, v := d.First(); k != nil; k, v = d.Next() {
      // Skip the description and nil values
      if v != nil && k != ".description" {
        fmt.Printf("Song: %s\n", k)
        fmt.Printf("Description: %s\n", v)
      }
    }
    return nil
  })
```

All the Song values contain a JSON object with information about the Song,
Real Name and File Name.
