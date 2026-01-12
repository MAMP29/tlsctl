package scan

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"
)

type WorkerSignal struct {
	WorkerID   int
	StatusCode int
	Domain     string
	Action     string
}

func LaunchPrincipalExecution() {

	generalBackground := context.Background()

	ctxInfo, cancelInfo := context.WithTimeout(generalBackground, 20*time.Second)
	defer cancelInfo()

	info := GetInfo(ctxInfo)

	fmt.Println("Obtención de la información")
	fmt.Println(info)

	ctxScan, cancelScan := context.WithTimeout(generalBackground, 10*time.Minute)
	defer cancelScan()

	const numWorkers = 2
	domains := []string{"ssllabs.com", "google.com", "bubble.io"}

	jobs := make(chan string, len(domains))
	results := make(chan ScanTask, len(domains))
	signals := make(chan WorkerSignal, numWorkers*2)
	var wg sync.WaitGroup

	for i := range numWorkers {
		wg.Add(1)
		go worker(i, ctxScan, jobs, results, signals, &wg)
	}

	go func() {
		ticker := time.NewTicker(time.Duration(info.NewAssessmentCoolOff * int(time.Millisecond)))
		defer ticker.Stop()

		for _, domain := range domains {
			<-ticker.C
			jobs <- domain
			slog.Info("--> Dominio enviado al pool", "dominio", domain)
		}
		close(jobs)

	}()

	go func() {
		for signal := range signals {
			switch signal.Action {
			case "rate_limit":
				slog.Warn("Rate limit detectado", "worker", signal.WorkerID, "domain", signal.Domain)
			case "service_down":
				slog.Error("Servicio caido", "code", signal.StatusCode)
				// TODO: Hacer CloseAll() para cerrar todo debido a problemas con el servidor
			}
		}
	}()

	wg.Wait()
	close(results)
	close(signals)

	var finalResults []ScanTask
	for result := range results {

		finalResults = append(finalResults, result)
		if result.Status == "failed" {
			slog.Warn("Dominio falló", "domain", result.Host, "error", *result.Error)
		} else {
			//slog.Info("Dominio completado", "domain", result.Host, "grade", result.Endpoints[0].Grade)
			fmt.Println(result)
		}
	}

	err := SaveResults(finalResults)
	if err != nil {
		slog.Error("Error durante el guardado de los resultados", "err", err)
	}
}

func worker(id int, ctx context.Context, jobs <-chan string,
	results chan<- ScanTask,
	signals chan<- WorkerSignal,
	wg *sync.WaitGroup) {
	defer wg.Done()

	for domain := range jobs {
		slog.Info("Escaneo inicializado", "Worker", id, "Dominio", domain)

		req, err := PrepareRequest(ctx, domain)
		if err != nil {
			continue
		}

		var retries429, retries500 int

		// Etiqueta de manejo de polling
	PollingLoop:
		for {
			status, result, err := Execute(&req)

			if err != nil {
				slog.Error("Error de red", "worker", id, "dom", domain, "err", err)
				break PollingLoop
			}

			switch status {
			case 429:
				retries429++
				if retries429 > 3 {
					slog.Error("Limite 429 excedido", "dom", domain)
					break PollingLoop
				}
				signals <- WorkerSignal{WorkerID: id, StatusCode: 429, Action: "rate_limit"}
				if !waitContext(ctx, time.Duration(30*retries429)*time.Second) {
					return
				}
				continue

			case 500:
				retries500++
				if retries500 > 4 {
					break PollingLoop
				}
				if !waitContext(ctx, 5*time.Second) {
					return
				}
				continue

			case 503, 529:
				signals <- WorkerSignal{WorkerID: id, StatusCode: status, Action: "service_down"}
				return
			}

			switch result.Status {
			case "READY", "ERROR":
				results <- result
				break PollingLoop

			case "IN_PROGRESS", "pending":
				slog.Debug("Esperando progreso...", "worker", id, "dom", domain)
				wait := 10 * time.Second
				jitter := time.Duration(rand.Float64() * float64(wait) * 0.2)
				if !waitContext(ctx, wait+jitter) {
					return
				}
			}
		}
	}
}

func waitContext(ctx context.Context, d time.Duration) bool {
	select {
	case <-time.After(d):
		return true
	case <-ctx.Done():
		return false
	}
}
