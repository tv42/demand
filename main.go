package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

var upgrade = flag.Bool("upgrade", false, "force upgrade even if older version exists")

var gopath = flag.Bool("gopath", false, "use GOPATH from environment instead of downloading all dependencies")

var run = flag.Bool("run", true, "run the command, can be disabled to just ensure caching")

func getCacheDir() string {
	var cache_dir string
	cache_dir = os.Getenv("DEMAND_CACHE_DIR")
	if cache_dir == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatalf("cannot determine home directory: %v", err)
		}
		cache_dir = filepath.Join(u.HomeDir, ".cache/demand")
	}
	return cache_dir
}

// Create directory if it doesn't exist. Fail fatally if not possible.
func mustMaybeMkdir(path string, perm os.FileMode) {
	err := os.Mkdir(path, perm)
	if err != nil && !os.IsExist(err) {
		log.Fatalf("cannot create directory: %v", err)
	}
}

// copy environment, but override GOPATH
func copyEnvWithGopath(gopath string) []string {
	old := os.Environ()
	env := make([]string, 0, len(old))
	for _, kv := range old {
		if strings.HasPrefix(kv, "GOPATH=") {
			continue
		}
		env = append(env, kv)
	}

	env = append(env, "GOPATH="+gopath)
	return env
}

func runBinary(binary string, args []string, env []string) error {
	// can't use os/exec etc, we want exec not fork+exec, and we want
	// to pass all fds to the child
	return syscall.Exec(binary, args, env)
}

type Spec struct {
	Go struct {
		Import string
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s [OPTS] SPEC_PATH [ARGS..]\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Use as an interpreter:\n")
	fmt.Fprintf(os.Stderr, "  #!/usr/bin/env demand\n")
	fmt.Fprintf(os.Stderr, "  go:\n")
	fmt.Fprintf(os.Stderr, "    import: GO_IMPORT_PATH_HERE\n")
}

func main() {
	prog := filepath.Base(os.Args[0])
	log.SetFlags(0)
	log.SetPrefix(prog + ": ")

	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	spec_path := flag.Arg(0)

	spec_base := filepath.Base(spec_path)
	if spec_base[0] == '.' {
		log.Fatalf("refusing to run hidden spec file: %s", spec_path)
	}

	// open it here to guard against typos; we don't need to read
	// until we know it's a cache miss
	spec_file, err := os.Open(spec_path)
	if err != nil {
		log.Fatalf("cannot open spec file: %v", err)
	}
	defer spec_file.Close()

	cache_dir := getCacheDir()
	cache_bin_dir := filepath.Join(cache_dir, "bin")
	arch := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	cache_bin_arch_dir := filepath.Join(cache_bin_dir, arch)

	binary := filepath.Join(cache_bin_arch_dir, spec_base)

	if *run && !*upgrade {
		err = runBinary(binary, flag.Args(), os.Environ())
		if err != nil && !os.IsNotExist(err) {
			log.Fatalf("cannot exec %s: %v", binary, err)
		}
	}

	// if we're still here, we don't have a cached binary, or we're
	// upgrading

	mustMaybeMkdir(cache_dir, 0750)
	mustMaybeMkdir(cache_bin_dir, 0750)
	mustMaybeMkdir(cache_bin_arch_dir, 0750)

	var spec Spec
	spec_data, err := ioutil.ReadAll(spec_file)
	err = goyaml.Unmarshal(spec_data, &spec)
	if err != nil {
		log.Fatalf("cannot parse spec file: %v", err)
	}
	if spec.Go.Import == "" {
		log.Fatalf("spec file does not specify import path: %s", spec_path)
	}

	tmp_gopath, err := ioutil.TempDir("", "demand-gopath-")
	if err != nil {
		log.Fatalf("cannot create temp directory: %v", err)
	}
	defer func() {
		err := os.RemoveAll(tmp_gopath)
		if err != nil {
			log.Printf("tempdir cleanup failed: %v", err)
		}
	}()

	env_gopath := tmp_gopath
	if *gopath {
		old_gopath := os.Getenv("GOPATH")
		if old_gopath != "" {
			env_gopath = env_gopath + string(filepath.ListSeparator) + old_gopath
		}
	}
	env := copyEnvWithGopath(env_gopath)

	// TODO -upgrade should be handled just by the fact that we have a
	// clean GOPATH, but double check what happens on -gopath -upgrade

	// need to do this in two steps, as "go get" won't let us control
	// destination
	cmd := exec.Command("go", "get", "-d", "--", spec.Go.Import)
	cmd.Dir = tmp_gopath
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = env
	err = cmd.Run()
	if err != nil {
		log.Fatalf("could not get go package: %v", err)
	}

	// poor man's tempfile atomicity; go build complains if
	// destination exists
	tmp_bin := fmt.Sprintf("%s.%d.tmp", binary, os.Getpid())
	defer func() {
		err := os.Remove(tmp_bin)
		if err != nil && !os.IsNotExist(err) {
			log.Printf("temp binary cleanup failed: %v", err)
		}
	}()
	cmd = exec.Command("go", "build", "-o", tmp_bin, "--", spec.Go.Import)
	cmd.Dir = tmp_gopath
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = env
	err = cmd.Run()
	if err != nil {
		log.Fatalf("could not build go package: %v", err)
	}

	err = os.Rename(tmp_bin, binary)
	if err != nil {
		log.Fatalf("could put new binary in place: %v", err)
	}

	if *run {
		// now run it (again); this time ENOENT means trouble
		err = runBinary(binary, flag.Args(), os.Environ())
		if err != nil {
			log.Fatalf("cannot exec %s: %v", binary, err)
		}
	}
}
