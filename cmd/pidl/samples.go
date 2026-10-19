package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/grokify/pidl"
)

// defaultSamplesRoot is where source repositories are checked out:
// $GOPATH/src, or ~/go/src.
func defaultSamplesRoot() string {
	if gp := os.Getenv("GOPATH"); gp != "" {
		return filepath.Join(filepath.SplitList(gp)[0], "src")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, "go", "src")
}

func cmdSamples(args []string, stdout, stderr io.Writer) int {
	usage := func(w io.Writer) {
		fmt.Fprint(w, `Usage: pidl samples <resolve|check> [options] <file>

Resolve reads every sample source (a marked region of a file in a
repository) and writes the code and its SHA-256 to a lock file beside the
document. Check fails when a source is not resolved, its region is
missing, or its code changed since it was resolved.

Options:
  --root <dir>   Directory holding repository checkouts at <root>/<repo>
                 (default: $GOPATH/src or ~/go/src)
  --lock <file>  Lock file (default: <file without .json>.samples.json)
`)
	}
	if len(args) == 0 {
		usage(stderr)
		return 1
	}
	sub := args[0]
	if sub == "-h" || sub == "--help" || sub == "help" {
		usage(stdout)
		return 0
	}
	if sub != "resolve" && sub != "check" {
		fmt.Fprintf(stderr, "Unknown samples command: %s\n\n", sub)
		usage(stderr)
		return 1
	}

	fs := flag.NewFlagSet("samples "+sub, flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", defaultSamplesRoot(), "Directory holding repository checkouts")
	lockPath := fs.String("lock", "", "Lock file")
	fs.Usage = func() { usage(stderr) }
	if err := fs.Parse(args[1:]); err != nil {
		return flagExitCode(err)
	}
	if fs.NArg() != 1 {
		usage(stderr)
		return 1
	}
	filename := fs.Arg(0)
	p, err := pidl.ParseFile(filename)
	if err != nil {
		fmt.Fprintf(stderr, "Error parsing %s: %v\n", filename, err)
		return 1
	}
	if errs := p.Validate(); errs.HasErrors() {
		fmt.Fprintf(stderr, "Validation errors in %s:\n%s", filename, errs)
		return 1
	}
	if *lockPath == "" {
		*lockPath = pidl.SampleLockPath(filename)
	}
	read := pidl.LocalSourceReader(*root)

	if sub == "resolve" {
		lock, err := p.ResolveSamples(read)
		if err != nil {
			fmt.Fprintf(stderr, "Error resolving samples:\n%v\n", err)
			return 1
		}
		if err := lock.WriteFile(*lockPath); err != nil {
			fmt.Fprintf(stderr, "Error writing %s: %v\n", *lockPath, err)
			return 1
		}
		fmt.Fprintf(stdout, "Resolved %d sample sources into %s\n", len(lock.Samples), *lockPath)
		return 0
	}

	lock, err := pidl.ReadSampleLock(*lockPath)
	if err != nil {
		if os.IsNotExist(err) && len(p.SampleSources()) == 0 {
			fmt.Fprintln(stdout, "No sample sources to check")
			return 0
		}
		fmt.Fprintf(stderr, "Error reading %s: %v\n", *lockPath, err)
		return 1
	}
	problems := p.CheckSamples(lock, read)
	for _, pr := range problems {
		fmt.Fprintln(stderr, pr)
	}
	if len(problems) > 0 {
		return 1
	}
	fmt.Fprintf(stdout, "%d sample sources match %s\n", len(lock.Samples), *lockPath)
	return 0
}
