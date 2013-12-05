  $ cat >foo <<EOF
  > go:
  >   import: github.com/tv42/humanize-bytes/cmd/bytes2human
  > EOF
  $ DEMAND_CACHE_DIR="$PWD/cache" demand foo 65536
  64KiB
