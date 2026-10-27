package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/observe"
	"github.com/grokify/pidl/observe/adapter/sandbox"
	"github.com/grokify/pidl/render"
)

// observeFlags are the generate options that add a run observed in a real
// system to a walkthrough.
type observeFlags struct {
	observations repeatedFlag
	binding      *string
	scenarios    *string
	from, to     *string
	raw          *bool
}

func addObserveFlags(fs *flag.FlagSet) *observeFlags {
	f := &observeFlags{}
	fs.Var(&f.observations, "observation", "walkthrough-json and viewer: an observation (pidl-observation/v1) of a real run to show beside the steps; make one with pidl observe import")
	f.binding = fs.String("binding", "", "With --observation: a binding (pidl-binding/v1) describing the deployment the run came from")
	f.scenarios = fs.String("scenarios", "", "With --observation: scenario IDs to look for, in order, comma-separated (default: every scenario that is not a fragment)")
	f.from = fs.String("from", "", "With --observation: match only exchanges at or after this RFC 3339 time")
	f.to = fs.String("to", "", "With --observation: match only exchanges at or before this RFC 3339 time")
	f.raw = fs.Bool("raw", false, "With --observation: keep raw token, signature and credential values instead of masking them")
	return f
}

// set reports whether an observation was given.
func (f *observeFlags) set() bool { return len(f.observations) > 0 }

// transform loads the observation and binding and returns the walkthrough
// transform that adds the observed run, reporting what it found to stderr.
func (f *observeFlags) transform(p *pidl.Protocol, stderr io.Writer) (func(*pidl.Walkthrough) error, error) {
	if len(f.observations) > 1 {
		return nil, errors.New("one --observation at a time; joining a client's observation to a server's is not supported yet")
	}
	path := f.observations[0]
	data, err := os.ReadFile(filepath.Clean(path)) //nolint:gosec // G304, G703: the path is a command-line argument
	if err != nil {
		return nil, err
	}
	if !pidl.IsObservation(data) {
		return nil, fmt.Errorf("%s is not an observation (pidl-observation/v1); convert a capture with pidl observe import", path)
	}
	o, err := pidl.ParseObservation(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if errs := o.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("%s: %w", path, errs)
	}
	var b *pidl.Binding
	if *f.binding != "" {
		if b, err = pidl.ReadBinding(*f.binding); err != nil {
			return nil, err
		}
		if errs := b.Validate(); len(errs) > 0 {
			return nil, fmt.Errorf("%s: %w", *f.binding, errs)
		}
	}
	opts := observe.Options{Raw: *f.raw}
	if *f.scenarios != "" {
		for _, id := range strings.Split(*f.scenarios, ",") {
			if id = strings.TrimSpace(id); id != "" {
				opts.Scenarios = append(opts.Scenarios, id)
			}
		}
	}
	if opts.From, err = parseTime("--from", *f.from); err != nil {
		return nil, err
	}
	if opts.To, err = parseTime("--to", *f.to); err != nil {
		return nil, err
	}
	return func(w *pidl.Walkthrough) error {
		report, err := observe.Apply(p, w, o, b, opts)
		if err != nil {
			return err
		}
		for _, wn := range report.Warnings {
			fmt.Fprintf(stderr, "warning: %s\n", wn)
		}
		var names []string
		for _, inst := range report.Instances {
			names = append(names, inst.Scenario)
		}
		fmt.Fprintf(stderr, "observed: %d scenario instance(s) [%s]; %d exchange(s) outside any instance, %d outside the time window\n",
			len(report.Instances), strings.Join(names, ", "), len(report.Outside), report.LeftOut)
		if opts.Raw {
			fmt.Fprintln(stderr, "warning: --raw keeps raw tokens and signatures; the output is as sensitive as the traffic it came from")
		}
		return nil
	}, nil
}

func parseTime(flagName, s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", flagName, err)
	}
	return &t, nil
}

// renderObserved renders a walkthrough or viewer page with an observed run.
func renderObserved(format render.Format, p *pidl.Protocol, transform func(*pidl.Walkthrough) error) (string, error) {
	w := render.NewWalkthroughJSON()
	w.Transform = transform
	if format == render.FormatViewer {
		v := render.NewViewer()
		v.Walkthrough = w
		return v.RenderString(p)
	}
	return w.RenderString(p)
}

// observeImporters read a capture as an observation, by --from.
var observeImporters = map[string]func(path string) (*pidl.Observation, error){
	"sandbox": sandbox.ReadFile,
}

func cmdObserve(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		printObserveUsage(stdout)
		return 0
	}
	switch args[0] {
	case "import":
		return cmdObserveImport(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "Unknown observe command: %s\n\n", args[0])
		printObserveUsage(stderr)
		return 1
	}
}

func printObserveUsage(w io.Writer) {
	fmt.Fprint(w, `Usage: pidl observe <command> [options]

Work with runs observed in real systems.

Commands:
  import   Convert a capture into an observation (pidl-observation/v1)

Show an observation beside a document's steps with
  pidl generate -f viewer --observation run.json --binding deployment.json <document>
`)
}

func cmdObserveImport(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("observe import", flag.ContinueOnError)
	fs.SetOutput(stderr)
	from := fs.String("from", "", "Capture format: sandbox (the aauth-go sandbox's trace)")
	output := fs.String("o", "", "Output file (default: stdout)")
	fs.Usage = func() {
		fmt.Fprint(stdout, "Usage: pidl observe import --from <format> [-o <file>] <capture>\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return flagExitCode(err)
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return 2
	}
	read, ok := observeImporters[*from]
	if !ok {
		fmt.Fprintf(stderr, "Error: --from must be one of: sandbox\n")
		return 2
	}
	o, err := read(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	if errs := o.Validate(); len(errs) > 0 {
		fmt.Fprintf(stderr, "Error: the observation does not validate: %v\n", errs)
		return 1
	}
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	data = append(data, '\n')
	if *output == "" {
		if _, err := stdout.Write(data); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	}
	if err := os.WriteFile(*output, data, 0o600); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Fprintf(stderr, "Wrote %s: %d exchange(s)\n", *output, len(o.Exchanges))
	return 0
}
