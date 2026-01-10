package scan

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type Info struct {
	EngineVersion        string `json:"engineVersion"`
	CriteriaVersion      string `json:"criteriaVersion"`
	MaxAssessments       int    `json:"maxAssessments"`       // Numero maximo de registros permitidos
	CurrentAssessments   int    `json:"currentAssessments"`   // Registros activos
	NewAssessmentCoolOff int    `json:"newAssessmentCoolOff"` // Cooldown entre pedidos
}

func (i Info) String() string {
	return fmt.Sprintf(
		"Info{EngineVersion:%q, CriteriaVersion:%q, MaxAssessments:%d, CurrentAssessments:%d, NewAssessmentCoolOff:%d}",
		i.EngineVersion,
		i.CriteriaVersion,
		i.MaxAssessments,
		i.CurrentAssessments,
		i.NewAssessmentCoolOff,
	)
}

func GetInfo(ctx context.Context) Info {
	var info Info
	req, err := http.NewRequestWithContext(ctx, "GET", ApiEntryPoint+"info", nil)
	if err != nil {
		slog.Error("Error al crear request", "err", err, "url", ApiEntryPoint)
		return Info{}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Error en la llamada HTTP", "err", err)
		return Info{}
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&info)
	if err != nil {
		slog.Error("Error al decodificar JSON", "err", err)
		return Info{}
	}
	return info
}
