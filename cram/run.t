  $ T="$(mktemp -d --suffix=".demand.cram")"
  $ trap "rm -rf -- \"$T\"" EXIT
  $ cat >"$T/foo" <<EOF
  > go:
  >   import: github.com/tv42/humanize-bytes/cmd/bytes2human
  > EOF
  $ DEMAND_CACHE_DIR="$T/cache" demand "$T/foo" 65536
  64KiB
