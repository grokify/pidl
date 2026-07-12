package pidl

// AnalyzeDataLineage extracts the data lineage graph from a process protocol.
// It traces data flow from sources through transformations to sinks.
func AnalyzeDataLineage(p *Protocol) *DataLineage {
	lineage := &DataLineage{
		ProtocolID: p.ProtocolMeta.ID,
		Edges:      []LineageEdge{},
		Sources:    []PortReference{},
		Sinks:      []PortReference{},
	}

	// Build entity map for quick lookup
	entityMap := make(map[string]*Entity)
	for i := range p.Entities {
		entityMap[p.Entities[i].ID] = &p.Entities[i]
	}

	// Track which ports have incoming/outgoing connections
	hasIncoming := make(map[string]bool) // "entity:port"
	hasOutgoing := make(map[string]bool)

	// Process flows to extract lineage edges
	for i, flow := range p.Flows {
		fromEntity := entityMap[flow.From]
		toEntity := entityMap[flow.To]

		if fromEntity == nil || toEntity == nil {
			continue
		}

		// If explicit data mappings exist, use them
		if len(flow.DataMappings) > 0 {
			for _, mapping := range flow.DataMappings {
				edge := LineageEdge{
					SourceEntity:   flow.From,
					SourcePort:     mapping.OutputPort,
					TargetEntity:   flow.To,
					TargetPort:     mapping.InputPort,
					FlowIndex:      i,
					Transformation: mapping.Transformation,
				}
				lineage.Edges = append(lineage.Edges, edge)

				hasOutgoing[flow.From+":"+mapping.OutputPort] = true
				hasIncoming[flow.To+":"+mapping.InputPort] = true
			}
		} else {
			// Infer connections from output/input port names
			lineage.Edges = append(lineage.Edges, inferPortConnections(fromEntity, toEntity, i, hasOutgoing, hasIncoming)...)
		}
	}

	// Identify sources (output ports with no incoming data)
	for _, entity := range p.Entities {
		for _, output := range entity.Outputs {
			key := entity.ID + ":" + output.Name
			if !hasIncoming[key] {
				lineage.Sources = append(lineage.Sources, PortReference{
					EntityID:  entity.ID,
					PortName:  output.Name,
					PortKind:  "output",
					Sensitive: output.Sensitive,
				})
			}
		}
	}

	// Identify sinks (input ports with no outgoing data)
	for _, entity := range p.Entities {
		for _, input := range entity.Inputs {
			key := entity.ID + ":" + input.Name
			if !hasOutgoing[key] {
				lineage.Sinks = append(lineage.Sinks, PortReference{
					EntityID:  entity.ID,
					PortName:  input.Name,
					PortKind:  "input",
					Sensitive: input.Sensitive,
				})
			}
		}
	}

	// Trace sensitive data paths
	lineage.SensitiveDataPaths = traceSensitivePaths(lineage, entityMap)

	return lineage
}

// inferPortConnections attempts to match output ports to input ports by name.
func inferPortConnections(from, to *Entity, flowIndex int, hasOutgoing, hasIncoming map[string]bool) []LineageEdge {
	var edges []LineageEdge

	// Build input port name map for target entity
	inputPorts := make(map[string]*DataPort)
	for i := range to.Inputs {
		inputPorts[to.Inputs[i].Name] = &to.Inputs[i]
	}

	// Try to match output ports to input ports
	for _, output := range from.Outputs {
		// Exact name match
		if input, ok := inputPorts[output.Name]; ok {
			edge := LineageEdge{
				SourceEntity: from.ID,
				SourcePort:   output.Name,
				TargetEntity: to.ID,
				TargetPort:   input.Name,
				FlowIndex:    flowIndex,
			}
			edges = append(edges, edge)

			hasOutgoing[from.ID+":"+output.Name] = true
			hasIncoming[to.ID+":"+input.Name] = true
		}
	}

	return edges
}

