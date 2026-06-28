package analyze

import (
	"fmt"

	"github.com/grokify/pidl"
)

// SecurityRule defines a security check.
type SecurityRule struct {
	// ID is a unique identifier for the rule.
	ID string

	// Name is a human-readable name.
	Name string

	// Description explains what the rule checks.
	Description string

	// Check runs the rule against a protocol and returns any risks found.
	Check func(p *pidl.Protocol) []SecurityRisk
}

// DefaultRules returns all built-in security rules.
func DefaultRules() []SecurityRule {
	return []SecurityRule{
		ruleUntrustToTrustWithoutSecurity(),
		ruleMissingEncryption(),
		ruleTokenWithoutBinding(),
		ruleFlowWithoutAuthentication(),
		ruleTokenWithoutAudience(),
		ruleBearerTokenExposure(),
		ruleMissingMTLS(),
		ruleExternalEntityWithoutTrustLevel(),
		ruleSensitiveDataInRedirect(),
		ruleWeakAuthenticationMethod(),
	}
}

// ruleUntrustToTrustWithoutSecurity checks for flows from untrusted to trusted
// entities without security requirements.
func ruleUntrustToTrustWithoutSecurity() SecurityRule {
	return SecurityRule{
		ID:          "SEC001",
		Name:        "Trust Boundary Violation",
		Description: "Flows from untrusted to trusted entities should have security controls",
		Check: func(p *pidl.Protocol) []SecurityRisk {
			var risks []SecurityRisk

			for i, flow := range p.Flows {
				fromEntity := p.EntityByID(flow.From)
				toEntity := p.EntityByID(flow.To)

				if fromEntity == nil || toEntity == nil {
					continue
				}

				// Check if crossing trust boundary without security
				fromTrust := trustWeight(fromEntity.TrustLevel)
				toTrust := trustWeight(toEntity.TrustLevel)

				if fromTrust < toTrust && !flow.HasSecurity() {
					risks = append(risks, SecurityRisk{
						ID:       "SEC001",
						Severity: SeverityHigh,
						Category: CategoryTrustBoundary,
						Title:    "Trust boundary crossing without security",
						Description: fmt.Sprintf(
							"Flow from %s (%s) to %s (%s) crosses trust boundary without security requirements",
							fromEntity.Name, fromEntity.TrustLevel, toEntity.Name, toEntity.TrustLevel),
						Location:    fmt.Sprintf("flows[%d]", i),
						Remediation: "Add security requirements (token, signature, mtls) to this flow",
					})
				}
			}

			return risks
		},
	}
}

// ruleMissingEncryption checks for confidential flows without encryption.
func ruleMissingEncryption() SecurityRule {
	return SecurityRule{
		ID:          "SEC002",
		Name:        "Missing Encryption",
		Description: "Confidential data flows should require encryption",
		Check: func(p *pidl.Protocol) []SecurityRisk {
			var risks []SecurityRisk

			for i, flow := range p.Flows {
				if flow.Security != nil && flow.Security.Confidential {
					// Check if encryption is required
					hasEncryption := false
					for _, req := range flow.Security.Requires {
						if req == pidl.SecurityRequirementEncryption || req == pidl.SecurityRequirementMTLS {
							hasEncryption = true
							break
						}
					}

					if !hasEncryption {
						risks = append(risks, SecurityRisk{
							ID:       "SEC002",
							Severity: SeverityHigh,
							Category: CategoryDataProtection,
							Title:    "Confidential flow without encryption",
							Description: fmt.Sprintf(
								"Flow %s->%s marked confidential but lacks encryption requirement",
								flow.From, flow.To),
							Location:    fmt.Sprintf("flows[%d]", i),
							Remediation: "Add 'encryption' or 'mtls' to the security requirements",
						})
					}
				}
			}

			return risks
		},
	}
}

// ruleTokenWithoutBinding checks for bearer tokens without binding mechanisms.
func ruleTokenWithoutBinding() SecurityRule {
	return SecurityRule{
		ID:          "SEC003",
		Name:        "Unbound Bearer Token",
		Description: "Bearer tokens should use binding mechanisms like mTLS or DPoP",
		Check: func(p *pidl.Protocol) []SecurityRisk {
			var risks []SecurityRisk

			if p.Metadata == nil {
				return risks
			}

			for _, token := range p.Metadata.Tokens {
				// Check if token is bearer (no binding)
				if token.Binding == "" || token.Binding == "bearer" {
					risks = append(risks, SecurityRisk{
						ID:       "SEC003",
						Severity: SeverityMedium,
						Category: CategoryTokenSecurity,
						Title:    "Bearer token without binding",
						Description: fmt.Sprintf(
							"Token '%s' uses bearer binding which is susceptible to theft",
							token.ID),
						Location:    fmt.Sprintf("metadata.tokens[%s]", token.ID),
						Remediation: "Consider using 'mtls' or 'dpop' binding for better security",
					})
				}
			}

			return risks
		},
	}
}

