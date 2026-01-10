package models

import "fmt"

type InfoInit struct {
	EngineVersion        string `json:"engineVersion"`
	CriteriaVersion      string `json:"criteriaVersion"`
	MaxAssessments       int    `json:"maxAssessments"`       // Numero maximo de registros permitidos
	CurrentAssessments   int    `json:"currentAssessments"`   // Registros activos
	NewAssessmentCoolOff int    `json:"newAssessmentCoolOff"` // Cooldown entre pedidos
}

func (i InfoInit) String() string {
	return fmt.Sprintf(
		"InfoInit{EngineVersion:%q, CriteriaVersion:%q, MaxAssessments:%d, CurrentAssessments:%d, NewAssessmentCoolOff:%d}",
		i.EngineVersion,
		i.CriteriaVersion,
		i.MaxAssessments,
		i.CurrentAssessments,
		i.NewAssessmentCoolOff,
	)
}
