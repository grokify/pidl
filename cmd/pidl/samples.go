package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

<file> is a PIDL document whose steps carry samples, or a sample set
(format pidl-samples/v1): samples an implementation keeps apart from the
document, each attached to a step by its address.

Resolve reads every sample source (a marked region of a file in a
repository) and records the code and its SHA-256: for a document, in a
lock file beside it; for a sample set, in the set itself. Check fails when
a source is not resolved, its region is missing, or its code changed since
it was resolved.

Options:
  --root <dir>       Directory holding repository checkouts at <root>/<repo>
                     (default: $GOPATH/src or ~/go/src)
  --lock <file>      Document lock file (default: <file without .json>.samples.json)
  --repo-dir <dir>   Read every source from this checkout instead of
                     <root>/<repo>, such as the repository a sample set
                     lives in (use . in its own CI)
  --protocol <file>  With a sample set: also check that every step address
                     names a step of this document
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
	protocolPath := fs.String("protocol", "", "Document a sample set's addresses must name steps of")
	repoDir := fs.String("repo-dir", "", "Read every source from this checkout")
	fs.Usage = func() { usage(stderr) }
	if err := fs.Parse(args[1:]); err != nil {
		return flagExitCode(err)
	}
	if fs.NArg() != 1 {
		usage(stderr)
		return 1
	}
	filename := filepath.Clean(fs.Arg(0))
	data, err := os.ReadFile(filename) //nolint:gosec // G304, G703: the path is a command-line argument
	if err != nil {
		fmt.Fprintf(stderr, "Error reading %s: %v\n", filename, err)
		return 1
	}
	read := pidl.LocalSourceReader(*root)
	if *repoDir != "" {
		read = pidl.CheckoutSourceReader(*repoDir)
	}
	if pidl.IsSampleSet(data) {
		return samplesSet(sub, filename, data, *protocolPath, read, stdout, stderr)
	}
	if *protocolPath != "" {
		fmt.Fprintln(stderr, "--protocol applies only to a sample set")
		return 1
	}
	p, err := pidl.Parse(data)
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

// samplesSet resolves or checks a sample set.
func samplesSet(sub, filename string, data []byte, protocolPath string, read pidl.SourceReader, stdout, stderr io.Writer) int {
	set, err := pidl.ParseSampleSet(data)
	if err != nil {
		fmt.Fprintf(stderr, "Error parsing %s: %v\n", filename, err)
		return 1
	}
	if errs := set.Validate(); errs.HasErrors() {
		fmt.Fprintf(stderr, "Validation errors in %s:\n%s", filename, errs)
		return 1
	}
	if sub == "resolve" {
		if err := set.Resolve(read); err != nil {
			fmt.Fprintf(stderr, "Error resolving samples:\n%v\n", err)
			return 1
		}
		if err := set.WriteFile(filename); err != nil {
			fmt.Fprintf(stderr, "Error writing %s: %v\n", filename, err)
			return 1
		}
		fmt.Fprintf(stdout, "Resolved %d samples in %s\n", len(set.Samples), filename)
		return 0
	}

	problems := set.Check(read)
	if protocolPath != "" {
		p, err := pidl.ParseFile(protocolPath)
		if err != nil {
			fmt.Fprintf(stderr, "Error parsing %s: %v\n", protocolPath, err)
			return 1
		}
		if _, err := p.WithSampleSet(set); err != nil {
			problems = append(problems, strings.Split(err.Error(), "\n")...)
		}
	}
	for _, pr := range problems {
		fmt.Fprintln(stderr, pr)
	}
	if len(problems) > 0 {
		return 1
	}
	fmt.Fprintf(stdout, "%d samples in %s are current\n", len(set.Samples), filename)
	return 0
}

// applySamples joins samples to a document for generation. Each file is a
// lock for the document's own sample sources or a sample set. With no
// files, the document's conventional lock applies when it exists.
func applySamples(p *pidl.Protocol, docPath string, files []string) (*pidl.Protocol, error) {
	if len(files) == 0 {
		lock, err := pidl.ReadSampleLock(pidl.SampleLockPath(docPath))
		switch {
		case err == nil:
			return p.WithSamples(lock), nil
		case os.IsNotExist(err):
			return p, nil
		default:
			return nil, err
		}
	}
	for _, path := range files {
		data, err := os.ReadFile(filepath.Clean(path)) //nolint:gosec // G304, G703: the path is a command-line argument
		if err != nil {
			return nil, err
		}
		if pidl.IsSampleSet(data) {
			set, err := pidl.ParseSampleSet(data)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			if p, err = p.WithSampleSet(set); err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			continue
		}
		lock, err := pidl.ParseSampleLock(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		p = p.WithSamples(lock)
	}
	return p, nil
}
