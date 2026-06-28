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

const version = "0.9.0"

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
	outputJSON := fs.Bool("json", false, "Output trace as JSON")
	verbose := fs.Bool("v", false, "Verbose output (show each step)")
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
  pidl simulate oauth2_with_states.json         Run full simulation
  pidl simulate -steps=5 protocol.json          Run first 5 steps
  pidl simulate -v protocol.json                Show each step
  pidl simulate -json protocol.json             Output trace as JSON
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

	if *outputJSON {
		jsonBytes, err := json.MarshalIndent(trace, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling trace: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonBytes))
	} else if !*verbose {
		printTraceSummary(trace)
	} else {
		// Verbose already printed steps, just print summary
		fmt.Println()
		printTraceSummary(trace)
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
