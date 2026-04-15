package services

import (
	"fmt"
	"strings"

	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/utils_oas/request"
)

func CrearObservacion(baseCrud string, historicoEstadoSolicitudId int, descripcion string) (int, error) {
	if historicoEstadoSolicitudId <= 0 {
		return 0, fmt.Errorf("historicoEstadoSolicitudId es obligatorio")
	}

	descripcion = strings.TrimSpace(descripcion)
	if descripcion == "" {
		return 0, fmt.Errorf("la descripcion de la observación es obligatoria")
	}

	postURL := helpers.JoinURL(baseCrud, "/observacion")
	if err := helpers.ValidateAbsoluteURL(postURL); err != nil {
		return 0, err
	}

	payload := map[string]interface{}{
		"HistoricoEstadoSolicitudId": map[string]interface{}{"Id": historicoEstadoSolicitudId},
		"Descripcion":                descripcion,
		"Activo":                     true,
	}

	var postResp map[string]interface{}
	if err := request.SendJson(postURL, "POST", &postResp, payload); err != nil {
		return 0, fmt.Errorf("error creando observación: %v", err)
	}

	observacionId := helpers.ExtractIdAtoi(postResp)
	if observacionId <= 0 {
		return 0, fmt.Errorf("se creó la observación pero no se pudo extraer su Id")
	}

	return observacionId, nil
}
