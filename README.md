# infinimesh HTTP FileServer

[![Build Docker Images](https://github.com/infinimesh/http-fs/actions/workflows/ci.yml/badge.svg)](https://github.com/infinimesh/http-fs/actions/workflows/ci.yml)

Simple HTTP FileServer made for infimesh, but can be extended and applied for other services with simillar purposes.

Right now only FileSystem storage is supported.

## General logic

Server understands two things:

- namespaces (folders)
- files (well, just files)

So these are routes:

GET /{ns} - returns stats (files and their props) in the requested namespace
DELETE /{ns} - deletes namespace(and its files)
GET /{ns}/{file} - returns file itself
POST /{ns}/{file} - uploads file
DELETE /{ns}/{file} - deletes file

See the Postman Collection to try it yourself.

## Installation

Docker(compose) service example:

```yaml
    http-fs:
        image: ghcr.io/infinimesh/http-fs:latest
        restart: always
        ports:
            - "80:8000"
        environment:
            ADDR: :8000
            REPO: repo:8000
            STATIC_DIR: /static # you should probably map this to some real volume
            LOG_LEVEL: -1 # -1 for debug, 0 for info, 1 for warning, 2 for error (defaults to info)
```

## Default logic

Namespaces are mapped to:

1. infinimesh namespaces - middleware determines access to the folder basing on the user's permissions
2. filesystem directory

These is the result of using [InfinimeshMiddleware](https://github.com/infinimesh/http-fs/blob/master/pkg/mw/infinimesh.go) and [FileSystem](https://github.com/infinimesh/http-fs/blob/2052af2e6f9ffa67bcb0c2cdbf1ac9f54e550bfd/pkg/io/fs/fs.go) [IOHandler](https://github.com/infinimesh/http-fs/blob/master/pkg/io/fs/fs.go#L17)

## How to extend

### Middleware

Middleware is a `mux` middleware which adds `Access` to context.

Access defined [here](https://github.com/infinimesh/http-fs/blob/2052af2e6f9ffa67bcb0c2cdbf1ac9f54e550bfd/pkg/mw/mw.go). It's a simple structure which defined if requestor has Read and Write access to the namespace/file.

You can also see the [SampleMiddleware](https://github.com/infinimesh/http-fs/blob/master/pkg/mw/sample.go) to (maybe) get a better idea.

### IOHandler

IOHandler is an interface which has few methods decribed [here](https://github.com/infinimesh/http-fs/blob/master/pkg/io/io.go#L18)

### Building

Replace used middlewares and IOHandler with your own and compile.
