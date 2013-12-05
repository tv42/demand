  $ T="$(mktemp -d --suffix=".demand.cram")"
  $ trap "rm -rf -- \"$T\"" EXIT
  $ DEMAND_CACHE_DIR="$T/cache" demand -build "$T/non-existent"
  demand: cannot open spec file: open .*/non-existent: no such file or directory (re)
  [1]

  $ T="$(mktemp -d --suffix=".demand.cram")"
  $ trap "rm -rf -- \"$T\"" EXIT
  $ cat >"$T/foo" <<EOF
  > %%%i am not even yaml!
  > EOF
  $ DEMAND_CACHE_DIR="$T/cache" demand -build "$T/foo"
  demand: cannot parse spec file: .* (re)
  [1]

  $ T="$(mktemp -d --suffix=".demand.cram")"
  $ trap "rm -rf -- \"$T\"" EXIT
  $ cat >"$T/foo" <<EOF
  > # broken on purpose
  > EOF
  $ DEMAND_CACHE_DIR="$T/cache" demand -build "$T/foo"
  demand: spec file does not specify import path: .*/foo (re)
  [1]
