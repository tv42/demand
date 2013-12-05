  $ DEMAND_CACHE_DIR="$PWD/cache" demand -build non-existent
  demand: cannot open spec file: open non-existent: no such file or directory
  [1]

  $ cat >foo <<EOF
  > %%%i am not even yaml!
  > EOF
  $ DEMAND_CACHE_DIR="$PWD/cache" demand -build foo
  demand: cannot parse spec file: .* (re)
  [1]

  $ cat >foo <<EOF
  > # broken on purpose
  > EOF
  $ DEMAND_CACHE_DIR="$PWD/cache" demand -build foo
  demand: spec file does not specify import path: foo
  [1]
