package export

import (
	"bytes"
	"encoding/xml"
	"fmt"

	"github.com/grokify/pidl"
)

// BPMNExporter exports PIDL process specs to BPMN 2.0 XML format.
type BPMNExporter struct {
	// TargetNamespace for the BPMN definition.
	TargetNamespace string
}

// NewBPMNExporter creates a new BPMN exporter with defaults.
func NewBPMNExporter() *BPMNExporter {
	return &BPMNExporter{
		TargetNamespace: "http://pidl.dev/bpmn",
	}
}

// BPMN XML structures
type bpmnDefinitions struct {
	XMLName         xml.Name    `xml:"definitions"`
	XMLNS           string      `xml:"xmlns,attr"`
	BPMN            string      `xml:"xmlns:bpmn,attr"`
	BPMNDI          string      `xml:"xmlns:bpmndi,attr"`
	TargetNamespace string      `xml:"targetNamespace,attr"`
	ID              string      `xml:"id,attr"`
	Process         bpmnProcess `xml:"process"`
}

type bpmnProcess struct {
	ID              string        `xml:"id,attr"`
	Name            string        `xml:"name,attr"`
	IsExecutable    bool          `xml:"isExecutable,attr"`
	StartEvent      *bpmnEvent    `xml:"startEvent,omitempty"`
	EndEvent        *bpmnEvent    `xml:"endEvent,omitempty"`
	Tasks           []bpmnTask    `xml:"task,omitempty"`
	ServiceTasks    []bpmnTask    `xml:"serviceTask,omitempty"`
	UserTasks       []bpmnTask    `xml:"userTask,omitempty"`
	ParallelGateway []bpmnGateway `xml:"parallelGateway,omitempty"`
	SequenceFlows   []bpmnFlow    `xml:"sequenceFlow,omitempty"`
}

type bpmnEvent struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr,omitempty"`
}

type bpmnTask struct {
	ID       string `xml:"id,attr"`
	Name     string `xml:"name,attr"`
	Incoming string `xml:"incoming,omitempty"`
	Outgoing string `xml:"outgoing,omitempty"`
}

type bpmnGateway struct {
	ID       string `xml:"id,attr"`
	Name     string `xml:"name,attr,omitempty"`
	Incoming string `xml:"incoming,omitempty"`
	Outgoing string `xml:"outgoing,omitempty"`
}

type bpmnFlow struct {
	ID        string `xml:"id,attr"`
	SourceRef string `xml:"sourceRef,attr"`
	TargetRef string `xml:"targetRef,attr"`
}

// Export converts a PIDL protocol to BPMN 2.0 XML.
func (e *BPMNExporter) Export(p *pidl.Protocol) (string, error) {
	if p.ProtocolMeta.Kind != pidl.ProtocolKindProcess {
		return "", fmt.Errorf("BPMN export only supports process specifications")
	}

	definitions := e.buildDefinitions(p)
	return e.renderXML(definitions)
}

func (e *BPMNExporter) buildDefinitions(p *pidl.Protocol) *bpmnDefinitions {
	processID := "Process_" + p.ProtocolMeta.ID

	definitions := &bpmnDefinitions{
		XMLNS:           "http://www.omg.org/spec/BPMN/20100524/MODEL",
		BPMN:            "http://www.omg.org/spec/BPMN/20100524/MODEL",
		BPMNDI:          "http://www.omg.org/spec/BPMN/20100524/DI",
		TargetNamespace: e.TargetNamespace,
		ID:              "Definitions_" + p.ProtocolMeta.ID,
		Process: bpmnProcess{
			ID:           processID,
			Name:         p.ProtocolMeta.Name,
			IsExecutable: true,
			StartEvent: &bpmnEvent{
				ID:   "StartEvent_1",
				Name: "Start",
			},
			EndEvent: &bpmnEvent{
				ID:   "EndEvent_1",
				Name: "End",
			},
		},
	}

	// Convert entities to BPMN tasks
	var lastTaskID string
	for i, entity := range p.Entities {
		taskID := "Task_" + entity.ID

		var task bpmnTask
		switch entity.StepType {
		case pidl.StepTypeHuman:
			task = bpmnTask{
				ID:   taskID,
				Name: entity.Name,
			}
			definitions.Process.UserTasks = append(definitions.Process.UserTasks, task)
		case pidl.StepTypeExternal, pidl.StepTypeLLM, pidl.StepTypeTool:
			task = bpmnTask{
				ID:   taskID,
				Name: entity.Name,
			}
			definitions.Process.ServiceTasks = append(definitions.Process.ServiceTasks, task)
		default:
			task = bpmnTask{
				ID:   taskID,
				Name: entity.Name,
			}
			definitions.Process.Tasks = append(definitions.Process.Tasks, task)
		}

		// Add parallel gateways if needed
		if entity.Parallel != nil {
			gateway := bpmnGateway{
				ID:   "Gateway_" + entity.ID,
				Name: "Parallel " + entity.Name,
			}
			definitions.Process.ParallelGateway = append(definitions.Process.ParallelGateway, gateway)
		}

		// Connect tasks with sequence flows
		if i == 0 {
			// First task connects from start event
			flow := bpmnFlow{
				ID:        "Flow_start_to_" + entity.ID,
				SourceRef: "StartEvent_1",
				TargetRef: taskID,
			}
			definitions.Process.SequenceFlows = append(definitions.Process.SequenceFlows, flow)
		} else if lastTaskID != "" {
			flow := bpmnFlow{
				ID:        "Flow_" + lastTaskID + "_to_" + taskID,
				SourceRef: lastTaskID,
				TargetRef: taskID,
			}
			definitions.Process.SequenceFlows = append(definitions.Process.SequenceFlows, flow)
		}

		lastTaskID = taskID
	}

	// Connect last task to end event
	if lastTaskID != "" {
		flow := bpmnFlow{
			ID:        "Flow_" + lastTaskID + "_to_end",
			SourceRef: lastTaskID,
			TargetRef: "EndEvent_1",
		}
		definitions.Process.SequenceFlows = append(definitions.Process.SequenceFlows, flow)
	}

	return definitions
}

func (e *BPMNExporter) renderXML(definitions *bpmnDefinitions) (string, error) {
	var buf bytes.Buffer

	buf.WriteString(xml.Header)

	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")

	if err := encoder.Encode(definitions); err != nil {
		return "", fmt.Errorf("XML encode error: %w", err)
	}

	return buf.String(), nil
}
