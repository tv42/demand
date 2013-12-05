demand -- An easy way to install apps
=====================================

`demand` will download, build, cache and run a Go app.

You can use it as an interpreter. Create a file ``bytes2human`` with
that contains:

    #!/usr/bin/env demand
    go:
      import: github.com/tv42/humanize-bytes/cmd/bytes2human

The format is YAML.

And then run:

    chmod a+x bytes2human

And now you can run that app with just:

    ./bytes2human 65536
    64KiB

If you put that directory in your `PATH`, the dot and slash won't be
needed anymore.

On first run, the source is downloaded, the app is built, and so on.
On later runs, a cached copy of the resulting binary is used.

All command line arguments are passed to the underlying command. To
pass flags to `demand` itself, always place them first, before any
YAML file name.


Caching
-------

Built binaries are stored in `~/.cache/demand/bin/${GOOS}_${GOARCH}/`.
Sharing the cache across different operating systems and architectures
is safe. The cache directory is assumed to have atomic file renames.


Upgrading
---------

If you want to grab a newer version of an app installed with `demand`,
run:

    demand -build -upgrade NAME

where `NAME` is one of:

- a command using `demand` as an interpreter (searched through
  `PATH`)

- a filename of the YAML mentioned above

To force `NAME` to be interpreted as a filename and not searched,
include a slash in it; e.g. `./foo` for relative paths.


Security
--------

`demand` by default runs code controlled by whoever controls the
import url. This is *unsafe*.

However, the default is no different from a `go get` followed by e.g.
`go test` or running the built binary or using the library. Or a `git
clone ... && cd ... && make`. So that may be just fine for you.


Developing commands
-------------------

If you want to use `demand` to run commands without uploading the
source to Github and such, you can tell `demand` to build and cache
the package using your current `GOPATH`:

    demand -build -upgrade -gopath bytes2human

Build results and missing dependencies are still written to a
temporary directory.