// ruleFlowWithoutAuthentication checks for flows to external entities without auth.
func ruleFlowWithoutAuthentication() SecurityRule {
	return SecurityRule{
		ID:          "SEC004",
		Name:        "Missing Authentication",
		Description: "Flows to sensitive endpoints should require authentication",
		Check: func(p *pidl.Protocol) []SecurityRisk {
			var risks []SecurityRisk

			for i, flow := range p.Flows {
				toEntity := p.EntityByID(flow.To)
				if toEntity == nil {
					continue
				}

				// Check if target is a server/service and flow lacks auth
				isServerType := toEntity.Type == pidl.EntityTypeServer ||
					toEntity.Type == pidl.EntityTypeResourceServer ||
					toEntity.Type == pidl.EntityTypeAuthorizationServer ||
					toEntity.Type == pidl.EntityTypeToolServer

				if isServerType && !flow.HasSecurity() {
					// Skip response/result flows
					if flow.Mode == pidl.FlowModeResponse || flow.Mode == pidl.FlowModeToolResult {
						continue
					}

					risks = append(risks, SecurityRisk{
						ID:       "SEC004",
						Severity: SeverityMedium,
						Category: CategoryAuthentication,
						Title:    "Flow to server without authentication",
						Description: fmt.Sprintf(
							"Flow from %s to %s (%s) lacks authentication requirement",
							flow.From, toEntity.Name, toEntity.Type),
						Location:    fmt.Sprintf("flows[%d]", i),
						Remediation: "Add security requirements (token, mtls, signature) to authenticate the request",
					})
				}
			}

			return risks
		},
	}
}

// ruleTokenWithoutAudience checks for tokens without defined audience.
func ruleTokenWithoutAudience() SecurityRule {
	return SecurityRule{
		ID:          "SEC005",
		Name:        "Token Without Audience",
		Description: "Tokens should have a defined audience to prevent misuse",
		Check: func(p *pidl.Protocol) []SecurityRisk {
			var risks []SecurityRisk

			if p.Metadata == nil {
				return risks
			}

			for _, token := range p.Metadata.Tokens {
				// JWT tokens should have audience
				if token.Type == "jwt" && token.Audience == "" {
					risks = append(risks, SecurityRisk{
						ID:       "SEC005",
						Severity: SeverityLow,
						Category: CategoryTokenSecurity,
						Title:    "JWT token without audience",
						Description: fmt.Sprintf(
							"Token '%s' is a JWT but has no defined audience",
							token.ID),
						Location:    fmt.Sprintf("metadata.tokens[%s]", token.ID),
						Remediation: "Define the token audience to prevent token confusion attacks",
					})
				}
			}

			return risks
		},
	}
}

// ruleBearerTokenExposure checks for bearer tokens in redirect flows.
func ruleBearerTokenExposure() SecurityRule {
	return SecurityRule{
		ID:          "SEC006",
		Name:        "Token in Redirect",
		Description: "Tokens should not be exposed in redirect URLs",
		Check: func(p *pidl.Protocol) []SecurityRisk {
			var risks []SecurityRisk

			for i, flow := range p.Flows {
				if flow.Mode != pidl.FlowModeRedirect && flow.Mode != pidl.FlowModeCallback {
					continue
				}

				// Check if flow passes token
				if flow.Security != nil && flow.Security.Token != "" {
					risks = append(risks, SecurityRisk{
						ID:       "SEC006",
						Severity: SeverityMedium,
						Category: CategoryTokenSecurity,
						Title:    "Token exposure in redirect",
						Description: fmt.Sprintf(
							"Redirect flow from %s to %s includes token '%s' which may be logged or leaked",
							flow.From, flow.To, flow.Security.Token),
						Location:    fmt.Sprintf("flows[%d]", i),
						Remediation: "Use authorization codes or fragment-based token delivery instead",
					})
				}
			}

			return risks
		},
	}
}

// ruleMissingMTLS checks for inter-service communication without mTLS.
func ruleMissingMTLS() SecurityRule {
	return SecurityRule{
		ID:          "SEC007",
		Name:        "Missing mTLS",
		Description: "Service-to-service communication should use mTLS",
		Check: func(p *pidl.Protocol) []SecurityRisk {
			var risks []SecurityRisk

			for i, flow := range p.Flows {
				fromEntity := p.EntityByID(flow.From)
				toEntity := p.EntityByID(flow.To)

				if fromEntity == nil || toEntity == nil {
					continue
				}

				// Check if both are servers/services
				isFromService := fromEntity.Type == pidl.EntityTypeServer ||
					fromEntity.Type == pidl.EntityTypeResourceServer ||
					fromEntity.Type == pidl.EntityTypeAuthorizationServer ||
					fromEntity.Type == pidl.EntityTypeToolServer

				isToService := toEntity.Type == pidl.EntityTypeServer ||
					toEntity.Type == pidl.EntityTypeResourceServer ||
					toEntity.Type == pidl.EntityTypeAuthorizationServer ||
					toEntity.Type == pidl.EntityTypeToolServer

				if isFromService && isToService {
					hasMTLS := false
					if flow.Security != nil {
						for _, req := range flow.Security.Requires {
							if req == pidl.SecurityRequirementMTLS {
								hasMTLS = true
								break
							}
						}
					}

					if !hasMTLS {
						risks = append(risks, SecurityRisk{
							ID:       "SEC007",
							Severity: SeverityInfo,
							Category: CategoryCommunication,
							Title:    "Service communication without mTLS",
							Description: fmt.Sprintf(
								"Service-to-service flow from %s to %s lacks mTLS",
								fromEntity.Name, toEntity.Name),
							Location:    fmt.Sprintf("flows[%d]", i),
							Remediation: "Consider adding mTLS for service-to-service authentication",
						})
					}
				}
			}

			return risks
		},
	}
}