// traceSensitivePaths finds all paths that carry sensitive data.
func traceSensitivePaths(lineage *DataLineage, entityMap map[string]*Entity) [][]PortReference {
	var paths [][]PortReference

	// Find all sensitive source ports
	var sensitiveStarts []PortReference
	for _, source := range lineage.Sources {
		if source.Sensitive {
			sensitiveStarts = append(sensitiveStarts, source)
		}
	}

	// Also check entities with sensitive outputs that aren't pure sources
	for _, entity := range entityMap {
		for _, output := range entity.Outputs {
			if output.Sensitive {
				ref := PortReference{
					EntityID:  entity.ID,
					PortName:  output.Name,
					PortKind:  "output",
					Sensitive: true,
				}
				// Check if not already in sensitiveStarts
				found := false
				for _, s := range sensitiveStarts {
					if s.EntityID == ref.EntityID && s.PortName == ref.PortName {
						found = true
						break
					}
				}
				if !found {
					sensitiveStarts = append(sensitiveStarts, ref)
				}
			}
		}
	}

	// Build adjacency map from edges
	adjacency := make(map[string][]LineageEdge)
	for _, edge := range lineage.Edges {
		key := edge.SourceEntity + ":" + edge.SourcePort
		adjacency[key] = append(adjacency[key], edge)
	}

	// DFS from each sensitive start
	for _, start := range sensitiveStarts {
		visited := make(map[string]bool)
		var path []PortReference
		tracePath(start, adjacency, visited, path, &paths)
	}

	return paths
}

// tracePath performs DFS to find all paths from a sensitive source.
func tracePath(current PortReference, adjacency map[string][]LineageEdge, visited map[string]bool, path []PortReference, paths *[][]PortReference) {
	key := current.EntityID + ":" + current.PortName
	if visited[key] {
		return
	}
	visited[key] = true
	path = append(path, current)

	// Find outgoing edges
	edges := adjacency[key]
	if len(edges) == 0 {
		// End of path - store it
		if len(path) > 1 {
			pathCopy := make([]PortReference, len(path))
			copy(pathCopy, path)
			*paths = append(*paths, pathCopy)
		}
	} else {
		for _, edge := range edges {
			next := PortReference{
				EntityID: edge.TargetEntity,
				PortName: edge.TargetPort,
				PortKind: "input",
			}
			tracePath(next, adjacency, visited, path, paths)
		}
	}
}

// GetUpstream returns all ports that feed data into the specified port.
func (l *DataLineage) GetUpstream(entityID, portName string) []PortReference {
	var upstream []PortReference
	for _, edge := range l.Edges {
		if edge.TargetEntity == entityID && edge.TargetPort == portName {
			upstream = append(upstream, PortReference{
				EntityID: edge.SourceEntity,
				PortName: edge.SourcePort,
				PortKind: "output",
			})
		}
	}
	return upstream
}

// GetDownstream returns all ports that receive data from the specified port.
func (l *DataLineage) GetDownstream(entityID, portName string) []PortReference {
	var downstream []PortReference
	for _, edge := range l.Edges {
		if edge.SourceEntity == entityID && edge.SourcePort == portName {
			downstream = append(downstream, PortReference{
				EntityID: edge.TargetEntity,
				PortName: edge.TargetPort,
				PortKind: "input",
			})
		}
	}
	return downstream
}

// GetImpactedEntities returns all entities that would be affected if
// the specified entity's output changes (downstream impact analysis).
func (l *DataLineage) GetImpactedEntities(entityID string) []string {
	impacted := make(map[string]bool)
	toVisit := []string{entityID}

	for len(toVisit) > 0 {
		current := toVisit[0]
		toVisit = toVisit[1:]

		for _, edge := range l.Edges {
			if edge.SourceEntity == current && !impacted[edge.TargetEntity] {
				impacted[edge.TargetEntity] = true
				toVisit = append(toVisit, edge.TargetEntity)
			}
		}
	}

	result := make([]string, 0, len(impacted))
	for id := range impacted {
		result = append(result, id)
	}
	return result
}

// GetDataProvenance returns all source entities that contribute to
// the specified entity's inputs (upstream provenance analysis).
func (l *DataLineage) GetDataProvenance(entityID string) []string {
	provenance := make(map[string]bool)
	toVisit := []string{entityID}

	for len(toVisit) > 0 {
		current := toVisit[0]
		toVisit = toVisit[1:]

		for _, edge := range l.Edges {
			if edge.TargetEntity == current && !provenance[edge.SourceEntity] {
				provenance[edge.SourceEntity] = true
				toVisit = append(toVisit, edge.SourceEntity)
			}
		}
	}

	result := make([]string, 0, len(provenance))
	for id := range provenance {
		result = append(result, id)
	}
	return result
}

// HasSensitiveDataFlow returns true if there's any sensitive data flowing
// through the lineage graph.
func (l *DataLineage) HasSensitiveDataFlow() bool {
	return len(l.SensitiveDataPaths) > 0
}
