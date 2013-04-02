  $ demand -help
  Usage of demand:
    demand [OPTS] SPEC_PATH [ARGS..]
    -gopath=false: use GOPATH from environment instead of downloading all dependencies
    -run=true: run the command, can be disabled to just ensure caching
    -upgrade=false: force upgrade even if older version exists
  
  Use as an interpreter:
    #!/usr/bin/env demand
    go:
      import: GO_IMPORT_PATH_HERE
  [2]

  $ demand
  Usage of demand:
    demand [OPTS] SPEC_PATH [ARGS..]
    -gopath=false: use GOPATH from environment instead of downloading all dependencies
    -run=true: run the command, can be disabled to just ensure caching
    -upgrade=false: force upgrade even if older version exists
  
  Use as an interpreter:
    #!/usr/bin/env demand
    go:
      import: GO_IMPORT_PATH_HERE
  [2]
