  $ mkdir one two three
  $ cat >one/foo <<EOF
  > go:
  >   import: github.com/tv42/demand/testutil/argv0
  > EOF
  $ C="$PWD/cache"
  $ DEMAND_CACHE_DIR="$C" demand -build -gopath one/foo
  $ # now replace it with something noticable, while not making it seem new
  $ touch -r one/foo stamp
  $ cat >one/foo <<EOF
  > go:
  >   import: github.com/tv42/demand/testutil/notcalled
  > EOF
  $ touch -r stamp one/foo
  $ # run via symlink; must use cached results for linked-to specfile
  $ ln -s ../one/foo two/bar
  $ ln -s ../two/bar three/quux
  $ XDG_CONFIG_HOME="$PWD/home" DEMAND_CACHE_DIR="$C" demand -gopath three/quux
  quux
  $ find "$PWD/cache/bin/$(go env GOOS)_$(go env GOARCH)" -mindepth 1 -type l \
  >  -printf '%P -> %l\n' | sort
  [^/]+!three/quux -> \.\./[^/]+!two/bar (re)
  [^/]+!two/bar -> \.\./[^/]+!one/foo (re)
  $ find "$PWD/cache/bin/$(go env GOOS)_$(go env GOARCH)" -mindepth 1 -type f \
  >  -printf 'name=%P\nmode=%M\ntype=%y\nlinks=%n\n.\n'
  name=[^/]+!one/foo (re)
  mode=-rwxr-xr-x
  type=f
  links=1
  .
