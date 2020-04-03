# seedboxsync

Tool that syncs downloaded series from a seedbox, to a local player.

## Requirements

You'll need:

* Go >= 1.14
* SSH agent running on both machines, with support for public/private key authentication
* A folder to where the downloads are copied once complete - the tool can't work against a directory in which the downloads are progressing

## Build & configure

### Build and copy sample files

```bash
go build
cp sample.db storage.db
cp sample.seedboxsync .seedboxsync
```

### Edit configuration

Edit the configuration file `.seedboxsync` which is already populated with some sample configuration parameters.

**host** hostname or IP address of the machine<br/>
**port** port to where the SSH agent is listening<br/>
**user** user with which the application should login<br/>
**key** location of the private key to use to SSH<br/>
**dir** root directory where the contents are read/written<br/>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*seedbox* - represents the location of where the completed downloads are found<br/>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*player* - represents the root of where the series should be placed<br/>
**temp_dir** temporary directory to where the rar'd/zip'd files should be extracted prior to be transferred (only applies to the *seedbox* block)

**IMPORTANT** The files should be placed on the directory once they're completely downloaded. This application does not know if a download is in progress or not.

### Add keys

You need to add the private keys used to SSH into both the seedbox and the player machine. You also need to have them already added to your `~/.ssh/known_hosts` file. So make sure that you've at least SSH'd to those machines once through your terminal (easier way to add them to the file).

## Execute

```bash
$ ./seedboxsync

== Configuration ==
{
  "seedbox": {
    "host": "example.com",
    "port": 2222,
    "user": "example",
    "key": "keys/seedbox",
    "dir": "/home/example/downloads",
    "temp_dir": "/temp"
  },
  "player": {
    "host": "player.example.com",
    "port": 22,
    "user": "player",
    "key": "keys/player",
    "dir": "/home/example/series"
  }
}
== Contents ==
Adding 'XXX.XXX.S16E20.1080p.WEB.H264-iNSiDiOUS' to queue
Copying XXX.XXX.S16E20.1080p.WEB.H264-iNSiDiOUS complete
Hashing '/home/example/downloads/XXX.XXX.S16E20.1080p.WEB.H264-iNSiDiOUS/'
Refreshing player's library...
Done
```
