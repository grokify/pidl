package pidl

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// SampleSetFormat names the sample set format and its major version.
const SampleSetFormat = "pidl-samples/v1"

// SampleSet holds code samples for a protocol, kept apart from the
// protocol document. An implementation of a protocol, such as an SDK,
// keeps its own set beside its tested code, so the protocol document
// stays independent of any language or implementation. Generation joins
// one or more sets to the document.
//
// A set is also its own lock: resolving a set writes each source's code
// and hash into it, so it generates without the source repository at
// hand.
type SampleSet struct {
	// Format names the format and its major version: pidl-samples/v1.
	Format string `json:"format"`

	// Protocol is the ID of the protocol the set is for. When set, the
	// set applies only to that protocol.
	Protocol string `json:"protocol,omitempty"`

	// Name names the implementation, such as "example-sdk-go". It is the
	// source cited by examples the set creates for steps without one.
	Name string `json:"name,omitempty"`

	// Samples are the set's samples, each attached to a step.
	Samples []SampleSetEntry `json:"samples"`
}

// SampleSetEntry is one sample and the step it belongs to.
type SampleSetEntry struct {
	// Step is the address of a step as written: "<scenario>/<step-id>" for
	// a step of a scenario or fragment, "<scenario>" for the scenario's
	// own samples, or "<step-id>" in a document without scenarios.
	Step string `json:"step"`

	Sample

	// SHA256 is the hex SHA-256 of Code, recorded when Source is resolved.
	SHA256 string `json:"sha256,omitempty"`
}

// IsSampleSet reports whether data is a sample set rather than a sample
// lock, by its format.
func IsSampleSet(data []byte) bool {
	var head struct {
		Format string `json:"format"`
	}
	return json.Unmarshal(data, &head) == nil && strings.HasPrefix(head.Format, "pidl-samples/")
}

// ParseSampleSet parses a sample set. It refuses another major version of
// the format.
func ParseSampleSet(data []byte) (*SampleSet, error) {
	var s SampleSet
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	if s.Format != SampleSetFormat {
		return nil, fmt.Errorf("sample set format %q is not %s", s.Format, SampleSetFormat)
	}
	return &s, nil
}

// ReadSampleSet reads a sample set file.
func ReadSampleSet(path string) (*SampleSet, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- the path is chosen by the caller
	if err != nil {
		return nil, err
	}
	s, err := ParseSampleSet(data)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return s, nil
}

// WriteFile writes the set, keeping the order of its samples.
func (s *SampleSet) WriteFile(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

// Validate checks each entry: a step address, a language that is not a
// generated representation, code or a source, and a source path inside
// its repository.
func (s *SampleSet) Validate() ValidationErrors {
	var errs ValidationErrors
	for i, e := range s.Samples {
		field := fmt.Sprintf("samples[%d]", i)
		if e.Step == "" {
			errs = append(errs, ValidationError{Field: field + ".step", Message: "required"})
		}
		errs = append(errs, validateSamples(field, []Sample{e.Sample}, nil)...)
	}
	return errs
}

// Resolve reads each entry's source and records its code and hash. It
// reports every source that cannot be read or extracted.
func (s *SampleSet) Resolve(read SourceReader) error {
	var errs []error
	for i := range s.Samples {
		e := &s.Samples[i]
		if e.Source == nil {
			continue
		}
		code, err := readRegion(read, *e.Source)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", e.Source.Key(), err))
			continue
		}
		e.Code, e.SHA256 = code, codeHash(code)
	}
	return errors.Join(errs...)
}

// Check reports entries whose source is not resolved, whose recorded code
// does not match its hash, or, when read is not nil, whose source region
// changed since it was resolved.
func (s *SampleSet) Check(read SourceReader) []string {
	var problems []string
	for _, e := range s.Samples {
		if e.Source == nil {
			continue
		}
		key := e.Source.Key()
		if e.Code == "" || e.SHA256 == "" {
			problems = append(problems, fmt.Sprintf("%s: not resolved; run pidl samples resolve", key))
			continue
		}
		if codeHash(e.Code) != e.SHA256 {
			problems = append(problems, fmt.Sprintf("%s: recorded code does not match its sha256", key))
		}
		if read == nil {
			continue
		}
		code, err := readRegion(read, *e.Source)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", key, err))
			continue
		}
		if codeHash(code) != e.SHA256 {
			problems = append(problems, fmt.Sprintf("%s: source changed since it was resolved; run pidl samples resolve", key))
		}
	}
	return problems
}

func readRegion(read SourceReader, src SampleSource) (string, error) {
	data, err := read(src.Repo, src.Ref, src.Path)
	if err != nil {
		return "", err
	}
	return ExtractRegion(data, src.Region)
}

// WithSampleSet returns a copy of the protocol with the set's samples
// attached to the steps they address, after the step's own samples. A step
// without an example gets one holding only the samples, citing the set's
// name. It fails when the set is for another protocol, an address names
// no step, or a source is not resolved. The receiver is not modified.
func (p *Protocol) WithSampleSet(set *SampleSet) (*Protocol, error) {
	if set.Protocol != "" && set.Protocol != p.ProtocolMeta.ID {
		return nil, fmt.Errorf("sample set is for protocol %q, not %q", set.Protocol, p.ProtocolMeta.ID)
	}
	cp := *p
	cp.Flows = append([]Flow(nil), p.Flows...)
	cp.Scenarios = make([]Scenario, len(p.Scenarios))
	for i, sc := range p.Scenarios {
		sc.Flows = append([]Flow(nil), sc.Flows...)
		sc.Samples = append([]Sample(nil), sc.Samples...)
		cp.Scenarios[i] = sc
	}

	source := set.Name
	if source == "" {
		source = "code samples"
	}
	var errs []error
	for _, e := range set.Samples {
		if e.Source != nil && e.Code == "" {
			errs = append(errs, fmt.Errorf("%s: %s not resolved; run pidl samples resolve", e.Step, e.Source.Key()))
			continue
		}
		sample := e.Sample
		scenario, flowID, hasFlow := strings.Cut(e.Step, "/")
		if len(cp.Scenarios) == 0 {
			scenario, flowID, hasFlow = "", e.Step, true
		}
		var flows []Flow
		if scenario != "" {
			sc := cp.ScenarioByID(scenario)
			if sc == nil {
				errs = append(errs, fmt.Errorf("%s: no scenario %q", e.Step, scenario))
				continue
			}
			if !hasFlow {
				sc.Samples = append(sc.Samples, sample)
				continue
			}
			flows = sc.Flows
		} else {
			flows = cp.Flows
		}
		f := flowByID(flows, flowID)
		if f == nil {
			errs = append(errs, fmt.Errorf("%s: no step %q", e.Step, flowID))
			continue
		}
		ex := FlowExample{Source: source, Kind: ExampleKindDerived}
		if f.Example != nil {
			ex = *f.Example
		}
		ex.Samples = append(append([]Sample(nil), ex.Samples...), sample)
		f.Example = &ex
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return &cp, nil
}

// flowByID returns a pointer to the message flow with the ID, or nil.
func flowByID(flows []Flow, id string) *Flow {
	for i := range flows {
		if flows[i].ID == id && id != "" {
			return &flows[i]
		}
	}
	return nil
}
