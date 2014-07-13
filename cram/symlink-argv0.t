  $ cat >foo <<EOF
  > go:
  >   import: github.com/tv42/demand/testutil/argv0
  > EOF
  $ ln -s foo bar
  $ umask 022
  $ XDG_CONFIG_HOME="$PWD/home" DEMAND_CACHE_DIR="$PWD/cache" demand -gopath bar
  bar
