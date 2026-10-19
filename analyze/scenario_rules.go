package analyze

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/grokify/pidl"
)

// flowSite is one flow with the location used in risk reports.
type flowSite struct {
	flow     pidl.Flow
	location string
}

// scenarioRuns returns the flow sequences to walk in order: each main
// scenario with its fragments expanded, or the top-level flows for a
// document without scenarios. Locations point at the flow as written,
// which for an expanded fragment flow is inside the fragment.
func scenarioRuns(p *pidl.Protocol) [][]flowSite {
	if !p.HasScenarios() {
		run := make([]flowSite, len(p.Flows))
		for i, f := range p.Flows {
			run[i] = flowSite{flow: f, location: fmt.Sprintf("flows[%d]", i)}
		}
		return [][]flowSite{run}
	}
	index := make(map[string]int, len(p.Scenarios))
	for i, sc := range p.Scenarios {
		index[sc.ID] = i
	}
	var expand func(id string, seen []string) []flowSite
	expand = func(id string, seen []string) []flowSite {
		if slices.Contains(seen, id) {
			return nil // include cycle; validation reports it
		}
		i, ok := index[id]
		if !ok {
			return nil
		}
		var out []flowSite
		for j, f := range p.Scenarios[i].Flows {
			if f.IsInclude() {
				out = append(out, expand(f.Include, append(slices.Clone(seen), id))...)
				continue
			}
			out = append(out, flowSite{flow: f, location: fmt.Sprintf("scenarios[%d].flows[%d] (%s)", i, j, id)})
		}
		return out
	}
	var runs [][]flowSite
	for _, id := range p.ScenarioIDs(false) {
		runs = append(runs, expand(id, nil))
	}
	return runs
}

// writtenFlows returns every message flow once, as written.
func writtenFlows(p *pidl.Protocol) []flowSite {
	if !p.HasScenarios() {
		return scenarioRuns(p)[0]
	}
	var out []flowSite
	for i, sc := range p.Scenarios {
		for j, f := range sc.Flows {
			if !f.IsInclude() {
				out = append(out, flowSite{flow: f, location: fmt.Sprintf("scenarios[%d].flows[%d] (%s)", i, j, sc.ID)})
			}
		}
	}
	return out
}

// ruleUnusedRole reports entities that no flow uses.
func ruleUnusedRole() SecurityRule {
	return SecurityRule{
		ID:          "SEC017",
		Name:        "Unused Role",
		Description: "Every entity should take part in at least one flow",
		Scope:       ScopeDocument,
		Check: func(p *pidl.Protocol) []SecurityRisk {
			used := map[string]bool{}
			for _, site := range writtenFlows(p) {
				used[site.flow.From] = true
				used[site.flow.To] = true
			}
			var risks []SecurityRisk
			for i, e := range p.Entities {
				if used[e.ID] {
					continue
				}
				risks = append(risks, SecurityRisk{
					ID:          "SEC017",
					Severity:    SeverityLow,
					Category:    CategoryConfiguration,
					Title:       "Unused role",
					Description: fmt.Sprintf("Entity %q takes part in no flow, so the document never shows what it does", e.ID),
					Location:    fmt.Sprintf("entities[%d]", i),
					Remediation: "Add the flows this role takes part in, or remove the entity",
				})
			}
			return risks
		},
	}
}

// ruleTokenCheckedBeforeIssued reports a flow that verifies a token before
// any earlier step of the same scenario issues it. Tokens that no flow in
// the document issues are assumed to come from outside and are skipped.
func ruleTokenCheckedBeforeIssued() SecurityRule {
	return SecurityRule{
		ID:          "SEC018",
		Name:        "Token Checked Before Issued",
		Description: "A scenario should issue a token before any step checks it",
		Scope:       ScopeDocument,
		Check: func(p *pidl.Protocol) []SecurityRisk {
			issuedAnywhere := map[string]bool{}
			for _, site := range writtenFlows(p) {
				if site.flow.Security != nil {
					for _, t := range site.flow.Security.Issues {
						issuedAnywhere[t] = true
					}
				}
			}
			var risks []SecurityRisk
			reported := map[string]bool{}
			for _, run := range scenarioRuns(p) {
				issued := map[string]bool{}
				for _, site := range run {
					sec := site.flow.Security
					if sec == nil {
						continue
					}
					for _, t := range sec.Verifies {
						key := site.location + "|" + t
						if issuedAnywhere[t] && !issued[t] && !reported[key] {
							reported[key] = true
							risks = append(risks, SecurityRisk{
								ID:          "SEC018",
								Severity:    SeverityMedium,
								Category:    CategoryTokenSecurity,
								Title:       "Token checked before it is issued",
								Description: fmt.Sprintf("This step checks token %q, but no earlier step of the scenario issues it", t),
								Location:    site.location,
								Remediation: "Move the step that issues the token earlier, or include the fragment that issues it",
							})
						}
					}
					for _, t := range sec.Issues {
						issued[t] = true
					}
				}
			}
			return risks
		},
	}
}

