  $ demand -help
  Usage of demand:
    demand [OPTS] SPEC_PATH [ARGS..]
  
  Only build, do not run command:
    demand -build [OPTS] SPEC_PATH..
  
  Options:
    -build=false: only build, do not run command (can pass multiple spec files)
    -gopath=false: use GOPATH from environment instead of downloading all dependencies
    -upgrade=false: force upgrade even if older version exists
  
  Use as an interpreter:
    #!/usr/bin/env demand
    go:
      import: GO_IMPORT_PATH_HERE
  [2]

  $ demand
  Usage of demand:
    demand [OPTS] SPEC_PATH [ARGS..]
  
  Only build, do not run command:
    demand -build [OPTS] SPEC_PATH..
  
  Options:
    -build=false: only build, do not run command (can pass multiple spec files)
    -gopath=false: use GOPATH from environment instead of downloading all dependencies
    -upgrade=false: force upgrade even if older version exists
  
  Use as an interpreter:
    #!/usr/bin/env demand
    go:
      import: GO_IMPORT_PATH_HERE
  [2]
