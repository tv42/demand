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

	"gopkg.in/yaml.v1"
)

/*

Cache structure

Given a spec file like

    /home/jdoe/bin/mycommand

the cached result for that is

    ~/.cache/demand/bin/${GOOS}_${GOARCH}/
      !home!jdoe!bin/
        mycommand

This gives us the following desirable properties:

1. Two spec files with the same basename do not collide.

2. Programs can change their behavior based on the name they are
    launched with.

3. Usage messages can include the basename dynamically and still match
    user expectations.

The execve(2) syscall allows the caller to pass an argv[0] that is
independent of the executable being launched. This is undesirable, as
it would prevent the application for re-opening argv[0] to e.g. access
bundled assets. Thus, we make the effort of having the actual
executable exist with the right basename.

If there are also symlinks

    /home/jdoe/bin/myalias -> mycommand
    /home/jdoe/foo/bar -> ../bin/myalias

the cache directory will contain symlinks

    !home!jdoe!bin/myalias -> mycommand
    !home!jdoe!foo/bar -> ../!home!jdoe!bin/myalias

thus enabling the same basename reasoning as above, while building the
command only once.

The exclamation-delimited paths are always absolute (so /specfile can
be cached as !/specfile), and escape all the possible characters into
something that is safe to use as a non-hidden path segment. The exact
rules are an implementation detail.

The cache path depends purely on the demand spec path. This way, if
two demand specs talk about the same import path, they'll be cached
separately, as there might be other options in the spec file that
affect the compilation, or they might just be manually controlled to
be specific versions with -gopath etc. To share cache across spec
files, symlink the spec files.

*/

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
	case syscall.EINVAL:
		// not a symlink, never mind
		return "", nil
	default:
		return path, err
	}
}

// like os.Symlink but also atomically overwrites existing files
func symlink(oldname, newname string) error {
	dir, base := filepath.Split(newname)
	tmp := filepath.Join(dir, fmt.Sprintf(".temp-%s.%d.tmp", base, os.Getpid()))
	if err := os.Symlink(oldname, tmp); err != nil {
		return err
	}
	if err := os.Rename(tmp, newname); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
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

func build(specPath string, specFile io.Reader, cacheSpecDir string, specBase string) error {
	err := maybeMkdirs(0750, cacheSpecDir)
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
	tmpBin := filepath.Join(cacheSpecDir, fmt.Sprintf(".temp-%s.%d.tmp", specBase, os.Getpid()))
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

	err = os.Rename(tmpBin, filepath.Join(cacheSpecDir, specBase))
	if err != nil {
		return fmt.Errorf("could not put new binary in place: %v", err)
	}
	return nil
}

// Escapes a string in a way that is human-readable and safe to use as
// a file basename without collisions. The escaping is never reversed.
func escapeFilename(s string) string {
	// no need for trailing slashes
	s = strings.TrimSuffix(s, "/")
	// first prevent collisions by escaping all of our escape characters
	s = strings.Replace(s, "%", "%25", -1)
	// NIL is illegal file basename
	s = strings.Replace(s, "\x00", "%00", -1)
	// slash is illegal file basename; it's common enough to deserve
	// its own special handling, so move '!' out of the way and use
	// that
	s = strings.Replace(s, "!", "%21", -1)
	s = strings.Replace(s, "/", "!", -1)
	// be friendly toward bad shell scripting and prevent most common
	// whitespace; strictly optional
	s = strings.Replace(s, " ", "%20", -1)
	s = strings.Replace(s, "\n", "%0a", -1)
	// avoid climbing up the directory tree
	if s[0] == '.' {
		s = "%2e" + s[1:]
	}
	return s
}

func cacheNames(specPath, cacheBinArchDir string) (cacheSpecDir, specBase string, err error) {
	specDir, specBase := filepath.Split(specPath)
	if specBase[0] == '.' {
		return "", "", fmt.Errorf("refusing to run hidden spec file: %s", specPath)
	}

	specAbsDir, err := filepath.Abs(specDir)
	if err != nil {
		return "", "", fmt.Errorf("determining absolute path: %v", err)
	}
	specDirSafe := escapeFilename(specAbsDir)
	cacheSpecDir = filepath.Join(cacheBinArchDir, specDirSafe)
	return cacheSpecDir, specBase, nil
}

func doit(args []string) error {
	specPath := args[0]

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

	// TODO combine this state tracking into a single struct, so there's less places to screw up

	// the current spec path, as we follow symlinks
	realSpecPath := specPath
	cacheSpecDir, specBase, err := cacheNames(realSpecPath, cacheBinArchDir)
	if err != nil {
		return err
	}

	// we always exec the original cachefile, to get argv[0] right
	binary := filepath.Join(cacheSpecDir, specBase)
	binArgs := make([]string, len(args))
	binArgs[0] = binary
	copy(binArgs[1:], args[1:])

	for symlinkIter := 1; ; symlinkIter++ {
		const maxSymlinkIter = 100
		if symlinkIter > maxSymlinkIter {
			return fmt.Errorf("too many levels of symbolic links: %s", specPath)
		}

		if !*onlyBuild && !*upgrade {
			err = runBinary(binary, binArgs, os.Environ())
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("cannot exec %s: %v", binary, err)
			}
		}

		// if we're still here, we don't have a cached binary, or we're
		// upgrading

		// make sure cache directory exists
		if err := maybeMkdirs(0750, cacheDir, cacheBinDir, cacheBinArchDir); err != nil {
			return fmt.Errorf("cannot create cache directory: %v", err)
		}

		target, err := maybeReadlink(realSpecPath)
		if err != nil {
			return err
		}
		if target == "" {
			// not a symlink
			break
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(realSpecPath), target)
		}

		// drop in a breadcrumb that can be used the next time around,
		// mirroring the specfile symlink
		cacheSpecDir, specBase, err = cacheNames(target, cacheBinArchDir)
		if err != nil {
			return err
		}
		p := filepath.Join("..", filepath.Base(cacheSpecDir), specBase)
		symCacheSpecDir, symSpecBase, err := cacheNames(realSpecPath, cacheBinArchDir)
		if err != nil {
			return err
		}
		if err := maybeMkdirs(0750, symCacheSpecDir); err != nil {
			return fmt.Errorf("cannot create cache directory: %v", err)
		}
		if err := symlink(p, filepath.Join(symCacheSpecDir, symSpecBase)); err != nil {
			return err
		}

		// finally, follow the symlink
		realSpecPath = target
	}

	// not a symlink
	err = build(specPath, specFile, cacheSpecDir, specBase)
	if err != nil {
		return err
	}

	if !*onlyBuild {
		// now run it (again); this time ENOENT means trouble
		err = runBinary(binary, binArgs, os.Environ())
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
