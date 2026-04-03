// Package helpdesk provides a help-desk skill that reads supplier data from the
// database via sqlx, detects and prioritises data incidents, and drafts
// structured responses to compliance and performance questionnaires.
package helpdesk

// GetIncidents returns all detected, prioritised incidents.
func GetIncidents() ([]Incident, error) {
	return DetectAndPrioritise()
}

// GetQuestionnaireDraft returns a fully drafted questionnaire response.
// incidentID == 0 means "use all incidents / generate a global response".
func GetQuestionnaireDraft(qType QuestionnaireType, incidentID int) (DraftResponse, error) {
	incidents, err := DetectAndPrioritise()
	if err != nil {
		return DraftResponse{}, err
	}

	var target *Incident
	if incidentID > 0 {
		for i := range incidents {
			if incidents[i].ID == incidentID {
				target = &incidents[i]
				break
			}
		}
	}

	return DraftQuestionnaire(DraftRequest{Type: qType, Incident: target}, incidents), nil
}
