  $ cat >.foo <<EOF
  > %%%i am not even yaml!
  > EOF
  $ C="$PWD/cache"
  $ DEMAND_CACHE_DIR="$C" demand .foo
  demand: refusing to run hidden spec file: .foo
  [1]
  $ DEMAND_CACHE_DIR="$C" demand "$PWD/.foo"
  demand: refusing to run hidden spec file: .*/\.foo (re)
  [1]
