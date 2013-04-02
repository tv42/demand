  $ T="$(mktemp -d --suffix=".demand.cram")"
  $ trap "rm -rf -- \"$T\"" EXIT
  $ cat >"$T/foo" <<EOF
  > go:
  >   import: github.com/tv42/humanize-bytes/cmd/bytes2human
  > EOF
  $ C="$T/cache"
  $ C_BIN="$T/cache/bin/$(go env GOOS)_$(go env GOARCH)"
  $ mkdir -p -- "$C_BIN"
  $ cat >"$C_BIN/foo" <<EOF
  > #!/bin/sh
  > echo mock cached binary
  > EOF
  $ chmod a+x -- "$C_BIN/foo"
  $ DEMAND_CACHE_DIR="$C" demand "$T/foo"
  mock cached binary
