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

func CambiarEstadoSolicitud(solicitudId int, req models.CambioEstadoSolicitudRequest) (models.CambioEstadoSolicitudResponse, error) {
	if solicitudId <= 0 {
		return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("solicitudId es obligatorio")
	}
	if strings.TrimSpace(req.NuevoEstado) == "" {
		return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("Estado Destino es obligatorio")
	}
	if strings.TrimSpace(req.RolUsuario) == "" {
		return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("Rol Usuario es obligatorio")
	}
	if strings.TrimSpace(req.NumeroIdentificacion) == "" {
		return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("Numero Identificacion es obligatorio")
	}

	// CRUD comisiones
	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	logs.Info("UrlComisionesCrud=%q", baseCrud)
	if baseCrud == "" {
		return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("no está configurado UrlComisionesCrud")
	}

	// CRUD terceros
	baseTerceros := strings.TrimSpace(beego.AppConfig.String("UrlTercerosCrud"))
	logs.Info("UrlTercerosCrud=%q", baseTerceros)
	if baseTerceros == "" {
		return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("no está configurado UrlTercerosCrud")
	}

	estadoDestinoId, err := getIdByCodigoAbreviacion(baseCrud, "estado_solicitud", req.NuevoEstado)
	if err != nil {
		return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("no se pudo resolver EstadoDestino=%s: %v", req.NuevoEstado, err)
	}

	terceroId, err := getTerceroIdByNumeroIdentificacion(baseTerceros, req.NumeroIdentificacion)
	if err != nil {
		return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("no se pudo resolver tercero por NumeroIdentificacion=%s: %v", req.NumeroIdentificacion, err)
	}

	histActual, err := getHistoricoActivoActual(baseCrud, solicitudId)
	if err != nil {
		return models.CambioEstadoSolicitudResponse{}, err
	}

	resp := models.CambioEstadoSolicitudResponse{
		SolicitudId:     solicitudId,
		EstadoDestinoId: estadoDestinoId,
		TerceroId:       terceroId,
		Mensaje:         "OK",
	}

	if histActual != nil {
		resp.HistoricoAnteriorId, _ = strconv.Atoi(fmt.Sprintf("%v", histActual["Id"]))

		if estObj, ok := histActual["EstadoSolicitudId"].(map[string]interface{}); ok {
			resp.EstadoAnteriorId, _ = strconv.Atoi(fmt.Sprintf("%v", estObj["Id"]))
		} else {
			resp.EstadoAnteriorId, _ = strconv.Atoi(fmt.Sprintf("%v", histActual["EstadoSolicitudId"]))
		}

		if resp.HistoricoAnteriorId > 0 {
			if err := desactivarHistorico(baseCrud, resp.HistoricoAnteriorId); err != nil {
				return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("error desactivando histórico anterior: %v", err)
			}
		}
	}

	payloadNuevo := map[string]interface{}{
		"SolicitudId":       map[string]interface{}{"Id": solicitudId},
		"EstadoSolicitudId": map[string]interface{}{"Id": estadoDestinoId},
		"RolUsuario":        strings.TrimSpace(req.RolUsuario),
		"TerceroId":         terceroId,
		"Activo":            true,
	}

	postHistoricoURL := helpers.JoinURL(baseCrud, "/historico_estado_solicitud")
	if err := helpers.ValidateAbsoluteURL(postHistoricoURL); err != nil {
		return models.CambioEstadoSolicitudResponse{}, err
	}

	var postHistoricoResp map[string]interface{}
	err = request.SendJson(postHistoricoURL, "POST", &postHistoricoResp, payloadNuevo)
	if err != nil {
		return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("error creando histórico nuevo: %v", err)
	}

	resp.CrudResponse = postHistoricoResp
	resp.HistoricoNuevoId = helpers.ExtractIdAtoi(postHistoricoResp)
	if resp.HistoricoNuevoId <= 0 {
		return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("se creó el histórico pero no se pudo extraer su Id de la respuesta del CRUD")
	}

	// Observación opcional del cambio de estado
	if strings.TrimSpace(req.Observacion) != "" {
		observacionId, err := CrearObservacion(baseCrud, resp.HistoricoNuevoId, req.Observacion)
		if err != nil {
			return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("error creando observación: %v", err)
		}
		resp.ObservacionId = observacionId
	}

	// Documentos múltiples opcionales asociados al nuevo histórico
	if len(req.Documentos) > 0 {
		documentoIds, documentoSolicitudIds, err := crearDocumentosSolicitudMultiples(baseCrud, resp.HistoricoNuevoId, req.Documentos)
		if err != nil {
			return models.CambioEstadoSolicitudResponse{}, err
		}

		resp.DocumentoIds = documentoIds
		resp.DocumentoSolicitudIds = documentoSolicitudIds

		// Compatibilidad hacia atrás, si aún tienes estos campos en el response
		if len(documentoIds) > 0 {
			resp.DocumentoId = documentoIds[0]
		}
		if len(documentoSolicitudIds) > 0 {
			resp.DocumentoSolicitudId = documentoSolicitudIds[0]
		}
	}

	// Crear comisión solo si el código abreviación es APROB_EJEC
	if strings.EqualFold(strings.TrimSpace(req.NuevoEstado), "APROB_EJEC") {
		comisionId, err := CrearComision(baseCrud, solicitudId, terceroId, req.RolUsuario)
		if err != nil {
			logs.Error("error creando comisión para solicitud %d: %v", solicitudId, err)
		} else if comisionId > 0 {
			resp.ComisionId = comisionId
		}
	}

	return resp, nil
}

