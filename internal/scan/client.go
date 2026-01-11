package scan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"time"
)

type HTTPError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("status %d: %s (internal: %v)", e.StatusCode, e.Message, e.Err)
}

type AssessmentsStatus struct {
	MaxAssessments     int
	CurrentAssessments int
}

func (as *AssessmentsStatus) String() string {
	return fmt.Sprintf("Número máximo de peticiones: %d, Número actual de petciones: %d", as.MaxAssessments, as.CurrentAssessments)
}

func ScanLogic(ctx context.Context, dom string) (int, ScanTask, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", ApiEntryPoint+"analyze", nil)
	if err != nil {
		slog.Error("Error al crear request", "err", err, "url", ApiEntryPoint)
		return 0, ScanTask{}, err
	}
	q := req.URL.Query()
	q.Add("host", dom)

	req.URL.RawQuery = q.Encode()
	req.Header.Set("email", DefaultEmail)

	var scanTask ScanTask
	var limit500 int
	var attemptCounter int

	for {
		select {
		case <-ctx.Done():
			return 0, ScanTask{}, errors.New("Timeout: El escaneo de los endpoints ha superado el limite de tiempo")
		default:
			attemptCounter++
			wait := 5 * time.Second
			slog.Info("Intento numero: ", "intento", attemptCounter)
			statusCode, err := LaunchExecution(*req, &scanTask)
			if err != nil {
				slog.Error("Error al momento de lanzar la ejecución", "err", err)
				return statusCode, ScanTask{}, err
			}

			switch statusCode {
			case 400:
				slog.Error("Parametros invalidos, revisar invocación")
				errMsg := "Parametros invalidos, revisar invocacion"
				return statusCode, ScanTask{Status: "failed", Error: &errMsg}, nil
			case 441:
				slog.Error("Usuario no autorizado, problema al registrar el correo")
				errMsg := "Usuario no autorizado, problema al registear el correo"
				return statusCode, ScanTask{Status: "failed", Error: &errMsg}, nil
			case 429:
				slog.Warn("Demasiados intentos, atrasando peticion")
				return statusCode, ScanTask{Status: "pending"}, nil
			case 500:
				limit500++
				slog.Info("Error interno del servicio, reitentando")
				if limit500 > 4 {
					slog.Warn("Limite de reintentos sobrepasado, cerrando petición")
					errMsg := "Error interno del servicio"
					return statusCode, ScanTask{Status: "failed", Error: &errMsg}, nil
				}
				continue

			// USUARIO DEBE REINTENTAR
			case 503:
				slog.Warn("Servicio no disponible por el momento, reintentar en 15 minutos")
				errMsg := "Servicio no disponible, error 503"
				return statusCode, ScanTask{Status: "failed", Error: &errMsg}, nil
			case 529:
				slog.Warn("Servicio sobrecargado, reintentar en 30 minutos")
				errMsg := "Servicio sobrecargado, error 529"
				return statusCode, ScanTask{Status: "failed", Error: &errMsg}, nil
			}

			status := scanTask.Status
			slog.Debug("Estado del progreso en la API: ", "estado", status)

			switch status {
			case "READY":
				slog.Debug("Estado de respuesta: ", "estado", statusCode)
				return statusCode, scanTask, nil
			case "ERROR":
				slog.Debug("Error del cliente al momento de procesar el pedido: ", "error", scanTask.Error)
				return statusCode, scanTask, nil
			case "IN_PROGRESS":
				slog.Debug("Esperando a que termine el proceso") // No es factible imprimir el progreso directo del scanTask porque el progreso real se visualiza en los endpoints
				wait = 10 * time.Second
			}

			jitter := time.Duration(rand.Float64() * float64(wait) * 0.2)
			time.Sleep(wait + jitter)
		}
	}
}

func LaunchExecution(request http.Request, scanTask *ScanTask) (int, error) {
	res, err := http.DefaultClient.Do(&request)
	if err != nil {
		slog.Error("Error de red al momento de hacer la petición", "err", err, "url", request.URL)
		return 0, err
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(scanTask)
	if err != nil {
		slog.Error("Error al decodificar el JSON de tareas", "err", err)
		return 0, err // Aunque la petición haya sido exitosa, no deberia de actuar ante un statusCode particular, pues el manejo del error radica en el area de errores a nivel de la aplicación. Por eso devuelve 0.
	}
	return res.StatusCode, nil
}
