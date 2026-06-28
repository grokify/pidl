package pidl

import (
	"testing"
)

func TestFlowDisplayLabel(t *testing.T) {
	tests := []struct {
		name     string
		flow     Flow
		expected string
	}{
		{
			name:     "label takes precedence",
			flow:     Flow{Action: "action", Label: "Label"},
			expected: "Label",
		},
		{
			name:     "falls back to action",
			flow:     Flow{Action: "action"},
			expected: "action",
		},
		{
			name:     "empty returns empty action",
			flow:     Flow{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flow.DisplayLabel(); got != tt.expected {
				t.Errorf("DisplayLabel() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFlowEffectiveMode(t *testing.T) {
	tests := []struct {
		name     string
		flow     Flow
		expected FlowMode
	}{
		{
			name:     "explicit mode",
			flow:     Flow{Mode: FlowModeResponse},
			expected: FlowModeResponse,
		},
		{
			name:     "defaults to request",
			flow:     Flow{},
			expected: FlowModeRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flow.EffectiveMode(); got != tt.expected {
				t.Errorf("EffectiveMode() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestProtocolEntityByID(t *testing.T) {
	p := &Protocol{
		Entities: []Entity{
			{ID: "client", Name: "Client"},
			{ID: "server", Name: "Server"},
		},
	}

	if e := p.EntityByID("client"); e == nil || e.Name != "Client" {
		t.Error("EntityByID should find client")
	}

	if e := p.EntityByID("unknown"); e != nil {
		t.Error("EntityByID should return nil for unknown ID")
	}
}

func TestProtocolPhaseByID(t *testing.T) {
	p := &Protocol{
		Phases: []Phase{
			{ID: "auth", Name: "Authorization"},
			{ID: "token", Name: "Token Exchange"},
		},
	}

	if ph := p.PhaseByID("auth"); ph == nil || ph.Name != "Authorization" {
		t.Error("PhaseByID should find auth phase")
	}

	if ph := p.PhaseByID("unknown"); ph != nil {
		t.Error("PhaseByID should return nil for unknown ID")
	}
}

func TestProtocolFlowsByPhase(t *testing.T) {
	p := &Protocol{
		Flows: []Flow{
			{From: "a", To: "b", Action: "x", Phase: "auth"},
			{From: "b", To: "c", Action: "y", Phase: "token"},
			{From: "a", To: "c", Action: "z", Phase: "auth"},
		},
	}

	authFlows := p.FlowsByPhase("auth")
	if len(authFlows) != 2 {
		t.Errorf("FlowsByPhase(auth) = %d flows, want 2", len(authFlows))
	}

	tokenFlows := p.FlowsByPhase("token")
	if len(tokenFlows) != 1 {
		t.Errorf("FlowsByPhase(token) = %d flows, want 1", len(tokenFlows))
	}

	unknownFlows := p.FlowsByPhase("unknown")
	if len(unknownFlows) != 0 {
		t.Errorf("FlowsByPhase(unknown) = %d flows, want 0", len(unknownFlows))
	}
}

func TestProtocolEntityIDs(t *testing.T) {
	p := &Protocol{
		Entities: []Entity{
			{ID: "a"},
			{ID: "b"},
			{ID: "c"},
		},
	}

	ids := p.EntityIDs()
	if len(ids) != 3 {
		t.Errorf("EntityIDs() = %d IDs, want 3", len(ids))
	}
	if ids[0] != "a" || ids[1] != "b" || ids[2] != "c" {
		t.Errorf("EntityIDs() = %v, want [a b c]", ids)
	}
}

func TestEntityHasStates(t *testing.T) {
	e := Entity{ID: "client"}
	if e.HasStates() {
		t.Error("HasStates() = true for entity without states, want false")
	}

	e.States = []EntityState{{ID: "idle"}}
	if !e.HasStates() {
		t.Error("HasStates() = false for entity with states, want true")
	}
}

func TestEntityStateByID(t *testing.T) {
	e := Entity{
		ID: "client",
		States: []EntityState{
			{ID: "idle", Name: "Idle"},
			{ID: "active", Name: "Active"},
		},
	}

	if s := e.StateByID("idle"); s == nil || s.Name != "Idle" {
		t.Error("StateByID should find idle state")
	}

	if s := e.StateByID("unknown"); s != nil {
		t.Error("StateByID should return nil for unknown ID")
	}
}

func TestEntityInitialState(t *testing.T) {
	e := Entity{
		ID: "client",
		States: []EntityState{
			{ID: "idle", Initial: true},
			{ID: "active"},
		},
	}

	if s := e.InitialState(); s == nil || s.ID != "idle" {
		t.Error("InitialState should find initial state")
	}

	e.States = []EntityState{{ID: "active"}}
	if s := e.InitialState(); s != nil {
		t.Error("InitialState should return nil when no initial state")
	}
}

func TestEntityStateIDs(t *testing.T) {
	e := Entity{
		ID: "client",
		States: []EntityState{
			{ID: "idle"},
			{ID: "active"},
			{ID: "error"},
		},
	}

	ids := e.StateIDs()
	if len(ids) != 3 {
		t.Errorf("StateIDs() = %d IDs, want 3", len(ids))
	}
	if ids[0] != "idle" || ids[1] != "active" || ids[2] != "error" {
		t.Errorf("StateIDs() = %v, want [idle active error]", ids)
	}
}

func TestFlowHasStateMutations(t *testing.T) {
	f := Flow{From: "a", To: "b", Action: "test"}
	if f.HasStateMutations() {
		t.Error("HasStateMutations() = true for flow without sets, want false")
	}

	f.Sets = []StateMutation{{Entity: "client", To: "active"}}
	if !f.HasStateMutations() {
		t.Error("HasStateMutations() = false for flow with sets, want true")
	}
}

func TestProtocolEntitiesWithStates(t *testing.T) {
	p := &Protocol{
		Entities: []Entity{
			{ID: "client", States: []EntityState{{ID: "idle"}}},
			{ID: "server"},
			{ID: "db", States: []EntityState{{ID: "ready"}}},
		},
	}

	entities := p.EntitiesWithStates()
	if len(entities) != 2 {
		t.Errorf("EntitiesWithStates() = %d entities, want 2", len(entities))
	}
	if entities[0].ID != "client" || entities[1].ID != "db" {
		t.Error("EntitiesWithStates() should return client and db")
	}
}

func TestProtocolStateTransitions(t *testing.T) {
	p := &Protocol{
		Entities: []Entity{
			{ID: "client", States: []EntityState{{ID: "idle"}, {ID: "active"}}},
		},
		Flows: []Flow{
			{
				From:   "client",
				To:     "server",
				Action: "login",
				Label:  "Login",
				Sets: []StateMutation{
					{Entity: "client", From: "idle", To: "active"},
				},
			},
			{
				From:   "server",
				To:     "client",
				Action: "response",
			},
		},
	}

	transitions := p.StateTransitions()
	if len(transitions) != 1 {
		t.Errorf("StateTransitions() = %d transitions, want 1", len(transitions))
	}

	t0 := transitions[0]
	if t0.EntityID != "client" {
		t.Errorf("EntityID = %q, want %q", t0.EntityID, "client")
	}
	if t0.FromState != "idle" {
		t.Errorf("FromState = %q, want %q", t0.FromState, "idle")
	}
	if t0.ToState != "active" {
		t.Errorf("ToState = %q, want %q", t0.ToState, "active")
	}
	if t0.FlowAction != "login" {
		t.Errorf("FlowAction = %q, want %q", t0.FlowAction, "login")
	}
	if t0.FlowLabel != "Login" {
		t.Errorf("FlowLabel = %q, want %q", t0.FlowLabel, "Login")
	}
}

func TestProtocolStateTransitionsForEntity(t *testing.T) {
	p := &Protocol{
		Entities: []Entity{
			{ID: "client", States: []EntityState{{ID: "idle"}, {ID: "active"}}},
			{ID: "server", States: []EntityState{{ID: "ready"}, {ID: "busy"}}},
		},
		Flows: []Flow{
			{
				From:   "client",
				To:     "server",
				Action: "request",
				Sets: []StateMutation{
					{Entity: "client", To: "active"},
					{Entity: "server", To: "busy"},
				},
			},
		},
	}

	clientTransitions := p.StateTransitionsForEntity("client")
	if len(clientTransitions) != 1 {
		t.Errorf("StateTransitionsForEntity(client) = %d, want 1", len(clientTransitions))
	}

	serverTransitions := p.StateTransitionsForEntity("server")
	if len(serverTransitions) != 1 {
		t.Errorf("StateTransitionsForEntity(server) = %d, want 1", len(serverTransitions))
	}

	unknownTransitions := p.StateTransitionsForEntity("unknown")
	if len(unknownTransitions) != 0 {
		t.Errorf("StateTransitionsForEntity(unknown) = %d, want 0", len(unknownTransitions))
	}
}

func TestFlowHasSecurity(t *testing.T) {
	// Flow without security
	f := Flow{From: "a", To: "b", Action: "test"}
	if f.HasSecurity() {
		t.Error("HasSecurity() = true for flow without security, want false")
	}

	// Flow with empty security struct
	f.Security = &FlowSecurity{}
	if f.HasSecurity() {
		t.Error("HasSecurity() = true for flow with empty security, want false")
	}

	// Flow with security requirements
	f.Security = &FlowSecurity{Requires: []SecurityRequirement{SecurityRequirementToken}}
	if !f.HasSecurity() {
		t.Error("HasSecurity() = false for flow with requirements, want true")
	}

	// Flow with token
	f.Security = &FlowSecurity{Token: "access_token"}
	if !f.HasSecurity() {
		t.Error("HasSecurity() = false for flow with token, want true")
	}

	// Flow with confidential flag
	f.Security = &FlowSecurity{Confidential: true}
	if !f.HasSecurity() {
		t.Error("HasSecurity() = false for flow with confidential, want true")
	}
}

func TestFlowRequiresEncryption(t *testing.T) {
	f := Flow{From: "a", To: "b", Action: "test"}

	if f.RequiresEncryption() {
		t.Error("RequiresEncryption() = true for flow without security, want false")
	}

	f.Security = &FlowSecurity{Requires: []SecurityRequirement{SecurityRequirementToken}}
	if f.RequiresEncryption() {
		t.Error("RequiresEncryption() = true for flow with token only, want false")
	}

	f.Security = &FlowSecurity{Requires: []SecurityRequirement{SecurityRequirementEncryption}}
	if !f.RequiresEncryption() {
		t.Error("RequiresEncryption() = false for flow with encryption, want true")
	}
}

func TestFlowRequiresToken(t *testing.T) {
	f := Flow{From: "a", To: "b", Action: "test"}

	if f.RequiresToken() {
		t.Error("RequiresToken() = true for flow without security, want false")
	}

	f.Security = &FlowSecurity{Requires: []SecurityRequirement{SecurityRequirementEncryption}}
	if f.RequiresToken() {
		t.Error("RequiresToken() = true for flow with encryption only, want false")
	}

	f.Security = &FlowSecurity{Requires: []SecurityRequirement{SecurityRequirementToken}}
	if !f.RequiresToken() {
		t.Error("RequiresToken() = false for flow with token requirement, want true")
	}

	f.Security = &FlowSecurity{Token: "access_token"}
	if !f.RequiresToken() {
		t.Error("RequiresToken() = false for flow with token reference, want true")
	}
}

func TestProtocolTokenByID(t *testing.T) {
	p := &Protocol{
		Metadata: &ProtocolMetadata{
			Tokens: []TokenDefinition{
				{ID: "access_token", Name: "Access Token", Type: "jwt"},
				{ID: "refresh_token", Name: "Refresh Token", Type: "opaque"},
			},
		},
	}

	token := p.TokenByID("access_token")
	if token == nil || token.Name != "Access Token" {
		t.Error("TokenByID should find access_token")
	}

	token = p.TokenByID("unknown")
	if token != nil {
		t.Error("TokenByID should return nil for unknown ID")
	}

	// Test with nil metadata
	p.Metadata = nil
	token = p.TokenByID("access_token")
	if token != nil {
		t.Error("TokenByID should return nil when metadata is nil")
	}
}

func TestIsValidTrustLevel(t *testing.T) {
	validLevels := []TrustLevel{
		TrustLevelTrusted,
		TrustLevelSemiTrusted,
		TrustLevelUntrusted,
		TrustLevelAuthoritative,
	}

	for _, level := range validLevels {
		if !IsValidTrustLevel(level) {
			t.Errorf("IsValidTrustLevel(%q) = false, want true", level)
		}
	}

	if IsValidTrustLevel("invalid") {
		t.Error("IsValidTrustLevel(invalid) = true, want false")
	}
}

func TestIsValidSecurityRequirement(t *testing.T) {
	validReqs := []SecurityRequirement{
		SecurityRequirementToken,
		SecurityRequirementSignature,
		SecurityRequirementEncryption,
		SecurityRequirementMTLS,
		SecurityRequirementMAC,
	}

	for _, req := range validReqs {
		if !IsValidSecurityRequirement(req) {
			t.Errorf("IsValidSecurityRequirement(%q) = false, want true", req)
		}
	}

	if IsValidSecurityRequirement("invalid") {
		t.Error("IsValidSecurityRequirement(invalid) = true, want false")
	}
}

func TestEntityHasProtocolRoles(t *testing.T) {
	e := Entity{ID: "test"}
	if e.HasProtocolRoles() {
		t.Error("HasProtocolRoles should return false for entity without roles")
	}

	e.ProtocolRoles = []ProtocolRole{
		{Protocol: ProtocolOAuth, Role: "authorization_server"},
	}
	if !e.HasProtocolRoles() {
		t.Error("HasProtocolRoles should return true for entity with roles")
	}
}

func TestEntityRolesByProtocol(t *testing.T) {
	e := Entity{
		ID: "test",
		ProtocolRoles: []ProtocolRole{
			{Protocol: ProtocolOAuth, Role: "authorization_server"},
			{Protocol: ProtocolOAuth, Role: "token_endpoint"},
			{Protocol: ProtocolSCIM, Role: "service_provider"},
		},
	}

	oauthRoles := e.RolesByProtocol(ProtocolOAuth)
	if len(oauthRoles) != 2 {
		t.Errorf("RolesByProtocol(oauth) returned %d roles, want 2", len(oauthRoles))
	}

	scimRoles := e.RolesByProtocol(ProtocolSCIM)
	if len(scimRoles) != 1 {
		t.Errorf("RolesByProtocol(scim) returned %d roles, want 1", len(scimRoles))
	}

	mcpRoles := e.RolesByProtocol(ProtocolMCP)
	if len(mcpRoles) != 0 {
		t.Errorf("RolesByProtocol(mcp) returned %d roles, want 0", len(mcpRoles))
	}
}

func TestEntityHasRole(t *testing.T) {
	e := Entity{
		ID: "test",
		ProtocolRoles: []ProtocolRole{
			{Protocol: ProtocolOAuth, Role: "authorization_server"},
		},
	}

	if !e.HasRole(ProtocolOAuth, "authorization_server") {
		t.Error("HasRole should return true for existing role")
	}

	if e.HasRole(ProtocolOAuth, "client") {
		t.Error("HasRole should return false for non-existing role")
	}

	if e.HasRole(ProtocolSCIM, "authorization_server") {
		t.Error("HasRole should return false for wrong protocol")
	}
}

func TestProtocolEntitiesWithProtocolRoles(t *testing.T) {
	p := &Protocol{
		Entities: []Entity{
			{ID: "e1"},
			{ID: "e2", ProtocolRoles: []ProtocolRole{{Protocol: ProtocolOAuth, Role: "client"}}},
			{ID: "e3", ProtocolRoles: []ProtocolRole{{Protocol: ProtocolSCIM, Role: "service_provider"}}},
		},
	}

	entities := p.EntitiesWithProtocolRoles()
	if len(entities) != 2 {
		t.Errorf("EntitiesWithProtocolRoles returned %d entities, want 2", len(entities))
	}
}

func TestProtocolEntitiesByProtocol(t *testing.T) {
	p := &Protocol{
		Entities: []Entity{
			{ID: "e1", ProtocolRoles: []ProtocolRole{{Protocol: ProtocolOAuth, Role: "client"}}},
			{ID: "e2", ProtocolRoles: []ProtocolRole{{Protocol: ProtocolOAuth, Role: "resource_server"}}},
			{ID: "e3", ProtocolRoles: []ProtocolRole{{Protocol: ProtocolSCIM, Role: "service_provider"}}},
		},
	}

	oauthEntities := p.EntitiesByProtocol(ProtocolOAuth)
	if len(oauthEntities) != 2 {
		t.Errorf("EntitiesByProtocol(oauth) returned %d entities, want 2", len(oauthEntities))
	}

	scimEntities := p.EntitiesByProtocol(ProtocolSCIM)
	if len(scimEntities) != 1 {
		t.Errorf("EntitiesByProtocol(scim) returned %d entities, want 1", len(scimEntities))
	}
}

func TestProtocolEntitiesByRole(t *testing.T) {
	p := &Protocol{
		Entities: []Entity{
			{ID: "e1", ProtocolRoles: []ProtocolRole{{Protocol: ProtocolOAuth, Role: "client"}}},
			{ID: "e2", ProtocolRoles: []ProtocolRole{{Protocol: ProtocolOAuth, Role: "client"}}},
			{ID: "e3", ProtocolRoles: []ProtocolRole{{Protocol: ProtocolOAuth, Role: "resource_server"}}},
		},
	}

	clients := p.EntitiesByRole(ProtocolOAuth, "client")
	if len(clients) != 2 {
		t.Errorf("EntitiesByRole(oauth, client) returned %d entities, want 2", len(clients))
	}
}

func TestProtocolComponentByID(t *testing.T) {
	p := &Protocol{
		Metadata: &ProtocolMetadata{
			Components: []DeploymentComponent{
				{ID: "idp", Name: "Identity Provider", Type: ComponentTypeIdP},
				{ID: "gateway", Name: "API Gateway", Type: ComponentTypeGateway},
			},
		},
	}

	c := p.ComponentByID("idp")
	if c == nil || c.Name != "Identity Provider" {
		t.Error("ComponentByID should find idp")
	}

	c = p.ComponentByID("unknown")
	if c != nil {
		t.Error("ComponentByID should return nil for unknown ID")
	}

	p.Metadata = nil
	c = p.ComponentByID("idp")
	if c != nil {
		t.Error("ComponentByID should return nil when metadata is nil")
	}
}

func TestProtocolComponentsByType(t *testing.T) {
	p := &Protocol{
		Metadata: &ProtocolMetadata{
			Components: []DeploymentComponent{
				{ID: "idp1", Type: ComponentTypeIdP},
				{ID: "idp2", Type: ComponentTypeIdP},
				{ID: "gateway", Type: ComponentTypeGateway},
			},
		},
	}

	idps := p.ComponentsByType(ComponentTypeIdP)
	if len(idps) != 2 {
		t.Errorf("ComponentsByType(idp) returned %d components, want 2", len(idps))
	}

	gateways := p.ComponentsByType(ComponentTypeGateway)
	if len(gateways) != 1 {
		t.Errorf("ComponentsByType(gateway) returned %d components, want 1", len(gateways))
	}
}

func TestProtocolEntitiesInComponent(t *testing.T) {
	p := &Protocol{
		Entities: []Entity{
			{ID: "as", Name: "Authorization Server"},
			{ID: "rs", Name: "Resource Server"},
			{ID: "client", Name: "Client"},
		},
		Metadata: &ProtocolMetadata{
			Components: []DeploymentComponent{
				{ID: "idp", Entities: []string{"as"}},
				{ID: "api", Entities: []string{"rs"}},
			},
		},
	}

	entities := p.EntitiesInComponent("idp")
	if len(entities) != 1 || entities[0].ID != "as" {
		t.Error("EntitiesInComponent should return correct entities")
	}

	entities = p.EntitiesInComponent("unknown")
	if entities != nil {
		t.Error("EntitiesInComponent should return nil for unknown component")
	}
}

func TestProtocolComponentForEntity(t *testing.T) {
	p := &Protocol{
		Metadata: &ProtocolMetadata{
			Components: []DeploymentComponent{
				{ID: "idp", Entities: []string{"as", "token_endpoint"}},
				{ID: "api", Entities: []string{"rs"}},
			},
		},
	}

	c := p.ComponentForEntity("as")
	if c == nil || c.ID != "idp" {
		t.Error("ComponentForEntity should find correct component")
	}

	c = p.ComponentForEntity("unknown")
	if c != nil {
		t.Error("ComponentForEntity should return nil for unknown entity")
	}
}

func TestProtocolTrustRelations(t *testing.T) {
	p := &Protocol{
		Metadata: &ProtocolMetadata{
			TrustRelations: []TrustRelationship{
				{ID: "tr1", From: "client", To: "idp", Type: TrustTypeTrusts},
				{ID: "tr2", From: "idp", To: "client", Type: TrustTypeAuthenticates},
				{ID: "tr3", From: "client", To: "api", Type: TrustTypeTrusts},
			},
		},
	}

	// Test TrustRelationByID
	tr := p.TrustRelationByID("tr1")
	if tr == nil || tr.From != "client" {
		t.Error("TrustRelationByID should find correct relation")
	}

	tr = p.TrustRelationByID("unknown")
	if tr != nil {
		t.Error("TrustRelationByID should return nil for unknown ID")
	}

	// Test TrustRelationsFrom
	fromClient := p.TrustRelationsFrom("client")
	if len(fromClient) != 2 {
		t.Errorf("TrustRelationsFrom(client) returned %d relations, want 2", len(fromClient))
	}

	// Test TrustRelationsTo
	toClient := p.TrustRelationsTo("client")
	if len(toClient) != 1 {
		t.Errorf("TrustRelationsTo(client) returned %d relations, want 1", len(toClient))
	}

	// Test TrustRelationsByType
	trusts := p.TrustRelationsByType(TrustTypeTrusts)
	if len(trusts) != 2 {
		t.Errorf("TrustRelationsByType(trusts) returned %d relations, want 2", len(trusts))
	}
}

func TestProtocolAllProtocols(t *testing.T) {
	p := &Protocol{
		Entities: []Entity{
			{ID: "e1", ProtocolRoles: []ProtocolRole{
				{Protocol: ProtocolOAuth, Role: "client"},
				{Protocol: ProtocolSCIM, Role: "service_provider"},
			}},
			{ID: "e2", ProtocolRoles: []ProtocolRole{
				{Protocol: ProtocolOAuth, Role: "authorization_server"},
			}},
		},
	}

	protocols := p.AllProtocols()
	if len(protocols) != 2 {
		t.Errorf("AllProtocols returned %d protocols, want 2", len(protocols))
	}
}

func TestProtocolAllComponentTypes(t *testing.T) {
	p := &Protocol{
		Metadata: &ProtocolMetadata{
			Components: []DeploymentComponent{
				{ID: "c1", Type: ComponentTypeIdP},
				{ID: "c2", Type: ComponentTypeIdP},
				{ID: "c3", Type: ComponentTypeGateway},
			},
		},
	}

	types := p.AllComponentTypes()
	if len(types) != 2 {
		t.Errorf("AllComponentTypes returned %d types, want 2", len(types))
	}
}

func TestProtocolAllTrustRelationTypes(t *testing.T) {
	p := &Protocol{
		Metadata: &ProtocolMetadata{
			TrustRelations: []TrustRelationship{
				{From: "a", To: "b", Type: TrustTypeTrusts},
				{From: "b", To: "c", Type: TrustTypeTrusts},
				{From: "c", To: "d", Type: TrustTypeAuthenticates},
			},
		},
	}

	types := p.AllTrustRelationTypes()
	if len(types) != 2 {
		t.Errorf("AllTrustRelationTypes returned %d types, want 2", len(types))
	}
}

func TestIsValidProtocol(t *testing.T) {
	validProtocols := []string{
		ProtocolOAuth, ProtocolSCIM, ProtocolSPIFFE, ProtocolAAuth,
		ProtocolIDJAG, ProtocolAuthZEN, ProtocolMCP, ProtocolA2A,
	}

	for _, p := range validProtocols {
		if !IsValidProtocol(p) {
			t.Errorf("IsValidProtocol(%q) = false, want true", p)
		}
	}

	if IsValidProtocol("invalid") {
		t.Error("IsValidProtocol(invalid) = true, want false")
	}
}

func TestIsValidComponentType(t *testing.T) {
	validTypes := []string{
		ComponentTypeIdP, ComponentTypeIGA, ComponentTypeAgentProvider,
		ComponentTypePersonServer, ComponentTypeAccessServer, ComponentTypePDP,
		ComponentTypeGateway, ComponentTypeMCPClient, ComponentTypeMCPServer,
		ComponentTypeResourceAPI, ComponentTypeSPIRE,
	}

	for _, ct := range validTypes {
		if !IsValidComponentType(ct) {
			t.Errorf("IsValidComponentType(%q) = false, want true", ct)
		}
	}

	if IsValidComponentType("invalid") {
		t.Error("IsValidComponentType(invalid) = true, want false")
	}
}

func TestIsValidTrustRelationType(t *testing.T) {
	validTypes := []string{
		TrustTypeAuthenticates, TrustTypeValidates, TrustTypeDelegates,
		TrustTypeAuthorizes, TrustTypeIssues, TrustTypeTrusts,
		TrustTypeProvisions, TrustTypeAttests,
	}

	for _, tt := range validTypes {
		if !IsValidTrustRelationType(tt) {
			t.Errorf("IsValidTrustRelationType(%q) = false, want true", tt)
		}
	}

	if IsValidTrustRelationType("invalid") {
		t.Error("IsValidTrustRelationType(invalid) = true, want false")
	}
}

func TestIsValidCredential(t *testing.T) {
	validCreds := []string{
		CredentialX509SVID, CredentialJWTSVID, CredentialJWTAssertion,
		CredentialAccessToken, CredentialIDToken, CredentialAAAgentJWT,
		CredentialAAAuthJWT, CredentialMTLS, CredentialAPIKey,
	}

	for _, c := range validCreds {
		if !IsValidCredential(c) {
			t.Errorf("IsValidCredential(%q) = false, want true", c)
		}
	}

	if IsValidCredential("invalid") {
		t.Error("IsValidCredential(invalid) = true, want false")
	}
}
