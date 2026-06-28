package pidl

// IsValidCategory checks if the category is valid.
func IsValidCategory(c Category) bool {
	switch c {
	case CategoryAuth, CategoryAgent, CategoryMessaging, CategoryProvisioning, CategoryOther:
		return true
	}
	return false
}

// IsValidEntityType checks if the entity type is valid.
func IsValidEntityType(t EntityType) bool {
	switch t {
	case EntityTypeClient, EntityTypeAuthorizationServer, EntityTypeResourceServer,
		EntityTypeUser, EntityTypeBrowser, EntityTypeAgent, EntityTypeToolServer,
		EntityTypeTool, EntityTypeDelegatedAgent, EntityTypeIdentityProvider,
		EntityTypeServiceProvider, EntityTypeServer, EntityTypeOther:
		return true
	}
	return false
}

// IsValidFlowMode checks if the flow mode is valid.
func IsValidFlowMode(m FlowMode) bool {
	switch m {
	case FlowModeRequest, FlowModeResponse, FlowModeRedirect, FlowModeCallback,
		FlowModeInteractive, FlowModeEvent, FlowModeToolCall, FlowModeToolResult:
		return true
	case "": // empty is valid, defaults to request
		return true
	}
	return false
}

// IsValidAnnotationType checks if the annotation type is valid.
func IsValidAnnotationType(t AnnotationType) bool {
	switch t {
	case AnnotationTypeSecurity, AnnotationTypePerformance, AnnotationTypeDeprecated,
		AnnotationTypeInfo, AnnotationTypeWarning, AnnotationTypeError:
		return true
	}
	return false
}

// IsValidTrustLevel checks if the trust level is valid.
func IsValidTrustLevel(t TrustLevel) bool {
	switch t {
	case TrustLevelTrusted, TrustLevelSemiTrusted, TrustLevelUntrusted, TrustLevelAuthoritative:
		return true
	case "": // empty is valid
		return true
	}
	return false
}

// IsValidSecurityRequirement checks if the security requirement is valid.
func IsValidSecurityRequirement(r SecurityRequirement) bool {
	switch r {
	case SecurityRequirementToken, SecurityRequirementSignature, SecurityRequirementEncryption,
		SecurityRequirementMTLS, SecurityRequirementMAC:
		return true
	}
	return false
}

// IsValidProtocol checks if the protocol identifier is a known protocol.
func IsValidProtocol(protocol string) bool {
	switch protocol {
	case ProtocolOAuth, ProtocolSCIM, ProtocolSPIFFE, ProtocolAAuth,
		ProtocolIDJAG, ProtocolAuthZEN, ProtocolMCP, ProtocolA2A:
		return true
	}
	return false
}

// IsValidComponentType checks if the component type is a known type.
func IsValidComponentType(t string) bool {
	switch t {
	case ComponentTypeIdP, ComponentTypeIGA, ComponentTypeAgentProvider,
		ComponentTypePersonServer, ComponentTypeAccessServer, ComponentTypePDP,
		ComponentTypeGateway, ComponentTypeMCPClient, ComponentTypeMCPServer,
		ComponentTypeResourceAPI, ComponentTypeSPIRE:
		return true
	}
	return false
}

// IsValidTrustRelationType checks if the trust relation type is a known type.
func IsValidTrustRelationType(t string) bool {
	switch t {
	case TrustTypeAuthenticates, TrustTypeValidates, TrustTypeDelegates,
		TrustTypeAuthorizes, TrustTypeIssues, TrustTypeTrusts,
		TrustTypeProvisions, TrustTypeAttests:
		return true
	}
	return false
}

// IsValidCredential checks if the credential type is a known type.
func IsValidCredential(c string) bool {
	switch c {
	case CredentialX509SVID, CredentialJWTSVID, CredentialJWTAssertion,
		CredentialAccessToken, CredentialIDToken, CredentialAAAgentJWT,
		CredentialAAAuthJWT, CredentialMTLS, CredentialAPIKey:
		return true
	}
	return false
}
