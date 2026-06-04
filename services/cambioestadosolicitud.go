package services

import (
	"encoding/json"
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

const historicoEstadoSolicitudPath = "/historico_estado_solicitud/%d"
const codigoTipoSolicitudProrroga = "SOL_PRORROGA"
const codigoEstadoAprobadoDecanatura = "APROB_EJEC"
const campoFechaFinalizacionAnteriorComision = "fecha_finalizacion_anterior_comision"

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

	estadoDestinoId, err := GetIdByCodigoAbreviacion(baseCrud, "estado_solicitud", req.NuevoEstado)
	if err != nil {
		return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("no se pudo resolver EstadoDestino=%s: %v", req.NuevoEstado, err)
	}

	terceroAprobadorId, err := GetTerceroIdByNumeroIdentificacion(baseTerceros, req.NumeroIdentificacion)
	if err != nil {
		return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("no se pudo resolver tercero por NumeroIdentificacion=%s: %v", req.NumeroIdentificacion, err)
	}

	histActual, err := GetHistoricoActivoActual(baseCrud, solicitudId)
	if err != nil {
		return models.CambioEstadoSolicitudResponse{}, err
	}

	resp := models.CambioEstadoSolicitudResponse{
		SolicitudId:     solicitudId,
		EstadoDestinoId: estadoDestinoId,
		TerceroId:       terceroAprobadorId,
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
			if err := DesActivarHistorico(baseCrud, resp.HistoricoAnteriorId); err != nil {
				return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("error desactivando histórico anterior: %v", err)
			}
		}
	}

	payloadNuevo := map[string]interface{}{
		"SolicitudId":       map[string]interface{}{"Id": solicitudId},
		"EstadoSolicitudId": map[string]interface{}{"Id": estadoDestinoId},
		"RolUsuario":        strings.TrimSpace(req.RolUsuario),
		"TerceroId":         terceroAprobadorId,
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
		documentoIds, documentoSolicitudIds, err := CrearDocumentosCambioEstado(
			baseCrud,
			resp.HistoricoNuevoId,
			req.Documentos,
		)
		if err != nil {
			rollbackErr := RevertirCambioEstado(
				baseCrud,
				resp.HistoricoAnteriorId,
				resp.HistoricoNuevoId,
				resp.ObservacionId,
			)
			if rollbackErr != nil {
				return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("error creando documentos: %v\n Tambien fallo el rollback: %v", err, rollbackErr)
			}
			return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("error creando documentos: %v.", err)
		}

		resp.DocumentoIds = documentoIds
		resp.DocumentoSolicitudIds = documentoSolicitudIds

		if len(documentoIds) > 0 {
			resp.DocumentoId = documentoIds[0]
		}
		if len(documentoSolicitudIds) > 0 {
			resp.DocumentoSolicitudId = documentoSolicitudIds[0]
		}
	}

	if err := ProcesarAprobacionProrrogaDecanatura(baseCrud, solicitudId, req); err != nil {
		return models.CambioEstadoSolicitudResponse{}, err
	}

	// Crear comisión solo si el código abreviación es APROB_EJEC
	if strings.EqualFold(strings.TrimSpace(req.NuevoEstado), "APROB_EJEC") {
		comisionId, err := CrearComision(baseCrud, solicitudId, terceroAprobadorId, req.RolUsuario, req.FechaInicio, req.FechaFinal)
		if err != nil {
			logs.Error("error creando comisión para solicitud %d: %v", solicitudId, err)
		} else if comisionId > 0 {
			resp.ComisionId = comisionId
		}
	}

	return resp, nil
}