func crearDocumentosSolicitudMultiples(
	baseCrud string,
	historicoId int,
	documentosReq []models.DocumentoCambioEstadoRequest) ([]int, []int, error) {
	if historicoId <= 0 {
		return nil, nil, fmt.Errorf("historicoId es obligatorio")
	}

	if len(documentosReq) == 0 {
		return []int{}, []int{}, nil
	}

	documentosGestor := make([]models.CrearDocumentoGestorDocumental, 0, len(documentosReq))

	for i, doc := range documentosReq {
		if doc.IdTipoDocumento <= 0 {
			return nil, nil, fmt.Errorf("Documentos[%d].IdTipoDocumento es obligatorio", i)
		}
		if strings.TrimSpace(doc.NombreArchivo) == "" {
			return nil, nil, fmt.Errorf("Documentos[%d].NombreArchivo es obligatorio", i)
		}
		if strings.TrimSpace(doc.File) == "" {
			return nil, nil, fmt.Errorf("Documentos[%d].File es obligatorio", i)
		}

		documentosGestor = append(documentosGestor, models.CrearDocumentoGestorDocumental{
			IdTipoDocumento: doc.IdTipoDocumento,
			Nombre:          strings.TrimSpace(doc.NombreArchivo),
			Descripcion:     strings.TrimSpace(doc.DescripcionDocumento),
			Metadatos:       doc.Metadatos,
			File:            strings.TrimSpace(doc.File),
		})
	}

	resultadoDocs, outputError := helpers.CrearDocumento(documentosGestor)
	if outputError != nil {
		return nil, nil, fmt.Errorf("error creando documentos en gestor documental: %v", outputError)
	}

	if len(resultadoDocs) == 0 {
		return nil, nil, fmt.Errorf("no se recibió respuesta con documentos creados")
	}

	if len(resultadoDocs) != len(documentosReq) {
		return nil, nil, fmt.Errorf("la cantidad de documentos creados no coincide con la cantidad enviada")
	}

	documentoIds := make([]int, 0, len(resultadoDocs))
	documentoSolicitudIds := make([]int, 0, len(resultadoDocs))

	for i, resultado := range resultadoDocs {
		documentoId, err := strconv.Atoi(fmt.Sprintf("%v", resultado["id"]))
		if err != nil || documentoId <= 0 {
			return nil, nil, fmt.Errorf("no se pudo extraer el id del documento creado en la posición %d", i)
		}

		var tipoDocumentoSolicitudId int
		if strings.TrimSpace(documentosReq[i].TipoDocumento) != "" {
			tipoDocumentoSolicitudId, err = getIdByCodigoAbreviacion(
				baseCrud,
				"tipo_documento_solicitud",
				documentosReq[i].TipoDocumento,
			)
			if err != nil {
				return nil, nil, fmt.Errorf("no se pudo resolver TipoDocumento del documento %d: %v", i, err)
			}
		}

		var estadoDocumentoId int
		if strings.TrimSpace(documentosReq[i].EstadoDocumento) != "" {
			estadoDocumentoId, err = getIdByCodigoAbreviacion(
				baseCrud,
				"estado_documento",
				documentosReq[i].EstadoDocumento,
			)
			if err != nil {
				return nil, nil, fmt.Errorf("no se pudo resolver EstadoDocumento del documento %d: %v", i, err)
			}
		}

		documentoSolicitudId, err := crearDocumentoSolicitud(
			baseCrud,
			historicoId,
			documentoId,
			tipoDocumentoSolicitudId,
			estadoDocumentoId,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error creando documento_solicitud para el documento %d: %v", i, err)
		}

		documentoIds = append(documentoIds, documentoId)
		documentoSolicitudIds = append(documentoSolicitudIds, documentoSolicitudId)
	}

	return documentoIds, documentoSolicitudIds, nil
}

