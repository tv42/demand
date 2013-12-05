  $ cat >foo <<EOF
  > go:
  >   import: github.com/tv42/humanize-bytes/cmd/bytes2human
  > EOF
  $ C="$PWD/cache"
  $ C_BIN="$PWD/cache/bin/$(go env GOOS)_$(go env GOARCH)"
  $ mkdir -p -- "$C_BIN"
  $ cat >"$C_BIN/foo" <<EOF
  > #!/bin/sh
  > echo mock cached binary
  > EOF
  $ chmod a+x -- "$C_BIN/foo"
  $ DEMAND_CACHE_DIR="$C" demand foo
  mock cached binary
