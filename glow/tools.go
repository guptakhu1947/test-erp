package glow

import (
	"encoding/json"
	"fmt"

	"erp-app/bizlogic"
	helpdesk "erp-app/help-desk-skill"
)

// ToolDefinition is the schema sent to Claude.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// toolResult is returned after executing a tool.
type toolResult struct {
	ToolUseID string `json:"tool_use_id"`
	Type      string `json:"type"`
	Content   string `json:"content"`
}

// ToolUseDisplay is returned to the frontend so it can show what Glow did.
type ToolUseDisplay struct {
	Name    string `json:"name"`
	Display string `json:"display"`
}

var tools = []ToolDefinition{
	{
		Name:        "get_supplier_overview",
		Description: "Retrieve all suppliers with their total order count, invoice count, and total spend. Use this to answer questions about supplier relationships, spend rankings, or portfolio overview.",
		InputSchema: json.RawMessage(`{"type":"object","properties":{},"required":[]}`),
	},
	{
		Name:        "get_incidents",
		Description: "Detect and return all prioritised data incidents from the supplier database, ranked by severity (Critical → High → Medium → Low). Each incident includes root cause category, affected entity, evidence, and recommended action. Use this for risk analysis, audit responses, or when asked about problems.",
		InputSchema: json.RawMessage(`{"type":"object","properties":{},"required":[]}`),
	},
	{
		Name:        "get_questionnaire_draft",
		Description: "Generate a full structured questionnaire response draft. Returns sections with Q&A pairs.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"type":{
					"type":"string",
					"enum":["root_cause_analysis","incident_report","supplier_performance","compliance_audit"],
					"description":"The type of questionnaire to draft"
				},
				"incident_id":{
					"type":"integer",
					"description":"Optional: scope the draft to a specific incident ID (0 = global across all incidents)"
				}
			},
			"required":["type"]
		}`),
	},
	{
		Name:        "get_insights",
		Description: "Retrieve spend trends (monthly), spend by supplier, spend by country, order/invoice status breakdowns, and strategic recommendations. Use this for trend analysis, financial reviews, or strategic questions.",
		InputSchema: json.RawMessage(`{"type":"object","properties":{},"required":[]}`),
	},
}

// executeTool runs the named tool and returns (result string, display label, error).
func executeTool(name string, inputRaw json.RawMessage) (string, string, error) {
	switch name {

	case "get_supplier_overview":
		suppliers, err := bizlogic.GetAllSuppliers()
		if err != nil {
			return "", "", err
		}
		b, _ := json.MarshalIndent(suppliers, "", "  ")
		return string(b), "Retrieved supplier overview", nil

	case "get_incidents":
		incidents, err := helpdesk.GetIncidents()
		if err != nil {
			return "", "", err
		}
		b, _ := json.MarshalIndent(incidents, "", "  ")
		return string(b), "Scanned incident database", nil

	case "get_questionnaire_draft":
		var input struct {
			Type       string `json:"type"`
			IncidentID int    `json:"incident_id"`
		}
		if err := json.Unmarshal(inputRaw, &input); err != nil {
			return "", "", fmt.Errorf("invalid input: %w", err)
		}
		draft, err := helpdesk.GetQuestionnaireDraft(helpdesk.QuestionnaireType(input.Type), input.IncidentID)
		if err != nil {
			return "", "", err
		}
		b, _ := json.MarshalIndent(draft, "", "  ")
		return string(b), fmt.Sprintf("Drafted %s questionnaire", input.Type), nil

	case "get_insights":
		dashboard, err := bizlogic.GetInsightsDashboard()
		if err != nil {
			return "", "", err
		}
		b, _ := json.MarshalIndent(dashboard, "", "  ")
		return string(b), "Analysed spend insights", nil

	default:
		return "", "", fmt.Errorf("unknown tool: %s", name)
	}
}
