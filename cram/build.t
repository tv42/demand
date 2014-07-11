  $ cat >foo <<EOF
  > go:
  >   import: github.com/tv42/humanize-bytes/cmd/bytes2human
  > EOF
  $ umask 022
  $ XDG_CONFIG_HOME="$PWD/home" DEMAND_CACHE_DIR="$PWD/cache" demand -build foo
  $ find "$PWD/cache/bin/$(go env GOOS)_$(go env GOARCH)" -mindepth 1 \
  >  -printf 'name=%P\nmode=%M\ntype=%y\nlinks=%n\n'
  name=foo
  mode=-rwxr-xr-x
  type=f
  links=1
