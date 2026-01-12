package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"tlsctl/internal/scan"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: tlsctl scan [-w workers] <dominios...>")
		os.Exit(1)
	}

	scanCmd := flag.NewFlagSet("scan", flag.ExitOnError)
	workers := scanCmd.Int("w", 5, "Número de workers")

	switch os.Args[1] {
	case "scan":
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		scanCmd.Parse(os.Args[2:])

		domains := scan.ValidateDomains(scanCmd.Args())
		if len(domains) == 0 {
			slog.Error("No se especificaron dominios")
			return
		}

		numWorkers := *workers
		if numWorkers < 1 {
			numWorkers = 1
		}
		if numWorkers > 5 {
			numWorkers = 5
		}

		scan.LaunchPrincipalExecution(ctx, stop, domains, numWorkers)

	default:
		fmt.Println("Comando no reconocido")
		os.Exit(1)
	}
}
