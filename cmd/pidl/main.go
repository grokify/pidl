// Command pidl is the CLI tool for the Protocol Interaction Description Language.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/analyze"
	"github.com/grokify/pidl/examples"
	"github.com/grokify/pidl/render"
)

// boundaryFlags is a custom flag type that collects multiple --boundary values.
type boundaryFlags []string

func (b *boundaryFlags) String() string {
	return strings.Join(*b, ", ")
}

func (b *boundaryFlags) Set(value string) error {
	*b = append(*b, value)
	return nil
}

const version = "0.4.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "validate":
		cmdValidate(os.Args[2:])
	case "generate", "gen":
		cmdGenerate(os.Args[2:])
	case "resolve":
		cmdResolve(os.Args[2:])
	case "simulate", "sim", "run":
		cmdSimulate(os.Args[2:])
	case "examples", "list-examples":
		cmdExamples(os.Args[2:])
	case "init":
		cmdInit(os.Args[2:])
	case "roles":
		cmdRoles(os.Args[2:])
	case "components":
		cmdComponents(os.Args[2:])
	case "trust":
		cmdTrust(os.Args[2:])
	case "diff":
		cmdDiff(os.Args[2:])
	case "debug":
		cmdDebug(os.Args[2:])
	case "analyze":
		cmdAnalyze(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("pidl version %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`pidl - Protocol Interaction Description Language CLI

Usage:
  pidl <command> [options] [arguments]

Commands:
  validate     Validate PIDL JSON files
  generate     Generate diagrams from PIDL files
  resolve      Resolve imports and extends, output merged protocol
  simulate     Simulate protocol execution and show trace
  diff         Compare two protocol files
  debug        Interactive protocol debugger
  analyze      Security analysis of protocol
  roles        List protocol roles from PIDL files
  components   List deployment components from PIDL files
  trust        List trust relationships from PIDL files
  examples     List or show built-in examples
  init         Create a new PIDL file from template
  version      Print version information
  help         Show this help message

Run 'pidl <command> -h' for more information on a command.
`)
}

func cmdValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	quiet := fs.Bool("q", false, "Quiet mode (only show errors)")
	fs.Usage = func() {
		fmt.Print(`Usage: pidl validate [options] <file> [file...]

Validate PIDL JSON files against the schema.

Options:
`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() == 0 {
		fs.Usage()
		os.Exit(1)
	}

	files := fs.Args()
	results := pidl.ValidateFiles(files)

	hasErrors := false
	for _, r := range results {
		if r.ParseErr != nil {
			fmt.Fprintf(os.Stderr, "%s: parse error: %v\n", r.Filename, r.ParseErr)
			hasErrors = true
			continue
		}

		if r.Errors.HasErrors() {
			fmt.Fprintf(os.Stderr, "%s: validation failed\n", r.Filename)
			for _, e := range r.Errors {
				fmt.Fprintf(os.Stderr, "  - %s\n", e)
			}
			hasErrors = true
			continue
		}

		if !*quiet {
			fmt.Printf("%s: valid (%s)\n", r.Filename, r.Protocol.ProtocolMeta.Name)
		}
	}

	if hasErrors {
		os.Exit(1)
	}
}

func cmdGenerate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	formatStr := fs.String("f", "plantuml", "Output format: plantuml, mermaid, mermaid-state, mermaid-component, mermaid-trust, markdown-matrix, dot, d2, d2-flow, d2-arch, svg, svg-animated, svg-network, svg-component, svg-trust")
	output := fs.String("o", "", "Output file (default: stdout)")
	template := fs.String("template", "", "SVG template name (default, minimal, sketch, blueprint, dark)")
	templateDir := fs.String("template-dir", "", "Path to custom SVG template directory")
	theme := fs.String("theme", "", "SVG theme (light, dark, auto)")
	entity := fs.String("entity", "", "Entity ID to filter (for mermaid-state format)")
	showSecurity := fs.Bool("show-security", true, "Show security annotations on flows")
	showTrust := fs.Bool("show-trust", true, "Show trust levels and infer boundaries from trust")
	resolveImports := fs.Bool("resolve", false, "Resolve imports and extends before generating")
	var boundaries boundaryFlags
	fs.Var(&boundaries, "boundary", "Network boundary assignment (format: boundary_id:entity1,entity2). Can be repeated.")
	fs.Usage = func() {
		fmt.Print(`Usage: pidl generate [options] <file>

Generate diagram output from a PIDL file.

Options:
`)
		fs.PrintDefaults()
		fmt.Print(`
Formats:
  plantuml, puml     PlantUML sequence diagram
  mermaid, mmd       Mermaid sequence diagram
  mermaid-state      Mermaid state diagram (requires entity states)
  mermaid-component  Mermaid deployment component diagram
  mermaid-trust      Mermaid trust relationship diagram
  markdown-matrix    Markdown protocol role matrix
  dot, graphviz      Graphviz DOT data flow diagram
  d2                 D2 sequence diagram
  d2-flow            D2 data flow diagram
  d2-arch            D2 architecture diagram
  svg                SVG sequence diagram
  svg-animated       Animated SVG with flow dots
  svg-network        Network boundary diagram
  svg-component      SVG deployment component diagram
  svg-trust          SVG trust relationship diagram

SVG Templates:
  default          Clean, professional styling
  minimal          Ultra-clean, reduced chrome
  sketch           Hand-drawn, informal look
  blueprint        Technical with monospace fonts
  dark             Dark background optimized

Network Boundary Examples:
  --boundary="dmz:auth,gateway"      Assign entities to DMZ boundary
  --boundary="internal:api,db"       Assign entities to internal boundary

State Diagram Examples:
  pidl generate -f mermaid-state example.json              All entities with states
  pidl generate -f mermaid-state --entity=client example.json   Single entity
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() == 0 {
		fs.Usage()
		os.Exit(1)
	}

	format, err := render.ParseFormat(*formatStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	filename := fs.Arg(0)

	// Check if it's an example name
	var p *pidl.Protocol
	if !strings.Contains(filename, "/") && !strings.Contains(filename, "\\") && !strings.HasSuffix(filename, ".json") {
		p, err = examples.GetProtocol(filename)
		if err != nil {
			// Try as file
			p, err = pidl.ParseFile(filename)
		}
	} else {
		p, err = pidl.ParseFile(filename)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", filename, err)
		os.Exit(1)
	}

	// Resolve imports/extends if requested or needed
	if *resolveImports || p.NeedsResolution() {
		opts := pidl.DefaultResolveOptions()
		opts.BasePath = filepath.Dir(filename)
		p, err = p.Resolve(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving %s: %v\n", filename, err)
			os.Exit(1)
		}
	}

	if errs := p.Validate(); errs.HasErrors() {
		fmt.Fprintf(os.Stderr, "Validation errors in %s:\n%s", filename, errs)
		os.Exit(1)
	}

	// Handle format-specific rendering
	var diagram string
	if format == render.FormatMermaidState {
		renderer := render.NewMermaidState()
		if *entity != "" {
			renderer.EntityFilter = *entity
		}
		diagram, err = renderer.RenderString(p)
	} else if format == render.FormatSVG || format == render.FormatSVGAnimated {
		var renderer *render.SVGRenderer
		if *templateDir != "" {
			renderer, err = render.NewSVGWithTemplateDir(*templateDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading template: %v\n", err)
				os.Exit(1)
			}
		} else if *template != "" {
			renderer, err = render.NewSVGWithTemplate(*template)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading template: %v\n", err)
				os.Exit(1)
			}
		} else {
			renderer = render.NewSVG()
		}

		if format == render.FormatSVGAnimated {
			renderer.Animated = true
		}
		if *theme != "" {
			renderer.Theme = *theme
		}
		renderer.ShowSecurity = *showSecurity

		diagram, err = renderer.RenderString(p)
	} else if format == render.FormatSVGNetwork {
		renderer := render.NewSVGNetwork()
		if *theme != "" {
			renderer.Theme = *theme
		}
		renderer.ShowTrust = *showTrust
		renderer.InferBoundariesFromTrust = *showTrust

		// Parse and apply boundary overrides
		for _, b := range boundaries {
			boundaryID, entityIDs, parseErr := parseBoundaryFlag(b)
			if parseErr != nil {
				fmt.Fprintf(os.Stderr, "Error parsing boundary flag %q: %v\n", b, parseErr)
				os.Exit(1)
			}
			renderer.AddBoundaryOverride(boundaryID, entityIDs)
		}

		diagram, err = renderer.RenderString(p)
	} else if format == render.FormatPlantUML {
		renderer := render.NewPlantUML()
		renderer.ShowSecurity = *showSecurity
		diagram, err = renderer.RenderString(p)
	} else if format == render.FormatMermaid {
		renderer := render.NewMermaid()
		renderer.ShowSecurity = *showSecurity
		diagram, err = renderer.RenderString(p)
	} else {
		diagram, err = render.RenderString(format, p)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering: %v\n", err)
		os.Exit(1)
	}

	if *output == "" {
		fmt.Print(diagram)
	} else {
		if err := os.WriteFile(*output, []byte(diagram), 0600); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", *output, err)
			os.Exit(1)
		}
		fmt.Printf("Wrote %s\n", *output)
	}
}

