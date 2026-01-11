package scan

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"time"
	"tlsctl/internal/project"
)

func SaveResults(results []ScanTask) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := "resultado_" + timestamp + ".json"

	root, err := project.Root()
	if err != nil {
		slog.Error("Hubo un error encontrando go.mod", "err", err)
		return err
	}

	resultsDir := filepath.Join(root, "results")

	err = os.MkdirAll(resultsDir, 0755)
	if err != nil {
		slog.Error("Hubo un error durante la creación de la ruta", "err", err)
		return err
	}

	filePath := filepath.Join(resultsDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		slog.Error("Hubo un error critico durante la creación del archivo de resultados", "err", err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")

	if err := encoder.Encode(results); err != nil {
		slog.Error("Error critico durante la escritura de los resultados a JSON", "err", err)
		return err
	}

	return nil
}
