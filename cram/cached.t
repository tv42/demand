  $ cat >foo <<EOF
  > go:
  >   import: github.com/tv42/demand/testutil/succeed
  > EOF
  $ C="$PWD/cache"
  $ DEMAND_CACHE_DIR="$C" demand -build -gopath foo
  $ # now replace it with something noticable, while not making it seem new
  $ touch -r foo stamp
  $ cat >foo <<EOF
  > go:
  >   import: github.com/tv42/demand/testutil/notcalled
  > EOF
  $ touch -r stamp foo
  $ DEMAND_CACHE_DIR="$C" demand -gopath foo
  ok
