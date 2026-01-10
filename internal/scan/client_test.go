package scan

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestScanLogic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)

	defer cancel()

	responseCode, scanTask, err := ScanLogic(ctx, "ssllabs.com")
	if err != nil {
		t.Fatalf("Hubo un error grave al procesar el escaneo %v\n", err)
	}

	if responseCode != 200 {
		t.Error("No fue una respusta exitosa")
	}

	fmt.Println(scanTask)
}
