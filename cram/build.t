  $ T="$(mktemp -d --suffix=".demand.cram")"
  $ trap "rm -rf -- \"$T\"" EXIT
  $ cat >"$T/foo" <<EOF
  > go:
  >   import: github.com/tv42/humanize-bytes/cmd/bytes2human
  > EOF
  $ DEMAND_CACHE_DIR="$T/cache" demand -run=false "$T/foo"
  $ find "$T/cache/bin/$(go env GOOS)_$(go env GOARCH)" -mindepth 1 \
  >  -printf 'name=%P\nmode=%M\ntype=%y\nlinks=%n\n'
  name=foo
  mode=-rwxrwxr-x
  type=f
  links=1