func CrearDocumentosCambioEstado(baseCrud string, historicoId int, documentosReq []models.DocumentoCambioEstadoRequest) ([]int, []int, error) {
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
		if strings.TrimSpace(doc.Nombre) == "" {
			return nil, nil, fmt.Errorf("Documentos[%d].Nombre es obligatorio", i)
		}
		if strings.TrimSpace(doc.File) == "" {
			return nil, nil, fmt.Errorf("Documentos[%d].File es obligatorio", i)
		}

		documentosGestor = append(documentosGestor, models.CrearDocumentoGestorDocumental{
			IdTipoDocumento: doc.IdTipoDocumento,
			Nombre:          doc.Nombre,
			Descripcion:     doc.Descripcion,
			Metadatos:       doc.Metadatos,
			File:            doc.File,
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
			tipoDocumentoSolicitudId, err = GetIdByCodigoAbreviacion(
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
			estadoDocumentoId, err = GetIdByCodigoAbreviacion(
				baseCrud,
				"estado_documento",
				documentosReq[i].EstadoDocumento,
			)
			if err != nil {
				return nil, nil, fmt.Errorf("no se pudo resolver EstadoDocumento del documento %d: %v", i, err)
			}
		}

		documentoSolicitudId, err := CrearDocumentoSolicitud(
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

func GetIdByCodigoAbreviacion(base, recurso, codigo string) (int, error) {
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

func GetTerceroIdByNumeroIdentificacion(baseTerceros, numero string) (int, error) {
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

func CrearDocumentoSolicitud(baseCrud string, historicoId int, id int, tipoDocumentoId int, estadoDocumentoId int) (int, error) {
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

func GetHistoricoActivoActual(base string, solicitudId int) (map[string]interface{}, error) {
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

func DesActivarHistorico(base string, historicoId int) error {
	getURL := helpers.JoinURL(base, fmt.Sprintf(historicoEstadoSolicitudPath, historicoId))
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

func RevertirCambioEstado(baseCrud string, historicoAnteriorId int, historicoNuevoId int, observacionId int) error {
	if observacionId > 0 {
		if err := EliminarObservacion(baseCrud, observacionId); err != nil {
			return fmt.Errorf("no se pudo eliminar la observación %d: %v", observacionId, err)
		}
	}

	if historicoNuevoId > 0 {
		if err := EliminarHistorico(baseCrud, historicoNuevoId); err != nil {
			return fmt.Errorf("no se pudo eliminar el histórico nuevo %d: %v", historicoNuevoId, err)
		}
	}

	if historicoAnteriorId > 0 {
		if err := ActivarHistorico(baseCrud, historicoAnteriorId); err != nil {
			return fmt.Errorf("no se pudo reactivar el histórico anterior %d: %v", historicoAnteriorId, err)
		}
	}
	return nil
}

func EliminarObservacion(baseCrud string, observacionId int) error {
	deleteURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/observacion/%d", observacionId))
	if err := helpers.ValidateAbsoluteURL(deleteURL); err != nil {
		return err
	}

	var deleteResp map[string]interface{}
	if err := request.SendJson(deleteURL, "DELETE", &deleteResp, map[string]interface{}{}); err != nil {
		return fmt.Errorf("error eliminando observación %d: %v", observacionId, err)
	}

	return nil
}

func EliminarHistorico(baseCrud string, historicoId int) error {
	deleteURL := helpers.JoinURL(baseCrud, fmt.Sprintf(historicoEstadoSolicitudPath, historicoId))
	if err := helpers.ValidateAbsoluteURL(deleteURL); err != nil {
		return err
	}

	var deleteResp map[string]interface{}
	if err := request.SendJson(deleteURL, "DELETE", &deleteResp, map[string]interface{}{}); err != nil {
		return fmt.Errorf("error eliminando histórico %d: %v", historicoId, err)
	}

	return nil
}

func ActivarHistorico(base string, historicoId int) error {
	getURL := helpers.JoinURL(base, fmt.Sprintf(historicoEstadoSolicitudPath, historicoId))
	if err := helpers.ValidateAbsoluteURL(getURL); err != nil {
		return err
	}

	var getResp map[string]interface{}
	if err := request.GetJson(getURL, &getResp); err != nil {
		return fmt.Errorf("error GET histórico %d: %v", historicoId, err)
	}

	obj := helpers.UnwrapDataToMap(getResp)
	if obj == nil {
		return fmt.Errorf("respuesta inválida al consultar histórico %d", historicoId)
	}

	obj["Activo"] = true

	var putResp map[string]interface{}
	if err := request.SendJson(getURL, "PUT", &putResp, obj); err != nil {
		return fmt.Errorf("error reactivando histórico %d: %v", historicoId, err)
	}

	return nil
}

func ProcesarAprobacionProrrogaDecanatura(baseCrud string, solicitudId int, req models.CambioEstadoSolicitudRequest) error {
	if !strings.EqualFold(strings.TrimSpace(req.NuevoEstado), codigoEstadoAprobadoDecanatura) {
		return nil
	}

	solicitudObj, err := ObtenerSolicitudPorId(baseCrud, solicitudId)
	if err != nil {
		return err
	}

	if !EsSolicitudProrroga(solicitudObj) {
		return nil
	}

	if strings.TrimSpace(req.FechaFinal) == "" || strings.TrimSpace(req.FechaFinalAnterior) == "" {
		return fmt.Errorf("FechaFinal y FechaFinalAnterior son obligatorios para aprobar una solicitud de prórroga")
	}

	comisionId := ExtraerComisionIdDesdeSolicitud(solicitudObj)
	if comisionId <= 0 {
		return fmt.Errorf("la solicitud de prórroga %d no tiene comisión asociada", solicitudId)
	}

	if err := PersistirFechaFinalAnteriorEnDetalleSolicitud(baseCrud, solicitudId, req); err != nil {
		return err
	}

	if err := ActualizarFechaFinalComision(baseCrud, comisionId, req.FechaFinal); err != nil {
		return err
	}

	return nil
}

func ObtenerSolicitudPorId(baseCrud string, solicitudId int) (map[string]interface{}, error) {
	getURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/solicitud/%d", solicitudId))
	if err := helpers.ValidateAbsoluteURL(getURL); err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := request.GetJson(getURL, &resp); err != nil {
		return nil, fmt.Errorf("error consultando solicitud %d: %v", solicitudId, err)
	}

	obj := helpers.UnwrapDataToMap(resp)
	if obj == nil {
		return nil, fmt.Errorf("respuesta inválida al consultar solicitud %d", solicitudId)
	}

	return obj, nil
}

func EsSolicitudProrroga(solicitudObj map[string]interface{}) bool {
	if solicitudObj == nil {
		return false
	}

	tipoObj, ok := solicitudObj["TipoSolicitudId"].(map[string]interface{})
	if !ok || tipoObj == nil {
		return false
	}

	codigo := strings.TrimSpace(fmt.Sprintf("%v", tipoObj["CodigoAbreviacion"]))
	if strings.EqualFold(codigo, codigoTipoSolicitudProrroga) {
		return true
	}

	codigo = strings.TrimSpace(fmt.Sprintf("%v", tipoObj["codigo_abreviacion"]))
	return strings.EqualFold(codigo, codigoTipoSolicitudProrroga)
}

func PersistirFechaFinalAnteriorEnDetalleSolicitud(baseCrud string, solicitudId int, req models.CambioEstadoSolicitudRequest) error {
	detalleId, detalleObj, err := obtenerDetalleSolicitudActivo(baseCrud, solicitudId)
	if err != nil {
		return err
	}
	if detalleObj == nil || detalleId <= 0 {
		return fmt.Errorf("no se encontró detalle_solicitud activo para la solicitud %d", solicitudId)
	}

	formulario, err := NormalizarFormularioProrroga(req.Formulario, detalleObj["Formulario"])
	if err != nil {
		return err
	}

	solicitudFormulario, ok := formulario["solicitud"].(map[string]interface{})
	if !ok || solicitudFormulario == nil {
		solicitudFormulario = map[string]interface{}{}
	}

	solicitudFormulario[campoFechaFinalizacionAnteriorComision] = strings.TrimSpace(req.FechaFinalAnterior)
	formulario["solicitud"] = solicitudFormulario

	formularioBytes, err := json.Marshal(formulario)
	if err != nil {
		return fmt.Errorf("error serializando formulario de prórroga: %v", err)
	}

	detalleObj["Formulario"] = string(formularioBytes)

	putURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/detalle_solicitud/%d", detalleId))
	if err := helpers.ValidateAbsoluteURL(putURL); err != nil {
		return err
	}

	var putResp map[string]interface{}
	if err := request.SendJson(putURL, "PUT", &putResp, detalleObj); err != nil {
		return fmt.Errorf("error actualizando detalle_solicitud %d con fecha final anterior: %v", detalleId, err)
	}

	return nil
}

func NormalizarFormularioProrroga(formularioPayload interface{}, formularioActual interface{}) (map[string]interface{}, error) {
	origen := formularioPayload
	if origen == nil {
		origen = formularioActual
	}

	if origen == nil {
		return map[string]interface{}{}, nil
	}

	switch formulario := origen.(type) {
	case map[string]interface{}:
		return formulario, nil

	case string:
		formulario = strings.TrimSpace(formulario)
		if formulario == "" {
			return map[string]interface{}{}, nil
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(formulario), &parsed); err != nil {
			return nil, fmt.Errorf("error parseando formulario de detalle_solicitud: %v", err)
		}
		return parsed, nil

	default:
		return nil, fmt.Errorf("formato de Formulario no soportado: %T", origen)
	}
}

func ActualizarFechaFinalComision(baseCrud string, comisionId int, fechaFinal string) error {
	getURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/comision/%d", comisionId))
	if err := helpers.ValidateAbsoluteURL(getURL); err != nil {
		return err
	}

	var getResp map[string]interface{}
	if err := request.GetJson(getURL, &getResp); err != nil {
		return fmt.Errorf("error consultando comisión %d: %v", comisionId, err)
	}

	comisionObj := helpers.UnwrapDataToMap(getResp)
	if comisionObj == nil {
		return fmt.Errorf("respuesta inválida al consultar comisión %d", comisionId)
	}

	comisionObj["Id"] = comisionId
	comisionObj["FechaFinal"] = strings.TrimSpace(fechaFinal)

	var putResp map[string]interface{}
	if err := request.SendJson(getURL, "PUT", &putResp, comisionObj); err != nil {
		return fmt.Errorf("error actualizando fecha_final de la comisión %d: %v", comisionId, err)
	}

	return nil
}
