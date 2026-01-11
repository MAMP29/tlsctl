package scan

import (
	"context"
	"fmt"
	"log/slog"
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

	/*
		 	baseDomain := "ssllabs.com"

			statusCode, scanTask, err := ScanLogic(ctxScan, baseDomain)
			if err != nil {
				fmt.Printf("Hubo un error al procesar el escaneo %s\n", err)
			}
			fmt.Printf("Código de estado final %d\n", statusCode)

			fmt.Println("Tarea escaneada:")
			fmt.Println(scanTask)
	*/
}

func worker(id int, ctx context.Context, jobs <-chan string,
	results chan<- ScanTask,
	signals chan<- WorkerSignal,
	wg *sync.WaitGroup) {
	defer wg.Done()
	for domain := range jobs {
		slog.Info("Escaneo inicializado", "Worker", id, "Dominio", domain)
		status, result, err := ScanLogic(ctx, domain)

		if status == 429 {
			signals <- WorkerSignal{
				WorkerID:   id,
				StatusCode: 429,
				Domain:     domain,
				Action:     "rate_limit",
			}

			time.Sleep(30 * time.Second)
			continue
		}

		if status == 503 || status == 529 {
			signals <- WorkerSignal{
				WorkerID:   id,
				StatusCode: status,
				Domain:     domain,
				Action:     "service_down",
			}
			return
		}

		if err != nil {
			slog.Error("Error grave al procesar el escaneo", "domain", domain, "error", err)
			continue
		}
		fmt.Printf("Estatus: %d\n", status)
		results <- result
	}
}
