package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"gopkg.in/v1/yaml"
)

var (
	upgrade   = flag.Bool("upgrade", false, "force upgrade even if older version exists")
	gopath    = flag.Bool("gopath", false, "use GOPATH from environment instead of downloading all dependencies")
	onlyBuild = flag.Bool("build", false, "only build, do not run command (can pass multiple spec files)")
)

func getCacheDir() (string, error) {
	var cacheDir string
	cacheDir = os.Getenv("DEMAND_CACHE_DIR")
	if cacheDir == "" {
		u, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %v", err)
		}
		cacheDir = filepath.Join(u.HomeDir, ".cache/demand")
	}
	return cacheDir, nil
}

// Create directories if they don't exist.
func maybeMkdirs(perm os.FileMode, paths ...string) error {
	for _, path := range paths {
		err := os.Mkdir(path, perm)
		if err != nil && !os.IsExist(err) {
			return err
		}
	}
	return nil
}

// like os.Readlink but returns empty string and success for non-symlinks
func maybeReadlink(path string) (string, error) {
	dest, err := os.Readlink(path)
	if err == nil {
		return dest, nil
	}
	err2, ok := err.(*os.PathError)
	if !ok {
		return path, err
	}
	switch err2.Err {
	case syscall.EINVAL, syscall.ENOENT:
		// not a symlink, never mind
		return "", nil
	default:
		return path, err
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

type specification struct {
	Go struct {
		Import string
	}
}

func build(cacheDir, cacheBinDir, cacheBinArchDir string,
	specPath string, specFile io.Reader, binary string) error {
	err := maybeMkdirs(0750, cacheDir, cacheBinDir, cacheBinArchDir)
	if err != nil {
		return fmt.Errorf("cannot create cache directory: %v", err)
	}

	var spec specification
	specData, err := ioutil.ReadAll(specFile)
	err = yaml.Unmarshal(specData, &spec)
	if err != nil {
		return fmt.Errorf("cannot parse spec file: %v", err)
	}
	if spec.Go.Import == "" {
		return fmt.Errorf("spec file does not specify import path: %s", specPath)
	}

	tmpGopath, err := ioutil.TempDir("", "demand-gopath-")
	if err != nil {
		return fmt.Errorf("cannot create temp directory: %v", err)
	}
	defer func() {
		err := os.RemoveAll(tmpGopath)
		if err != nil {
			log.Printf("tempdir cleanup failed: %v", err)
		}
	}()

	envGopath := tmpGopath
	if *gopath {
		oldGopath := os.Getenv("GOPATH")
		if oldGopath != "" {
			envGopath = envGopath + string(filepath.ListSeparator) + oldGopath
		}
	}
	env := copyEnvWithGopath(envGopath)

	// TODO -upgrade should be handled just by the fact that we have a
	// clean GOPATH, but double check what happens on -gopath -upgrade

	// need to do this in two steps, as "go get" won't let us control
	// destination
	cmd := exec.Command("go", "get", "-d", "--", spec.Go.Import)
	cmd.Dir = tmpGopath
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = env
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("could not get go package: %v", err)
	}

	// poor man's tempfile atomicity; go build complains if
	// destination exists
	tmpBin := fmt.Sprintf("%s.%d.tmp", binary, os.Getpid())
	defer func() {
		err := os.Remove(tmpBin)
		if err != nil && !os.IsNotExist(err) {
			log.Printf("temp binary cleanup failed: %v", err)
		}
	}()
	cmd = exec.Command("go", "build", "-o", tmpBin, "--", spec.Go.Import)
	cmd.Dir = tmpGopath
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = env
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("could not build go package: %v", err)
	}

	err = os.Rename(tmpBin, binary)
	if err != nil {
		return fmt.Errorf("could put new binary in place: %v", err)
	}
	return nil
}

func doit(args []string) error {
	specPath := args[0]

	specBase := filepath.Base(specPath)
	if specBase[0] == '.' {
		return fmt.Errorf("refusing to run hidden spec file: %s", specPath)
	}

	dest, err := maybeReadlink(specPath)
	if err != nil {
		return fmt.Errorf("readlink: %v", err)
	}
	if dest != "" {
		specBase = filepath.Base(dest)
	}

	// open it here to guard against typos; we don't need to read
	// until we know it's a cache miss
	specFile, err := os.Open(specPath)
	if err != nil {
		return fmt.Errorf("cannot open spec file: %v", err)
	}
	// we rely on Go's automatic use of O_CLOEXEC to close this in
	// syscall.Exec, this defer is a nice gesture for error cases
	defer func() {
		// silence errcheck
		_ = specFile.Close()
	}()

	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}
	cacheBinDir := filepath.Join(cacheDir, "bin")
	arch := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	cacheBinArchDir := filepath.Join(cacheBinDir, arch)

	binary := filepath.Join(cacheBinArchDir, specBase)

	if !*onlyBuild && !*upgrade {
		err = runBinary(binary, flag.Args(), os.Environ())
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("cannot exec %s: %v", binary, err)
		}
	}

	// if we're still here, we don't have a cached binary, or we're
	// upgrading
	err = build(cacheDir, cacheBinDir, cacheBinArchDir, specPath, specFile, binary)
	if err != nil {
		return err
	}

	if !*onlyBuild {
		// now run it (again); this time ENOENT means trouble
		err = runBinary(binary, flag.Args(), os.Environ())
		if err != nil {
			return fmt.Errorf("cannot exec %s: %v", binary, err)
		}
	}
	return nil
}

func dobuild(specs []string) error {
	var err error
	for i := range specs {
		args := specs[i : i+1]
		err = doit(args)
		if err != nil {
			return err
		}
	}
	return nil
}

var prog = filepath.Base(os.Args[0])

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", prog)
	fmt.Fprintf(os.Stderr, "  %s [OPTS] SPEC_PATH [ARGS..]\n", prog)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Only build, do not run command:\n")
	fmt.Fprintf(os.Stderr, "  %s -build [OPTS] SPEC_PATH..\n", prog)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Use as an interpreter:\n")
	fmt.Fprintf(os.Stderr, "  #!/usr/bin/env demand\n")
	fmt.Fprintf(os.Stderr, "  go:\n")
	fmt.Fprintf(os.Stderr, "    import: GO_IMPORT_PATH_HERE\n")
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(prog + ": ")

	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	var err error
	if *onlyBuild {
		err = dobuild(flag.Args())
	} else {
		err = doit(flag.Args())
	}

	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
