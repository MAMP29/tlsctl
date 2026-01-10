package scan

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

func LaunchPrincipalExecution() {

	generalBackground := context.Background()

	ctxInfo, cancelInfo := context.WithTimeout(generalBackground, 20*time.Second)
	defer cancelInfo()

	info := GetInfo(ctxInfo)

	fmt.Println("Obtención de la información")
	fmt.Println(info)

	ctxScan, cancelScan := context.WithTimeout(generalBackground, 3*time.Minute)
	defer cancelScan()

	const numWorkers = 2
	domains := []string{"ssllabs.com", "google.com", "bubble.io"}

	jobs := make(chan string, len(domains))
	results := make(chan ScanTask, len(domains))
	var wg sync.WaitGroup

	for i := range numWorkers {
		wg.Add(1)
		go worker(i, ctxScan, jobs, results, &wg)
	}

	for _, domain := range domains {
		jobs <- domain
	}
	close(jobs)

	wg.Wait()
	close(results)

	for results := range results {
		fmt.Println(results)
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

func worker(id int, ctx context.Context, jobs <-chan string, results chan<- ScanTask, wg *sync.WaitGroup) {
	defer wg.Done()
	for domain := range jobs {
		slog.Info("Escaneo inicializado", "Worker", id, "Dominio", domain)
		status, result, err := ScanLogic(ctx, domain)
		if err != nil {
			slog.Error("Error grave al procesar el escaneo", "error", err)
			continue
		}
		fmt.Printf("Estatus: %d\n", status)
		results <- result
	}
}
