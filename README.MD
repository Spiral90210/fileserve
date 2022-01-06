# fileserve

I needed to serve up some files in a specific directory structure, and couldn't decide between HTTP and FTP as a means. FTP meant easier code (no html parsing), but an awful human experience.

I wrote fileserve to give me the best of both worlds - a human acceptable view for when it was needed, and a simple json api for the computers to use.

## Usage (command)

Requires go >= 1.17

```shell
go run ./cmd/fileserve --listen=127.0.0.1:8007 --data-dir=/var/data --include-hidden
```

## Usage (docker)

The docker command serves `/var/data` on `:8007` of the container, excluding hidden files:

```shell
docker run --rm -it -v "$(pwd):/var/data" -p 8007:8007 spiral90210/fileserve:latest
```

You can include hidden files by appending the `--include-hidden` flag at the end.

## The API

By default, directories will render as a html view of their contents, and files can be downloaded via links. However, add an `Accept: application/json` header to the http request and get a json response instead. This is intended to make it slightly easier to consume for apps. Note paths are relative to the server root (and no support yet for running behind a reverse proxy, sorry!)

An example json response is below, for a server configured to share the source root directory. Note that `path` is relative to the server root, so you can just download it.

`GET http://localhost:8007/cmd/fileserve`

```json
{
  "path": "/cmd/fileserve",
  "contents": [
    {
      "name": "favicon.ico",
      "size": 15406,
      "path": "/cmd/fileserve/favicon.ico",
      "isDir": false
    },
    {
      "name": "main.go",
      "size": 1067,
      "path": "/cmd/fileserve/main.go",
      "isDir": false
    }
  ]
}
```
