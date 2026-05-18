// Command pidl is the CLI tool for the Protocol Interaction Description Language.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

const version = "0.1.0"

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
	case "examples", "list-examples":
		cmdExamples(os.Args[2:])
	case "init":
		cmdInit(os.Args[2:])
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
  validate   Validate PIDL JSON files
  generate   Generate diagrams from PIDL files
  examples   List or show built-in examples
  init       Create a new PIDL file from template
  version    Print version information
  help       Show this help message

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
	formatStr := fs.String("f", "plantuml", "Output format: plantuml, mermaid, dot, d2, d2-flow, d2-arch, svg, svg-animated, svg-network")
	output := fs.String("o", "", "Output file (default: stdout)")
	template := fs.String("template", "", "SVG template name (default, minimal, sketch, blueprint, dark)")
	templateDir := fs.String("template-dir", "", "Path to custom SVG template directory")
	theme := fs.String("theme", "", "SVG theme (light, dark, auto)")
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
  plantuml, puml   PlantUML sequence diagram
  mermaid, mmd     Mermaid sequence diagram
  dot, graphviz    Graphviz DOT data flow diagram
  d2               D2 sequence diagram
  d2-flow          D2 data flow diagram
  d2-arch          D2 architecture diagram
  svg              SVG sequence diagram
  svg-animated     Animated SVG with flow dots
  svg-network      Network boundary diagram

SVG Templates:
  default          Clean, professional styling
  minimal          Ultra-clean, reduced chrome
  sketch           Hand-drawn, informal look
  blueprint        Technical with monospace fonts
  dark             Dark background optimized

Network Boundary Examples:
  --boundary="dmz:auth,gateway"      Assign entities to DMZ boundary
  --boundary="internal:api,db"       Assign entities to internal boundary
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

	if errs := p.Validate(); errs.HasErrors() {
		fmt.Fprintf(os.Stderr, "Validation errors in %s:\n%s", filename, errs)
		os.Exit(1)
	}

	// Handle SVG-specific rendering with templates and options
	var diagram string
	if format == render.FormatSVG || format == render.FormatSVGAnimated {
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

		diagram, err = renderer.RenderString(p)
	} else if format == render.FormatSVGNetwork {
		renderer := render.NewSVGNetwork()
		if *theme != "" {
			renderer.Theme = *theme
		}

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

	if err := pidl.WriteProtocolFile(filename, p); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created %s\n", filename)
	fmt.Printf("Protocol: %s\n", p.ProtocolMeta.Name)
	fmt.Printf("ID: %s\n", p.ProtocolMeta.ID)
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
