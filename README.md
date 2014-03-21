# picsizer

On-demand image resizing in Go. Work in progress.

The ultimate aim is to build something that will work as a missing file handler
in conjunction with S3, by pulling down originals, resizing them, uploading
them, then redirecting to the resized version.

I'm also using this to learn Go.

## TODO

* Prevent path-traversal attacks
* Convert only once when the same resource is requested simultaneously
* Allow specification of configuration file on command line
* Make directory and file masks configurable
* S3 stuff
* Hot-reload configuration
