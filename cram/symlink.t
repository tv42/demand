  $ cat >foo <<EOF
  > go:
  >   import: github.com/tv42/demand/testutil/succeed
  > EOF
  $ C="$PWD/cache"
  $ DEMAND_CACHE_DIR="$C" demand -build -gopath foo
  $ # now replace it with something noticable, while not making it seem new
  $ touch -r foo stamp
  $ cat >foo <<EOF
  > go:
  >   import: github.com/tv42/demand/testutil/notcalled
  > EOF
  $ touch -r stamp foo
  $ # run via symlink; must use cached results for linked-to specfile
  $ ln -s foo bar
  $ XDG_CONFIG_HOME="$PWD/home" DEMAND_CACHE_DIR="$C" demand -gopath bar
  ok
  $ find "$PWD/cache/bin/$(go env GOOS)_$(go env GOARCH)" -mindepth 1 -type l \
  >  -printf 'name=%P\ntarget=%l\n.\n'
  name=[^/]+/bar (re)
  target=../[^/]+/foo (re)
  .
  $ find "$PWD/cache/bin/$(go env GOOS)_$(go env GOARCH)" -mindepth 1 -type f \
  >  -printf 'name=%P\nmode=%M\ntype=%y\nlinks=%n\n.\n'
  name=[^/]+/foo (re)
  mode=-rwxr-xr-x
  type=f
  links=1
  .
