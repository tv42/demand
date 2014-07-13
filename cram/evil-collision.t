Cache must differentiate entries with the same basename.

  $ mkdir bad good
  $ cat >bad/foo <<EOF
  > go:
  >   import: github.com/tv42/demand/testutil/notcalled
  > EOF
  $ cat >good/foo <<EOF
  > go:
  >   import: github.com/tv42/demand/testutil/argv0
  > EOF
  $ DEMAND_CACHE_DIR="$PWD/cache" demand -build -gopath bad/foo
  $ DEMAND_CACHE_DIR="$PWD/cache" demand -gopath good/foo
  foo
