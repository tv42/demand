  $ cat >foo <<EOF
  > go:
  >   import: github.com/tv42/demand/testutil/notcalled
  > EOF
  $ C="$PWD/cache"
  $ DEMAND_CACHE_DIR="$C" demand -build -gopath foo
  $ cat >foo <<EOF
  > go:
  >   import: github.com/tv42/demand/testutil/succeed
  > EOF
  $ DEMAND_CACHE_DIR="$C" demand -gopath foo
  ok
