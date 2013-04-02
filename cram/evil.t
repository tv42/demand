  $ T="$(mktemp -d --suffix=".demand.cram")"
  $ trap "rm -rf -- \"$T\"" EXIT
  $ cat >"$T/.foo" <<EOF
  > %%%i am not even yaml!
  > EOF
  $ C="$T/cache"
  $ DEMAND_CACHE_DIR="$C" demand "$T/.foo"
  demand: refusing to run hidden spec file: .*/\.foo (re)
  [1]
