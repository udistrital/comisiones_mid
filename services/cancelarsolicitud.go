package services

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/utils_oas/request"
)

func CancelarSolicitud(solicitudId int) (models.CancelarSolicitudResponse, error) {
	if solicitudId <= 0 {
		return models.CancelarSolicitudResponse{}, fmt.Errorf("solicitudId es obligatorio")
	}

	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	logs.Info("UrlComisionesCrud=%q", baseCrud)
	if baseCrud == "" {
		return models.CancelarSolicitudResponse{}, fmt.Errorf("no está configurado UrlComisionesCrud")
	}

	resultado := models.CancelarSolicitudResponse{
		SolicitudId:                     solicitudId,
		SolicitudDesactivada:            false,
		DetalleSolicitudDesactivados:    []int{},
		HistoricosDesactivados:          []int{},
		ObservacionesDesactivadas:       []int{},
		DocumentosSolicitudDesactivados: []int{},
	}

	if err := desactivarSolicitud(baseCrud, solicitudId); err != nil {
		return models.CancelarSolicitudResponse{}, fmt.Errorf("error desactivando solicitud: %v", err)
	}
	resultado.SolicitudDesactivada = true

	detalleIds, err := obtenerIdsPorQuery(
		baseCrud,
		"detalle_solicitud",
		fmt.Sprintf("SolicitudId:%d,Activo:true", solicitudId),
	)
	if err != nil {
		return models.CancelarSolicitudResponse{}, fmt.Errorf("error consultando detalle_solicitud: %v", err)
	}

	for _, id := range detalleIds {
		if err := desactivarRecursoPorId(baseCrud, "detalle_solicitud", id); err != nil {
			return models.CancelarSolicitudResponse{}, fmt.Errorf("error desactivando detalle_solicitud %d: %v", id, err)
		}
	}
	resultado.DetalleSolicitudDesactivados = detalleIds

	historicoIds, err := obtenerIdsPorQuery(
		baseCrud,
		"historico_estado_solicitud",
		fmt.Sprintf("SolicitudId:%d,Activo:true", solicitudId),
	)
	if err != nil {
		return models.CancelarSolicitudResponse{}, fmt.Errorf("error consultando históricos de solicitud: %v", err)
	}

	for _, historicoId := range historicoIds {
		// Desactivar observaciones asociadas al histórico
		observacionIds, err := obtenerIdsPorQuery(
			baseCrud,
			"observacion",
			fmt.Sprintf("HistoricoEstadoSolicitudId:%d,Activo:true", historicoId),
		)
		if err != nil {
			return models.CancelarSolicitudResponse{}, fmt.Errorf("error consultando observaciones del histórico %d: %v", historicoId, err)
		}

		for _, obsId := range observacionIds {
			if err := desactivarRecursoPorId(baseCrud, "observacion", obsId); err != nil {
				return models.CancelarSolicitudResponse{}, fmt.Errorf("error desactivando observación %d: %v", obsId, err)
			}
		}
		resultado.ObservacionesDesactivadas = append(resultado.ObservacionesDesactivadas, observacionIds...)

		// Desactivar documentos asociados al histórico
		documentoIds, err := obtenerIdsPorQuery(
			baseCrud,
			"documento_solicitud",
			fmt.Sprintf("HistoricoEstadoSolicitudId:%d,Activo:true", historicoId),
		)
		if err != nil {
			return models.CancelarSolicitudResponse{}, fmt.Errorf("error consultando documentos del histórico %d: %v", historicoId, err)
		}

		for _, docId := range documentoIds {
			if err := desactivarRecursoPorId(baseCrud, "documento_solicitud", docId); err != nil {
				return models.CancelarSolicitudResponse{}, fmt.Errorf("error desactivando documento_solicitud %d: %v", docId, err)
			}
		}
		resultado.DocumentosSolicitudDesactivados = append(resultado.DocumentosSolicitudDesactivados, documentoIds...)

		if err := desactivarHistorico(baseCrud, historicoId); err != nil {
			return models.CancelarSolicitudResponse{}, fmt.Errorf("error desactivando histórico %d: %v", historicoId, err)
		}
		resultado.HistoricosDesactivados = append(resultado.HistoricosDesactivados, historicoId)
	}

	return resultado, nil
}

func desactivarSolicitud(baseCrud string, solicitudId int) error {
	getURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/solicitud/%d", solicitudId))
	if err := helpers.ValidateAbsoluteURL(getURL); err != nil {
		return err
	}

	var getResp map[string]interface{}
	err := request.GetJson(getURL, &getResp)
	if err != nil {
		return fmt.Errorf("error GET solicitud: %v", err)
	}

	obj := helpers.UnwrapDataToMap(getResp)
	if obj == nil {
		return fmt.Errorf("respuesta inválida al GET de la solicitud %d: %v", solicitudId, getResp)
	}

	obj["Activo"] = false

	var putResp map[string]interface{}
	err = request.SendJson(getURL, "PUT", &putResp, obj)
	if err != nil {
		return fmt.Errorf("error PUT solicitud: %v", err)
	}

	return nil
}

func desactivarRecursoPorId(baseCrud, recurso string, id int) error {
	getURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/%s/%d", recurso, id))
	if err := helpers.ValidateAbsoluteURL(getURL); err != nil {
		return err
	}

	var getResp map[string]interface{}
	err := request.GetJson(getURL, &getResp)
	if err != nil {
		return fmt.Errorf("error GET %s: %v", recurso, err)
	}

	obj := helpers.UnwrapDataToMap(getResp)
	if obj == nil {
		return fmt.Errorf("respuesta inválida al GET de %s %d: %v", recurso, id, getResp)
	}

	obj["Activo"] = false

	var putResp map[string]interface{}
	err = request.SendJson(getURL, "PUT", &putResp, obj)
	if err != nil {
		return fmt.Errorf("error PUT %s: %v", recurso, err)
	}

	return nil
}

func obtenerIdsPorQuery(baseCrud, recurso, query string) ([]int, error) {
	u, err := url.Parse(helpers.JoinURL(baseCrud, "/"+recurso))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("query", query)
	q.Set("limit", "0")
	u.RawQuery = q.Encode()

	finalURL := u.String()
	logs.Info("URL consulta %s = %s", recurso, finalURL)

	var envelope map[string]interface{}
	err = request.GetJson(finalURL, &envelope)
	if err != nil {
		return nil, fmt.Errorf("error consultando %s: %v", recurso, err)
	}

	raw := envelope["Data"]
	if raw == nil {
		return []int{}, nil
	}

	arr, ok := raw.([]interface{})
	if !ok || len(arr) == 0 {
		return []int{}, nil
	}

	ids := make([]int, 0, len(arr))
	for _, item := range arr {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		id, err := strconv.Atoi(fmt.Sprintf("%v", row["Id"]))
		if err == nil && id > 0 {
			ids = append(ids, id)
		}
	}

	return ids, nil
}