func cmdResolve(args []string) {
	fs := flag.NewFlagSet("resolve", flag.ExitOnError)
	output := fs.String("o", "", "Output file (default: stdout)")
	validate := fs.Bool("validate", true, "Validate resolved protocol")
	fs.Usage = func() {
		fmt.Print(`Usage: pidl resolve [options] <file>

Resolve imports and extends in a PIDL file, outputting the merged protocol.

This command recursively resolves:
- extends: Merges base protocol entities, phases, and flows
- imports: Selectively imports entities, phases, and flows from other files

Options:
`)
		fs.PrintDefaults()
		fmt.Print(`
Examples:
  pidl resolve composed.json              Resolve and print to stdout
  pidl resolve composed.json -o out.json  Resolve and write to file
  pidl resolve composed.json --validate=false  Skip validation
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() == 0 {
		fs.Usage()
		os.Exit(1)
	}

	filename := fs.Arg(0)
	p, err := pidl.ParseFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", filename, err)
		os.Exit(1)
	}

	if !p.NeedsResolution() {
		fmt.Fprintf(os.Stderr, "Note: %s has no imports or extends to resolve\n", filename)
	}

	opts := pidl.DefaultResolveOptions()
	opts.BasePath = filepath.Dir(filename)
	resolved, err := p.Resolve(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving %s: %v\n", filename, err)
		os.Exit(1)
	}

	if *validate {
		if errs := resolved.Validate(); errs.HasErrors() {
			fmt.Fprintf(os.Stderr, "Validation errors in resolved protocol:\n%s", errs)
			os.Exit(1)
		}
	}

	jsonBytes, err := json.MarshalIndent(resolved, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	if *output == "" || *output == "-" {
		fmt.Println(string(jsonBytes))
	} else {
		if err := os.WriteFile(*output, jsonBytes, 0600); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", *output, err)
			os.Exit(1)
		}
		fmt.Printf("Wrote resolved protocol to %s\n", *output)
	}
}

func cmdSimulate(args []string) {
	fs := flag.NewFlagSet("simulate", flag.ExitOnError)
	steps := fs.Int("steps", 0, "Number of steps to execute (0 = all)")
	outputJSON := fs.Bool("json", false, "Output trace as JSON (shorthand for --trace-format=json)")
	verbose := fs.Bool("v", false, "Verbose output (show each step)")
	traceFormat := fs.String("trace-format", "text", "Trace output format: text, json, svg, mermaid")
	traceOutput := fs.String("trace-output", "", "Output file for trace (default: stdout)")
	showStates := fs.Bool("show-states", true, "Show entity state changes in trace")
	showTimings := fs.Bool("show-timings", false, "Show timing information in trace")
	fs.Usage = func() {
		fmt.Print(`Usage: pidl simulate [options] <file>

Simulate protocol execution and display the execution trace.

The simulator executes flows in order, tracking entity state changes
defined by the 'sets' field on each flow. This helps visualize the
protocol's state machine behavior.

Options:
`)
		fs.PrintDefaults()
		fmt.Print(`
Examples:
  pidl simulate oauth2_with_states.json                Run full simulation
  pidl simulate -steps=5 protocol.json                 Run first 5 steps
  pidl simulate -v protocol.json                       Show each step
  pidl simulate -json protocol.json                    Output trace as JSON
  pidl simulate --trace-format=svg -o trace.svg example.json  Output SVG trace
  pidl simulate --trace-format=mermaid example.json    Output Mermaid sequence
  pidl simulate --show-states --show-timings example.json
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() == 0 {
		fs.Usage()
		os.Exit(1)
	}

	filename := fs.Arg(0)
	p, err := loadProtocol(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Resolve if needed
	if p.NeedsResolution() {
		opts := pidl.DefaultResolveOptions()
		opts.BasePath = filepath.Dir(filename)
		p, err = p.Resolve(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving: %v\n", err)
			os.Exit(1)
		}
	}

	exec := pidl.NewExecutor(p)
	ctx := exec.NewContext()

	var trace *pidl.ExecutionTrace
	if *steps > 0 {
		if *verbose {
			for i := 0; i < *steps && !ctx.Completed; i++ {
				step, err := exec.Step(ctx)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error at step %d: %v\n", i+1, err)
					os.Exit(1)
				}
				if step != nil {
					printStep(step)
				}
			}
			trace = ctx.Trace
		} else {
			trace, err = exec.RunN(ctx, *steps)
		}
	} else {
		if *verbose {
			for !ctx.Completed {
				step, err := exec.Step(ctx)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				if step != nil {
					printStep(step)
				}
			}
			trace = ctx.Trace
		} else {
			trace, err = exec.Run(ctx)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
		os.Exit(1)
	}

	// Determine output format
	format := *traceFormat
	if *outputJSON {
		format = "json"
	}

	// Render trace in requested format
	var result string
	traceRenderer := render.NewTraceRenderer()
	traceRenderer.ShowStates = *showStates
	traceRenderer.ShowTimings = *showTimings

	switch format {
	case "json":
		jsonBytes, err := trace.ToJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling trace: %v\n", err)
			os.Exit(1)
		}
		result = string(jsonBytes)
	case "svg":
		svg, err := traceRenderer.RenderSVG(trace, p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering SVG: %v\n", err)
			os.Exit(1)
		}
		result = svg
	case "mermaid", "mmd":
		result = traceRenderer.RenderMermaid(trace, p)
	default:
		// Text format
		if *verbose {
			// Verbose already printed steps, just print summary
			fmt.Println()
			printTraceSummary(trace)
			return
		}
		opts := render.TraceTextOptions{
			ShowTimestamps: *showTimings,
			ShowStates:     *showStates,
			Compact:        false,
		}
		result = traceRenderer.RenderText(trace, opts)
	}

	// Output result
	if *traceOutput != "" {
		if err := os.WriteFile(*traceOutput, []byte(result), 0600); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", *traceOutput, err)
			os.Exit(1)
		}
		fmt.Printf("Wrote trace to %s\n", *traceOutput)
	} else {
		fmt.Print(result)
	}
}

func printStep(step *pidl.ExecutionStep) {
	status := "✓"
	if step.Skipped {
		status = "○"
	}

	label := step.Action
	if step.Label != "" {
		label = step.Label
	}

	fmt.Printf("%s Step %d: %s -> %s: %s", status, step.StepNumber, step.From, step.To, label)

	if step.Skipped {
		fmt.Printf(" (skipped: %s)", step.SkipReason)
	}

	if len(step.StateChanges) > 0 {
		fmt.Print(" [")
		for i, sc := range step.StateChanges {
			if i > 0 {
				fmt.Print(", ")
			}
			if sc.FromState != "" {
				fmt.Printf("%s: %s→%s", sc.Entity, sc.FromState, sc.ToState)
			} else {
				fmt.Printf("%s: →%s", sc.Entity, sc.ToState)
			}
		}
		fmt.Print("]")
	}

	fmt.Println()
}

func printTraceSummary(trace *pidl.ExecutionTrace) {
	fmt.Printf("Protocol: %s\n", trace.ProtocolName)
	fmt.Printf("Steps executed: %d\n", trace.StepCount())

	if trace.SkippedCount() > 0 {
		fmt.Printf("Steps skipped: %d\n", trace.SkippedCount())
	}

	if trace.StateChangeCount() > 0 {
		fmt.Printf("State changes: %d\n", trace.StateChangeCount())
	}

	fmt.Printf("Duration: %v\n", trace.Duration().Round(100*time.Microsecond))

	if trace.Completed {
		fmt.Println("Status: completed")
	} else {
		fmt.Println("Status: partial")
	}

	if len(trace.FinalStates) > 0 {
		fmt.Println("Final states:")
		for entity, state := range trace.FinalStates {
			fmt.Printf("  %s: %s\n", entity, state)
		}
	}

	if trace.Error != "" {
		fmt.Printf("Error: %s\n", trace.Error)
	}
}

func cmdExamples(args []string) {
	fs := flag.NewFlagSet("examples", flag.ExitOnError)
	showJSON := fs.Bool("json", false, "Show example JSON content")
	fs.Usage = func() {
		fmt.Print(`Usage: pidl examples [options] [name]

List built-in example protocols or show a specific example.

Options:
`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() == 0 {
		// List all examples
		names := examples.List()
		fmt.Println("Available examples:")
		for _, name := range names {
			ex, err := examples.Get(name)
			if err != nil {
				fmt.Printf("  %s\n", name)
				continue
			}
			p, err := ex.Protocol()
			if err != nil {
				fmt.Printf("  %s\n", name)
				continue
			}
			fmt.Printf("  %-30s %s\n", name, p.ProtocolMeta.Name)
		}
		fmt.Println("\nUse 'pidl examples <name>' to show details.")
		fmt.Println("Use 'pidl examples <name> -json' to show JSON content.")
		fmt.Println("Use 'pidl generate <name>' to generate diagrams.")
		return
	}

	name := fs.Arg(0)

	if *showJSON {
		data, err := examples.GetJSON(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
		return
	}

	ex, err := examples.Get(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	p, err := ex.Protocol()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Example: %s\n", ex.Name)
	fmt.Printf("Protocol: %s\n", p.ProtocolMeta.Name)
	fmt.Printf("ID: %s\n", p.ProtocolMeta.ID)
	if p.ProtocolMeta.Description != "" {
		fmt.Printf("Description: %s\n", p.ProtocolMeta.Description)
	}
	fmt.Printf("Category: %s\n", p.ProtocolMeta.Category)
	fmt.Printf("Entities: %d\n", len(p.Entities))
	fmt.Printf("Phases: %d\n", len(p.Phases))
	fmt.Printf("Flows: %d\n", len(p.Flows))

	if len(p.ProtocolMeta.References) > 0 {
		fmt.Println("References:")
		for _, ref := range p.ProtocolMeta.References {
			fmt.Printf("  - %s: %s\n", ref.Name, ref.URL)
		}
	}
}

func cmdInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	name := fs.String("name", "", "Protocol name")
	from := fs.String("from", "", "Initialize from example (e.g., oauth2_authorization_code)")
	fs.Usage = func() {
		fmt.Print(`Usage: pidl init [options] <filename>

Create a new PIDL file from a template or example.

Options:
`)
		fs.PrintDefaults()
		fmt.Print(`
Examples:
  pidl init my-protocol.json
  pidl init -name "My Protocol" my-protocol.json
  pidl init -from oauth2_authorization_code my-oauth.json
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() == 0 {
		fs.Usage()
		os.Exit(1)
	}

	filename := fs.Arg(0)

	// Check if file exists
	if _, err := os.Stat(filename); err == nil {
		fmt.Fprintf(os.Stderr, "Error: file already exists: %s\n", filename)
		os.Exit(1)
	}

	var p *pidl.Protocol

	if *from != "" {
		// Copy from example
		ex, err := examples.Get(*from)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		p, err = ex.Protocol()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		// Deep copy by serializing and deserializing
		data, _ := p.ToJSON()
		p, _ = pidl.Parse(data)
	} else {
		// Create minimal protocol
		id := pidl.SanitizeID(strings.TrimSuffix(filepath.Base(filename), ".json"))
		protocolName := *name
		if protocolName == "" {
			protocolName = strings.ReplaceAll(id, "_", " ")
			protocolName = strings.ReplaceAll(protocolName, "-", " ")
			protocolName = pidl.TitleCase(protocolName)
		}
		p = pidl.NewMinimalProtocol(id, protocolName)
	}

	// Override name if provided
	if *name != "" {
		p.ProtocolMeta.Name = *name
	}

	if err := p.WriteFile(filename); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created %s\n", filename)
	fmt.Printf("Protocol: %s\n", p.ProtocolMeta.Name)
	fmt.Printf("ID: %s\n", p.ProtocolMeta.ID)
}

func cmdRoles(args []string) {
	fs := flag.NewFlagSet("roles", flag.ExitOnError)
	formatStr := fs.String("f", "table", "Output format: table, json")
	protocol := fs.String("protocol", "", "Filter by protocol (e.g., oauth, scim, mcp)")
	fs.Usage = func() {
		fmt.Print(`Usage: pidl roles [options] <file>

List protocol roles defined in a PIDL file.

Options:
`)
		fs.PrintDefaults()
		fmt.Print(`
Examples:
  pidl roles example.json
  pidl roles -protocol=oauth example.json
  pidl roles -f json example.json
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() == 0 {
		fs.Usage()
		os.Exit(1)
	}

	p, err := loadProtocol(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	entities := p.EntitiesWithProtocolRoles()
	if *protocol != "" {
		entities = p.EntitiesByProtocol(*protocol)
	}

	if *formatStr == "json" {
		type roleEntry struct {
			EntityID   string `json:"entity_id"`
			EntityName string `json:"entity_name"`
			Protocol   string `json:"protocol"`
			Role       string `json:"role"`
			Variant    string `json:"variant,omitempty"`
		}
		var entries []roleEntry
		for _, e := range entities {
			for _, r := range e.ProtocolRoles {
				if *protocol == "" || r.Protocol == *protocol {
					entries = append(entries, roleEntry{
						EntityID:   e.ID,
						EntityName: e.Name,
						Protocol:   r.Protocol,
						Role:       r.Role,
						Variant:    r.Variant,
					})
				}
			}
		}
		data, _ := jsonMarshalIndent(entries)
		fmt.Println(string(data))
		return
	}

	// Table format
	fmt.Printf("Protocol Roles in %s\n\n", p.ProtocolMeta.Name)
	fmt.Printf("%-20s %-25s %-12s %-20s %s\n", "ENTITY ID", "ENTITY NAME", "PROTOCOL", "ROLE", "VARIANT")
	fmt.Println(strings.Repeat("-", 90))
	for _, e := range entities {
		for _, r := range e.ProtocolRoles {
			if *protocol == "" || r.Protocol == *protocol {
				fmt.Printf("%-20s %-25s %-12s %-20s %s\n", e.ID, truncate(e.Name, 25), r.Protocol, r.Role, r.Variant)
			}
		}
	}
}

func cmdComponents(args []string) {
	fs := flag.NewFlagSet("components", flag.ExitOnError)
	formatStr := fs.String("f", "table", "Output format: table, json")
	typeFilter := fs.String("type", "", "Filter by component type (e.g., idp, gateway, pdp)")
	fs.Usage = func() {
		fmt.Print(`Usage: pidl components [options] <file>

List deployment components defined in a PIDL file.

Options:
`)
		fs.PrintDefaults()
		fmt.Print(`
Examples:
  pidl components example.json
  pidl components -type=idp example.json
  pidl components -f json example.json
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() == 0 {
		fs.Usage()
		os.Exit(1)
	}

	p, err := loadProtocol(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	components := p.Metadata.Components
	if *typeFilter != "" {
		components = p.ComponentsByType(*typeFilter)
	}

	if *formatStr == "json" {
		data, _ := jsonMarshalIndent(components)
		fmt.Println(string(data))
		return
	}

	// Table format
	fmt.Printf("Deployment Components in %s\n\n", p.ProtocolMeta.Name)
	fmt.Printf("%-20s %-30s %-15s %-30s\n", "ID", "NAME", "TYPE", "ENTITIES")
	fmt.Println(strings.Repeat("-", 100))
	for _, c := range components {
		entities := strings.Join(c.Entities, ", ")
		fmt.Printf("%-20s %-30s %-15s %-30s\n", c.ID, truncate(c.Name, 30), c.Type, truncate(entities, 30))
	}

	if len(components) > 0 {
		fmt.Printf("\nTotal: %d components\n", len(components))
	} else {
		fmt.Println("No components defined.")
	}
}

func cmdTrust(args []string) {
	fs := flag.NewFlagSet("trust", flag.ExitOnError)
	formatStr := fs.String("f", "table", "Output format: table, json")
	typeFilter := fs.String("type", "", "Filter by relationship type (e.g., authenticates, delegates, issues)")
	fs.Usage = func() {
		fmt.Print(`Usage: pidl trust [options] <file>

List trust relationships defined in a PIDL file.

Options:
`)
		fs.PrintDefaults()
		fmt.Print(`
Examples:
  pidl trust example.json
  pidl trust -type=authenticates example.json
  pidl trust -f json example.json
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() == 0 {
		fs.Usage()
		os.Exit(1)
	}

	p, err := loadProtocol(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	relations := p.Metadata.TrustRelations
	if *typeFilter != "" {
		relations = p.TrustRelationsByType(*typeFilter)
	}

	if *formatStr == "json" {
		data, _ := jsonMarshalIndent(relations)
		fmt.Println(string(data))
		return
	}

	// Table format
	fmt.Printf("Trust Relationships in %s\n\n", p.ProtocolMeta.Name)
	fmt.Printf("%-15s %-20s %-15s %-20s %-25s\n", "ID", "FROM", "TYPE", "TO", "CREDENTIALS")
	fmt.Println(strings.Repeat("-", 100))
	for _, t := range relations {
		creds := strings.Join(t.Credentials, ", ")
		mutual := ""
		if t.Mutual {
			mutual = " (mutual)"
		}
		fmt.Printf("%-15s %-20s %-15s %-20s %-25s%s\n",
			truncate(t.ID, 15), truncate(t.From, 20), t.Type, truncate(t.To, 20), truncate(creds, 25), mutual)
	}

	if len(relations) > 0 {
		fmt.Printf("\nTotal: %d trust relationships\n", len(relations))
	} else {
		fmt.Println("No trust relationships defined.")
	}
}

func cmdDiff(args []string) {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	formatStr := fs.String("f", "text", "Output format: text, json, markdown")
	output := fs.String("o", "", "Output file (default: stdout)")
	ignoreMetadata := fs.Bool("ignore-metadata", false, "Ignore metadata changes")
	ignoreDescriptions := fs.Bool("ignore-descriptions", false, "Ignore description changes")
	quiet := fs.Bool("q", false, "Quiet mode (summary only)")
	fs.Usage = func() {
		fmt.Print(`Usage: pidl diff [options] <base-file> <new-file>

Compare two PIDL protocol files and show differences.

Options:
`)
		fs.PrintDefaults()
		fmt.Print(`
Examples:
  pidl diff base.json new.json
  pidl diff -f json base.json new.json
  pidl diff -f markdown -o diff.md base.json new.json
  pidl diff --ignore-metadata base.json new.json
  pidl diff -q base.json new.json
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() != 2 {
		fs.Usage()
		os.Exit(1)
	}

	baseFile := fs.Arg(0)
	newFile := fs.Arg(1)

	base, err := loadProtocol(baseFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading base file: %v\n", err)
		os.Exit(1)
	}

	newProto, err := loadProtocol(newFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading new file: %v\n", err)
		os.Exit(1)
	}

	opts := pidl.DefaultDiffOptions()
	opts.IgnoreMetadata = *ignoreMetadata
	opts.IgnoreDescriptions = *ignoreDescriptions

	diff := pidl.Compare(base, newProto, opts)

	var result string
	switch *formatStr {
	case "json":
		jsonBytes, err := diff.ToJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}
		result = string(jsonBytes)
	case "markdown", "md":
		result = diff.ToMarkdown()
	default:
		if *quiet {
			if !diff.HasChanges() {
				result = "No differences found.\n"
			} else {
				result = fmt.Sprintf("Changes: %d (+%d/-%d/~%d)\n",
					diff.Summary.TotalChanges, diff.Summary.Added, diff.Summary.Removed, diff.Summary.Modified)
			}
		} else {
			result = diff.String()
		}
	}

	if *output == "" {
		fmt.Print(result)
	} else {
		if err := os.WriteFile(*output, []byte(result), 0600); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", *output, err)
			os.Exit(1)
		}
		fmt.Printf("Wrote diff to %s\n", *output)
	}
}

func cmdDebug(args []string) {
	fs := flag.NewFlagSet("debug", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Print(`Usage: pidl debug <file>

Interactive protocol debugger. Step through protocol execution,
set breakpoints, and inspect entity states.

Commands:
  step, s              Execute next flow
  continue, c          Run until breakpoint or completion
  break <idx> [cond]   Set breakpoint at flow index
  delete <idx>         Remove breakpoint
  breakpoints, bp      List all breakpoints
  inspect, i           Show current state
  inspect entity <id>  Show entity details
  inspect flow <idx>   Show flow details
  list, l              List all flows with position marker
  set <entity> <state> Set entity state
  reset, r             Restart execution
  trace                Show execution trace
  help, h              Show this help
  quit, q              Exit debugger

`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() == 0 {
		fs.Usage()
		os.Exit(1)
	}

	filename := fs.Arg(0)
	p, err := loadProtocol(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Resolve if needed
	if p.NeedsResolution() {
		opts := pidl.DefaultResolveOptions()
		opts.BasePath = filepath.Dir(filename)
		p, err = p.Resolve(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving: %v\n", err)
			os.Exit(1)
		}
	}

	session := pidl.NewDebugSession(p)

	fmt.Printf("PIDL Debugger: %s\n", p.ProtocolMeta.Name)
	fmt.Printf("Entities: %d, Flows: %d\n", len(p.Entities), len(p.Flows))
	fmt.Println("Type 'help' for commands, 'quit' to exit.")

	runDebugREPL(session)
}

func runDebugREPL(session *pidl.DebugSession) {
	reader := strings.NewReader("")
	var input string

	for {
		// Print prompt
		state := session.Inspect()
		prompt := fmt.Sprintf("(pidl:%d) ", state.FlowIndex)
		if state.IsCompleted {
			prompt = "(pidl:done) "
		}
		fmt.Print(prompt)

		// Read input
		_, err := fmt.Scanln(&input)
		if err != nil {
			// Handle empty input or EOF
			if err.Error() == "unexpected newline" || err.Error() == "EOF" {
				continue
			}
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			continue
		}

		// Handle empty reader
		_ = reader

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToLower(parts[0])
		args := parts[1:]

		switch cmd {
		case "step", "s":
			step, err := session.Step()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			if step == nil {
				fmt.Println("Execution complete.")
			} else {
				printDebugStep(step)
			}

		case "continue", "c":
			step, err := session.Continue()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			state := session.Inspect()
			if state.IsCompleted {
				fmt.Println("Execution complete.")
			} else if state.AtBreakpoint {
				fmt.Printf("Breakpoint hit at flow %d\n", state.FlowIndex)
			}
			if step != nil {
				printDebugStep(step)
			}

		case "break", "b":
			if len(args) < 1 {
				fmt.Println("Usage: break <flow-index> [condition]")
				continue
			}
			var idx int
			if _, err := fmt.Sscanf(args[0], "%d", &idx); err != nil {
				fmt.Printf("Invalid flow index: %s\n", args[0])
				continue
			}
			condition := ""
			if len(args) > 1 {
				condition = strings.Join(args[1:], " ")
			}
			if err := session.SetBreakpoint(idx, condition); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Breakpoint set at flow %d\n", idx)
			}

		case "delete", "d":
			if len(args) < 1 {
				fmt.Println("Usage: delete <flow-index>")
				continue
			}
			var idx int
			if _, err := fmt.Sscanf(args[0], "%d", &idx); err != nil {
				fmt.Printf("Invalid flow index: %s\n", args[0])
				continue
			}
			if err := session.RemoveBreakpoint(idx); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Breakpoint removed at flow %d\n", idx)
			}

		case "breakpoints", "bp":
			bps := session.ListBreakpoints()
			if len(bps) == 0 {
				fmt.Println("No breakpoints set.")
			} else {
				fmt.Println("Breakpoints:")
				for _, bp := range bps {
					status := "enabled"
					if !bp.Enabled {
						status = "disabled"
					}
					cond := ""
					if bp.Condition != "" {
						cond = fmt.Sprintf(" if %s", bp.Condition)
					}
					fmt.Printf("  %d: %s%s (hit %d times)\n", bp.FlowIndex, status, cond, bp.HitCount)
				}
			}

		case "inspect", "i":
			if len(args) == 0 {
				state := session.Inspect()
				fmt.Print(state.String())
			} else if args[0] == "entity" && len(args) > 1 {
				entity, entityState, err := session.InspectEntity(args[1])
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Printf("Entity: %s (%s)\n", entity.Name, entity.ID)
					fmt.Printf("Type: %s\n", entity.Type)
					if entity.TrustLevel != "" {
						fmt.Printf("Trust Level: %s\n", entity.TrustLevel)
					}
					if entityState != "" {
						fmt.Printf("Current State: %s\n", entityState)
					}
					if entity.HasStates() {
						fmt.Println("Available States:")
						for _, s := range entity.States {
							marker := "  "
							if s.ID == entityState {
								marker = "* "
							}
							fmt.Printf("  %s%s", marker, s.ID)
							if s.Initial {
								fmt.Print(" (initial)")
							}
							if s.Final {
								fmt.Print(" (final)")
							}
							fmt.Println()
						}
					}
				}
			} else if args[0] == "flow" && len(args) > 1 {
				var idx int
				if _, err := fmt.Sscanf(args[1], "%d", &idx); err != nil {
					fmt.Printf("Invalid flow index: %s\n", args[1])
					continue
				}
				flow, err := session.InspectFlow(idx)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Printf("Flow %d:\n", idx)
					fmt.Printf("  From: %s\n", flow.From)
					fmt.Printf("  To: %s\n", flow.To)
					fmt.Printf("  Action: %s\n", flow.Action)
					if flow.Label != "" {
						fmt.Printf("  Label: %s\n", flow.Label)
					}
					if flow.Mode != "" {
						fmt.Printf("  Mode: %s\n", flow.Mode)
					}
					if flow.Phase != "" {
						fmt.Printf("  Phase: %s\n", flow.Phase)
					}
					if flow.Condition != "" {
						fmt.Printf("  Condition: %s\n", flow.Condition)
					}
					if len(flow.Sets) > 0 {
						fmt.Println("  State Mutations:")
						for _, m := range flow.Sets {
							if m.From != "" {
								fmt.Printf("    %s: %s -> %s\n", m.Entity, m.From, m.To)
							} else {
								fmt.Printf("    %s: -> %s\n", m.Entity, m.To)
							}
						}
					}
				}
			} else {
				fmt.Println("Usage: inspect | inspect entity <id> | inspect flow <idx>")
			}

		case "list", "l":
			fmt.Print(session.FormatFlowList())

		case "set":
			if len(args) < 2 {
				fmt.Println("Usage: set <entity-id> <state-id>")
				continue
			}
			if err := session.SetEntityState(args[0], args[1]); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Set %s state to %s\n", args[0], args[1])
			}

		case "reset", "r":
			session.Reset()
			fmt.Println("Execution reset to beginning.")

		case "trace", "t":
			trace := session.Trace()
			fmt.Printf("Steps executed: %d\n", trace.StepCount())
			fmt.Printf("Steps skipped: %d\n", trace.SkippedCount())
			fmt.Printf("State changes: %d\n", trace.StateChangeCount())
			for _, step := range trace.Steps {
				status := "+"
				if step.Skipped {
					status = "-"
				}
				fmt.Printf("  %s %d: %s -> %s: %s\n", status, step.StepNumber, step.From, step.To, step.Action)
			}

		case "help", "h", "?":
			fmt.Print(`Commands:
  step, s              Execute next flow
  continue, c          Run until breakpoint or completion
  break <idx> [cond]   Set breakpoint at flow index
  delete <idx>         Remove breakpoint
  breakpoints, bp      List all breakpoints
  inspect, i           Show current state
  inspect entity <id>  Show entity details
  inspect flow <idx>   Show flow details
  list, l              List all flows with position marker
  set <entity> <state> Set entity state
  reset, r             Restart execution
  trace, t             Show execution trace
  help, h              Show this help
  quit, q              Exit debugger
`)

		case "quit", "q", "exit":
			fmt.Println("Goodbye.")
			return

		default:
			fmt.Printf("Unknown command: %s (type 'help' for commands)\n", cmd)
		}
	}
}

func printDebugStep(step *pidl.ExecutionStep) {
	label := step.Action
	if step.Label != "" {
		label = step.Label
	}

	status := "executed"
	if step.Skipped {
		status = fmt.Sprintf("skipped (%s)", step.SkipReason)
	}

	fmt.Printf("Step %d: %s -> %s: %s [%s]\n", step.StepNumber, step.From, step.To, label, status)

	if len(step.StateChanges) > 0 {
		for _, sc := range step.StateChanges {
			if sc.FromState != "" {
				fmt.Printf("  State: %s: %s -> %s\n", sc.Entity, sc.FromState, sc.ToState)
			} else {
				fmt.Printf("  State: %s: -> %s\n", sc.Entity, sc.ToState)
			}
		}
	}
}

func cmdAnalyze(args []string) {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	formatStr := fs.String("f", "text", "Output format: text, json, markdown")
	output := fs.String("o", "", "Output file (default: stdout)")
	minSeverity := fs.String("min-severity", "info", "Minimum severity: critical, high, medium, low, info")
	failOn := fs.String("fail-on", "", "Exit with error if risks at this severity or above found")
	quiet := fs.Bool("q", false, "Quiet mode (summary only)")
	var categories categoryFlags
	fs.Var(&categories, "category", "Filter by category (repeatable): trust_boundary, authentication, data_protection, token_security, communication, configuration")
	fs.Usage = func() {
		fmt.Print(`Usage: pidl analyze [options] <file>

Perform security analysis on a PIDL protocol file.

The analyzer checks for common security issues including:
  - Trust boundary violations
  - Missing encryption for confidential data
  - Unbound bearer tokens
  - Missing authentication
  - Token without audience
  - And more...

Options:
`)
		fs.PrintDefaults()
		fmt.Print(`
Examples:
  pidl analyze protocol.json
  pidl analyze -f json protocol.json
  pidl analyze --min-severity high protocol.json
  pidl analyze --fail-on high protocol.json
  pidl analyze --category authentication protocol.json
  pidl analyze -q protocol.json
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() == 0 {
		fs.Usage()
		os.Exit(1)
	}

	filename := fs.Arg(0)
	p, err := loadProtocol(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Resolve if needed
	if p.NeedsResolution() {
		opts := pidl.DefaultResolveOptions()
		opts.BasePath = filepath.Dir(filename)
		p, err = p.Resolve(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving: %v\n", err)
			os.Exit(1)
		}
	}

	// Build analysis options
	opts := analyze.DefaultAnalysisOptions()

	if *minSeverity != "" {
		sev, err := analyze.ParseSeverity(*minSeverity)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid severity: %v\n", err)
			os.Exit(1)
		}
		opts.MinSeverity = sev
	}

	for _, cat := range categories {
		parsedCat, err := analyze.ParseCategory(cat)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid category: %v\n", err)
			os.Exit(1)
		}
		opts.Categories = append(opts.Categories, parsedCat)
	}

	// Run analysis
	analysis := analyze.Analyze(p, opts)

	// Format output
	var result string
	switch *formatStr {
	case "json":
		jsonBytes, err := analysis.ToJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}
		result = string(jsonBytes)
	case "markdown", "md":
		result = analysis.ToMarkdown()
	default:
		if *quiet {
			if !analysis.HasRisks() {
				result = fmt.Sprintf("No risks found. Score: %d/100\n", analysis.Summary.Score)
			} else {
				result = fmt.Sprintf("Risks: %d (C:%d H:%d M:%d L:%d I:%d) Score: %d/100\n",
					analysis.Summary.TotalRisks,
					analysis.Summary.BySeverity[analyze.SeverityCritical],
					analysis.Summary.BySeverity[analyze.SeverityHigh],
					analysis.Summary.BySeverity[analyze.SeverityMedium],
					analysis.Summary.BySeverity[analyze.SeverityLow],
					analysis.Summary.BySeverity[analyze.SeverityInfo],
					analysis.Summary.Score)
			}
		} else {
			result = analysis.String()
		}
	}

	// Output result
	if *output == "" {
		fmt.Print(result)
	} else {
		if err := os.WriteFile(*output, []byte(result), 0600); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", *output, err)
			os.Exit(1)
		}
		fmt.Printf("Wrote analysis to %s\n", *output)
	}

	// Check fail-on condition
	if *failOn != "" {
		failSeverity, err := analyze.ParseSeverity(*failOn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid fail-on severity: %v\n", err)
			os.Exit(1)
		}
		if analysis.HasRisksAtOrAbove(failSeverity) {
			os.Exit(1)
		}
	}
}

// categoryFlags is a custom flag type that collects multiple --category values.
type categoryFlags []string

func (c *categoryFlags) String() string {
	return strings.Join(*c, ", ")
}

func (c *categoryFlags) Set(value string) error {
	*c = append(*c, value)
	return nil
}

// loadProtocol loads a protocol from a file or example name.
func loadProtocol(filename string) (*pidl.Protocol, error) {
	var p *pidl.Protocol
	var err error

	if !strings.Contains(filename, "/") && !strings.Contains(filename, "\\") && !strings.HasSuffix(filename, ".json") {
		p, err = examples.GetProtocol(filename)
		if err != nil {
			p, err = pidl.ParseFile(filename)
		}
	} else {
		p, err = pidl.ParseFile(filename)
	}

	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filename, err)
	}

	if errs := p.Validate(); errs.HasErrors() {
		return nil, fmt.Errorf("validation errors in %s:\n%s", filename, errs)
	}

	return p, nil
}

// truncate truncates a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// jsonMarshalIndent marshals v to indented JSON.
func jsonMarshalIndent(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// parseBoundaryFlag parses a boundary flag in the format "boundary_id:entity1,entity2".
func parseBoundaryFlag(s string) (boundaryID string, entityIDs []string, err error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("invalid format, expected 'boundary_id:entity1,entity2'")
	}

	boundaryID = strings.TrimSpace(parts[0])
	if boundaryID == "" {
		return "", nil, fmt.Errorf("boundary ID cannot be empty")
	}

	entityList := strings.TrimSpace(parts[1])
	if entityList == "" {
		return "", nil, fmt.Errorf("entity list cannot be empty")
	}

	for _, e := range strings.Split(entityList, ",") {
		e = strings.TrimSpace(e)
		if e != "" {
			entityIDs = append(entityIDs, e)
		}
	}

	if len(entityIDs) == 0 {
		return "", nil, fmt.Errorf("at least one entity ID is required")
	}

	return boundaryID, entityIDs, nil
}