// ruleSignerDiffersFromSender reports requests whose signer is not the
// entity a key-binding check would expect: either signed_by names someone
// other than the sender, or the request carries a token that an earlier
// step delivered to a different entity than the signer. In both cases any
// check that binds a token to the request signature must use the signer's
// key, a common source of protocol mistakes.
func ruleSignerDiffersFromSender() SecurityRule {
	return SecurityRule{
		ID:          "SEC019",
		Name:        "Signer Differs from Sender or Token Holder",
		Description: "Key-binding checks must use the key of the entity that actually signs",
		Scope:       ScopeDocument,
		Check: func(p *pidl.Protocol) []SecurityRisk {
			var risks []SecurityRisk
			reported := map[string]bool{}
			report := func(site flowSite, desc string) {
				if reported[site.location] {
					return
				}
				reported[site.location] = true
				risks = append(risks, SecurityRisk{
					ID:          "SEC019",
					Severity:    SeverityLow,
					Category:    CategoryAuthentication,
					Title:       "Request signed by someone other than the sender or token holder",
					Description: desc,
					Location:    site.location,
					Remediation: "State which key each check uses, and add a security annotation explaining the signer",
				})
			}
			for _, run := range scenarioRuns(p) {
				holder := map[string]string{} // token ID -> entity it was delivered to
				for _, site := range run {
					f := site.flow
					if f.Security == nil {
						continue
					}
					signer := f.From
					if f.Security.SignedBy != "" {
						signer = f.Security.SignedBy
					}
					if !hasSecurityAnnotation(f) {
						if signer != f.From {
							report(site, fmt.Sprintf("%s sends this request but %s signs it; any check that binds a token to the request signature must use %s's key",
								f.From, signer, signer))
						}
						for _, t := range f.Security.Verifies {
							if h, ok := holder[t]; ok && h != signer {
								report(site, fmt.Sprintf("Token %q was delivered to %s, but %s signs the request that presents it; a check binding the token to the signing key (such as a key thumbprint claim) will compare against %s's key",
									t, h, signer, signer))
							}
						}
					}
					for _, t := range f.Security.Issues {
						holder[t] = f.To
					}
				}
			}
			return risks
		},
	}
}

// ruleVerifierMismatch reports a flow whose receiver checks a token that
// does not list the receiver in verified_by.
func ruleVerifierMismatch() SecurityRule {
	return SecurityRule{
		ID:          "SEC020",
		Name:        "Token Verifier Mismatch",
		Description: "The receiver of a flow that checks a token should be one of the token's verifiers",
		Scope:       ScopeDocument,
		Check: func(p *pidl.Protocol) []SecurityRisk {
			var risks []SecurityRisk
			for _, site := range writtenFlows(p) {
				f := site.flow
				if f.Security == nil {
					continue
				}
				for _, id := range f.Security.Verifies {
					t := p.TokenByID(id)
					if t == nil || len(t.VerifiedBy) == 0 || slices.Contains(t.VerifiedBy, f.To) {
						continue
					}
					risks = append(risks, SecurityRisk{
						ID:          "SEC020",
						Severity:    SeverityLow,
						Category:    CategoryTokenSecurity,
						Title:       "Token checked by an entity not listed as a verifier",
						Description: fmt.Sprintf("%s checks token %q, which lists only %v as verifiers", f.To, id, p.OrderEntities(t.VerifiedBy)),
						Location:    site.location,
						Remediation: fmt.Sprintf("Add %s to the token's verified_by, or correct the flow", f.To),
					})
				}
			}
			return risks
		},
	}
}

// ruleUnusedFragment reports fragment scenarios that no scenario includes.
func ruleUnusedFragment() SecurityRule {
	return SecurityRule{
		ID:          "SEC021",
		Name:        "Unused Fragment",
		Description: "Every fragment scenario should be included somewhere",
		Scope:       ScopeDocument,
		Check: func(p *pidl.Protocol) []SecurityRisk {
			included := map[string]bool{}
			for _, sc := range p.Scenarios {
				for _, f := range sc.Flows {
					if f.IsInclude() {
						included[f.Include] = true
					}
				}
			}
			var risks []SecurityRisk
			for i, sc := range p.Scenarios {
				if !sc.Fragment || included[sc.ID] {
					continue
				}
				risks = append(risks, SecurityRisk{
					ID:          "SEC021",
					Severity:    SeverityInfo,
					Category:    CategoryConfiguration,
					Title:       "Unused fragment",
					Description: fmt.Sprintf("Fragment %q is not included by any scenario", sc.ID),
					Location:    fmt.Sprintf("scenarios[%d]", i),
					Remediation: "Include the fragment where it applies, or make it a main scenario",
				})
			}
			return risks
		},
	}
}

// hasSecurityAnnotation reports whether a flow carries a security annotation.
func hasSecurityAnnotation(f pidl.Flow) bool {
	for _, a := range f.Annotations {
		if a.Type == pidl.AnnotationTypeSecurity {
			return true
		}
	}
	return false
}

// flattenedFlowLocation matches the locations flow rules report.
var flattenedFlowLocation = regexp.MustCompile(`^flows\[(\d+)\]$`)

// remapFlattenedLocations rewrites "flows[i]" locations, which index the
// flattened protocol, to the flow as written, such as
// "scenarios[1].flows[0] (consent)". writtenFlows lists flows in the same
// order Protocol.Flatten does.
func remapFlattenedLocations(p *pidl.Protocol, risks []SecurityRisk) {
	sites := writtenFlows(p)
	for i := range risks {
		m := flattenedFlowLocation.FindStringSubmatch(risks[i].Location)
		if m == nil {
			continue
		}
		var idx int
		if _, err := fmt.Sscanf(m[1], "%d", &idx); err != nil || idx >= len(sites) {
			continue // not a flattened index we produced; leave it as reported
		}
		risks[i].Location = sites[idx].location
	}
}