// ruleExternalEntityWithoutTrustLevel checks for entities missing trust classification.
func ruleExternalEntityWithoutTrustLevel() SecurityRule {
	return SecurityRule{
		ID:          "SEC008",
		Name:        "Missing Trust Level",
		Description: "Entities should have explicit trust level classification",
		Check: func(p *pidl.Protocol) []SecurityRisk {
			var risks []SecurityRisk

			for _, entity := range p.Entities {
				if entity.TrustLevel == "" {
					// External-facing entity types need explicit trust
					if entity.Type == pidl.EntityTypeClient ||
						entity.Type == pidl.EntityTypeUser ||
						entity.Type == pidl.EntityTypeBrowser ||
						entity.Type == pidl.EntityTypeAgent {
						risks = append(risks, SecurityRisk{
							ID:       "SEC008",
							Severity: SeverityLow,
							Category: CategoryConfiguration,
							Title:    "Entity without trust level",
							Description: fmt.Sprintf(
								"Entity '%s' (%s) has no explicit trust level",
								entity.Name, entity.Type),
							Location:    fmt.Sprintf("entities[%s]", entity.ID),
							Remediation: "Add trust_level: untrusted, semi_trusted, trusted, or authoritative",
						})
					}
				}
			}

			return risks
		},
	}
}

// ruleSensitiveDataInRedirect checks for sensitive data in redirect flows.
func ruleSensitiveDataInRedirect() SecurityRule {
	return SecurityRule{
		ID:          "SEC009",
		Name:        "Sensitive Data in Redirect",
		Description: "Confidential data should not be passed in redirect URLs",
		Check: func(p *pidl.Protocol) []SecurityRisk {
			var risks []SecurityRisk

			for i, flow := range p.Flows {
				if flow.Mode != pidl.FlowModeRedirect && flow.Mode != pidl.FlowModeCallback {
					continue
				}

				if flow.Security != nil && flow.Security.Confidential {
					risks = append(risks, SecurityRisk{
						ID:       "SEC009",
						Severity: SeverityHigh,
						Category: CategoryDataProtection,
						Title:    "Confidential data in redirect",
						Description: fmt.Sprintf(
							"Redirect from %s to %s is marked confidential but redirects may expose data in URLs",
							flow.From, flow.To),
						Location:    fmt.Sprintf("flows[%d]", i),
						Remediation: "Use back-channel communication for confidential data",
					})
				}
			}

			return risks
		},
	}
}

// ruleWeakAuthenticationMethod checks for potentially weak auth methods.
func ruleWeakAuthenticationMethod() SecurityRule {
	return SecurityRule{
		ID:          "SEC010",
		Name:        "Weak Authentication",
		Description: "Check for potentially weak authentication patterns",
		Check: func(p *pidl.Protocol) []SecurityRisk {
			var risks []SecurityRisk

			if p.Metadata == nil {
				return risks
			}

			for _, token := range p.Metadata.Tokens {
				// API keys are weaker than other token types
				if token.Type == "api_key" {
					risks = append(risks, SecurityRisk{
						ID:       "SEC010",
						Severity: SeverityInfo,
						Category: CategoryAuthentication,
						Title:    "API key authentication",
						Description: fmt.Sprintf(
							"Token '%s' uses API key which is less secure than JWT or OAuth tokens",
							token.ID),
						Location:    fmt.Sprintf("metadata.tokens[%s]", token.ID),
						Remediation: "Consider using JWT tokens with proper signing and rotation",
					})
				}
			}

			return risks
		},
	}
}

// trustWeight returns a numeric weight for trust levels (higher = more trusted).
func trustWeight(level pidl.TrustLevel) int {
	switch level {
	case pidl.TrustLevelAuthoritative:
		return 4
	case pidl.TrustLevelTrusted:
		return 3
	case pidl.TrustLevelSemiTrusted:
		return 2
	case pidl.TrustLevelUntrusted:
		return 1
	default:
		return 0 // Unknown/unset
	}
}