func getIdByCodigoAbreviacion(base, recurso, codigo string) (int, error) {
	u, err := url.Parse(helpers.JoinURL(base, "/"+recurso))
	if err != nil {
		return 0, err
	}

	q := u.Query()
	q.Set("query", fmt.Sprintf("CodigoAbreviacion:%s,Activo:true", strings.TrimSpace(codigo)))
	q.Set("limit", "1")
	u.RawQuery = q.Encode()

	var envelope map[string]interface{}
	err = request.GetJson(u.String(), &envelope)
	if err != nil {
		return 0, err
	}

	raw := envelope["Data"]
	arr, ok := raw.([]interface{})
	if !ok || len(arr) == 0 {
		return 0, fmt.Errorf("no existe registro en %s con CodigoAbreviacion=%s", recurso, codigo)
	}

	row, ok := arr[0].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("respuesta inválida: Data[0] no es objeto")
	}

	return strconv.Atoi(fmt.Sprintf("%v", row["Id"]))
}

func getTerceroIdByNumeroIdentificacion(baseTerceros, numero string) (int, error) {
	u, err := url.Parse(helpers.JoinURL(baseTerceros, "/datos_identificacion"))
	if err != nil {
		return 0, err
	}

	q := u.Query()
	q.Set("query", fmt.Sprintf("numero:%s", strings.TrimSpace(numero)))
	q.Set("limit", "1")
	u.RawQuery = q.Encode()

	var rawResp interface{}
	err = request.GetJson(u.String(), &rawResp)
	if err != nil {
		return 0, err
	}

	row, err := helpers.FirstRowFromResponse(rawResp)
	if err != nil {
		return 0, err
	}

	if tObj, ok := row["TerceroId"].(map[string]interface{}); ok {
		return strconv.Atoi(fmt.Sprintf("%v", tObj["Id"]))
	}

	if v, ok := row["TerceroId"]; ok {
		return strconv.Atoi(fmt.Sprintf("%v", v))
	}

	return 0, fmt.Errorf("respuesta inválida: no existe TerceroId en datos_identificacion")
}

func crearDocumentoSolicitud(baseCrud string, historicoId int, id int, tipoDocumentoId int, estadoDocumentoId int) (int, error) {
	postURL := helpers.JoinURL(baseCrud, "/documento_solicitud")
	if err := helpers.ValidateAbsoluteURL(postURL); err != nil {
		return 0, err
	}

	payload := map[string]interface{}{
		"DocumentoId":                id,
		"HistoricoEstadoSolicitudId": map[string]interface{}{"Id": historicoId},
		"Activo":                     true,
	}

	if tipoDocumentoId > 0 {
		payload["TipoDocumentoId"] = map[string]interface{}{"Id": tipoDocumentoId}
	}
	if estadoDocumentoId > 0 {
		payload["EstadoDocumentoId"] = map[string]interface{}{"Id": estadoDocumentoId}
	}

	var postResp map[string]interface{}
	err := request.SendJson(postURL, "POST", &postResp, payload)
	if err != nil {
		return 0, fmt.Errorf("error creando documento_solicitud: %v", err)
	}

	id = helpers.ExtractIdAtoi(postResp)
	return id, nil
}

func getHistoricoActivoActual(base string, solicitudId int) (map[string]interface{}, error) {
	u, err := url.Parse(helpers.JoinURL(base, "/historico_estado_solicitud"))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("query", fmt.Sprintf("SolicitudId:%d,Activo:true", solicitudId))
	q.Set("sortby", "FechaCreacion")
	q.Set("order", "desc")
	q.Set("limit", "1")
	u.RawQuery = q.Encode()

	getURL := u.String()
	logs.Info("URL CRUD FINAL = %s", getURL)

	var envelope map[string]interface{}
	err = request.GetJson(getURL, &envelope)
	if err != nil {
		return nil, fmt.Errorf("error consultando histórico actual: %v", err)
	}

	raw := envelope["Data"]
	if raw == nil {
		return nil, nil
	}

	arr, ok := raw.([]interface{})
	if !ok || len(arr) == 0 {
		return nil, nil
	}

	row, ok := arr[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("respuesta inválida: Data[0] no es objeto (type=%T)", arr[0])
	}

	return row, nil
}

func desactivarHistorico(base string, historicoId int) error {
	getURL := helpers.JoinURL(base, fmt.Sprintf("/historico_estado_solicitud/%d", historicoId))
	if err := helpers.ValidateAbsoluteURL(getURL); err != nil {
		return err
	}

	var getResp map[string]interface{}
	err := request.GetJson(getURL, &getResp)
	if err != nil {
		return fmt.Errorf("error GET histórico: %v", err)
	}

	obj := helpers.UnwrapDataToMap(getResp)
	if obj == nil {
		return fmt.Errorf("respuesta inválida al GET del histórico %d: %v", historicoId, getResp)
	}

	obj["Activo"] = false

	var putResp map[string]interface{}
	err = request.SendJson(getURL, "PUT", &putResp, obj)
	if err != nil {
		return fmt.Errorf("error PUT histórico: %v", err)
	}

	return nil
}
