Basename of argv[0] matches basename of the demand script.

  $ cat >this-is-my-basename <<EOF
  > go:
  >   import: github.com/tv42/demand/testutil/argv0
  > EOF
  $ DEMAND_CACHE_DIR="$PWD/cache" demand -gopath this-is-my-basename
  this-is-my-basename
