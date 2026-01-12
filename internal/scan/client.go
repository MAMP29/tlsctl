package scan

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
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

func PrepareRequest(ctx context.Context, dom string) (http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", ApiEntryPoint+"analyze", nil)
	if err != nil {
		slog.Error("Error al crear request", "err", err, "url", ApiEntryPoint)
		return http.Request{}, err
	}

	q := req.URL.Query()
	q.Add("host", dom)

	req.URL.RawQuery = q.Encode()
	req.Header.Set("email", DefaultEmail)
	return *req, nil
}

func Execute(req *http.Request) (int, ScanTask, error) {
	var scanTask ScanTask

	statusCode, err := FetchScanTask(*req, &scanTask)
	if err != nil {
		slog.Error("Error al momento de lanzar la ejecución", "err", err)
		return statusCode, ScanTask{}, err
	}

	switch statusCode {
	case 400:
		return fail(statusCode, "Parametros invalidos, revisar invocacion")
	case 441:
		return fail(statusCode, "Usuario no autorizado, problema al registrar el correo")
	case 429:
		slog.Warn("Demasiados intentos, atrasando peticion")
		return statusCode, ScanTask{Status: "pending"}, nil

	// USUARIO DEBE REINTENTAR
	case 503:
		return fail(statusCode, "Servicio no disponible, reintentar en 15 minutos")
	case 529:
		return fail(statusCode, "Servicio sobrecargado, reintentar en 30 minutos")
	}

	return statusCode, scanTask, nil

}

func FetchScanTask(request http.Request, scanTask *ScanTask) (int, error) {
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

func fail(code int, msg string) (int, ScanTask, error) {
	slog.Error(msg)
	return code, ScanTask{
		Status: "failed",
		Error:  &msg,
	}, nil
}
