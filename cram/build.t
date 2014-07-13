  $ cat >foo <<EOF
  > go:
  >   import: github.com/tv42/demand/testutil/notcalled
  > EOF
  $ umask 022
  $ XDG_CONFIG_HOME="$PWD/home" DEMAND_CACHE_DIR="$PWD/cache" demand -build -gopath foo
  $ find "$PWD/cache/bin/$(go env GOOS)_$(go env GOARCH)" -mindepth 1 \! -type d \
  >  -printf 'name=%P\nmode=%M\ntype=%y\nlinks=%n\n.\n'
  name=[^/]+/foo (re)
  mode=-rwxr-xr-x
  type=f
  links=1
  .
