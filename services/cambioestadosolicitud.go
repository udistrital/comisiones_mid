package services

import (
	"fmt"
	"net/url"
	"os"
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
	baseCrud, err := getBaseURL("UrlComisionesCrud", "COMISIONES_MID_V1_COMISIONES_CRUD")
	logs.Info("UrlComisionesCrud=%q", beego.AppConfig.String("UrlComisionesCrud"))
	if err != nil {
		return models.CambioEstadoSolicitudResponse{}, err
	}

	// CRUD terceros
	baseTerceros, err := getBaseURL("UrlTercerosCrud", "COMISIONES_MID_V1_TERCEROS")
	logs.Info("UrlTercerosCrud=%q", beego.AppConfig.String("UrlTercerosCrud"))
	if err != nil {
		return models.CambioEstadoSolicitudResponse{}, err
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

	postHistoricoURL := joinURL(baseCrud, "/historico_estado_solicitud")
	if err := validateAbsoluteURL(postHistoricoURL); err != nil {
		return models.CambioEstadoSolicitudResponse{}, err
	}

	var postHistoricoResp map[string]interface{}
	err = request.SendJson(postHistoricoURL, "POST", &postHistoricoResp, payloadNuevo)
	if err != nil {
		return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("error creando histórico nuevo: %v", err)
	}

	resp.CrudResponse = postHistoricoResp
	resp.HistoricoNuevoId = extractId(postHistoricoResp)
	if resp.HistoricoNuevoId <= 0 {
		return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("se creó el histórico pero no se pudo extraer su Id de la respuesta del CRUD")
	}

	var tipoDocumentoId int
	if strings.TrimSpace(req.TipoDocumento) != "" {
		tipoDocumentoId, err = getIdByCodigoAbreviacion(baseCrud, "tipo_documento_solicitud", req.TipoDocumento)
		if err != nil {
			return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("no se pudo resolver TipoDocumento=%s: %v", req.TipoDocumento, err)
		}
	}

	var estadoDocumentoId int
	if strings.TrimSpace(req.EstadoDocumento) != "" {
		estadoDocumentoId, err = getIdByCodigoAbreviacion(baseCrud, "estado_documento", req.EstadoDocumento)
		if err != nil {
			return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("no se pudo resolver EstadoDocumento=%s: %v", req.EstadoDocumento, err)
		}
	}

	if strings.TrimSpace(req.NombreArchivo) != "" {
		documento := models.Documento{
			Nombre:      strings.TrimSpace(req.NombreArchivo),
			Descripcion: strings.TrimSpace(req.DescripcionDocumento),
			Metadatos:   strings.TrimSpace(req.Metadatos),
			Activo:      true,
		}

		if tipoDocumentoId > 0 {
			documento.TipoDocumento = &models.TipoDocumento{
				Id: tipoDocumentoId,
			}
		}

		documentoId := helpers.CrearDocumento(documento)
		if documentoId <= 0 {
			return models.CambioEstadoSolicitudResponse{}, fmt.Errorf("no se pudo crear el documento con la función de prueba")
		}

		docSolId, err := crearDocumentoSolicitud(baseCrud, resp.HistoricoNuevoId, documentoId, tipoDocumentoId, estadoDocumentoId)
		if err != nil {
			return models.CambioEstadoSolicitudResponse{}, err
		}

		resp.DocumentoId = documentoId
		resp.DocumentoSolicitudId = docSolId
	}

	return resp, nil
}

func getIdByCodigoAbreviacion(base, recurso, codigo string) (int, error) {
	u, err := url.Parse(joinURL(base, "/"+recurso))
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
	u, err := url.Parse(joinURL(baseTerceros, "/datos_identificacion"))
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

	row, err := firstRowFromResponse(rawResp)
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

/*func getUuidDocumentoDesdeApi(baseDocs string, documentoApiId int) (string, error) {
	getURL := joinURL(baseDocs, fmt.Sprintf("/documento/%d", documentoApiId))

	var rawResp interface{}
	err := request.GetJson(getURL, &rawResp)
	if err != nil {
		return "", err
	}

	var obj map[string]interface{}
	switch t := rawResp.(type) {
	case map[string]interface{}:
		if _, hasData := t["Data"]; hasData {
			row, err := firstRowFromResponse(t)
			if err != nil {
				return "", err
			}
			obj = row
		} else {
			obj = t
		}
	default:
		row, err := firstRowFromResponse(rawResp)
		if err != nil {
			return "", err
		}
		obj = row
	}

	var enlace string
	if v, ok := obj["enlace"].(string); ok {
		enlace = v
	} else if v, ok := obj["Enlace"].(string); ok {
		enlace = v
	}

	enlace = strings.TrimSpace(enlace)
	if enlace == "" {
		return "", fmt.Errorf("no se encontró campo enlace en documento")
	}

	return enlace, nil
}
*/

func crearDocumentoSolicitud(baseCrud string, historicoId int, id int, tipoDocumentoId int, estadoDocumentoId int) (int, error) {
	postURL := joinURL(baseCrud, "/documento_solicitud")
	if err := validateAbsoluteURL(postURL); err != nil {
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

	id = extractId(postResp)
	return id, nil
}

func getHistoricoActivoActual(base string, solicitudId int) (map[string]interface{}, error) {
	u, err := url.Parse(joinURL(base, "/historico_estado_solicitud"))
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
	getURL := joinURL(base, fmt.Sprintf("/historico_estado_solicitud/%d", historicoId))
	if err := validateAbsoluteURL(getURL); err != nil {
		return err
	}

	var getResp map[string]interface{}
	err := request.GetJson(getURL, &getResp)
	if err != nil {
		return fmt.Errorf("error GET histórico: %v", err)
	}

	obj := unwrapDataToMap(getResp)
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

func firstRowFromResponse(raw interface{}) (map[string]interface{}, error) {
	if m, ok := raw.(map[string]interface{}); ok {
		if d, ok := m["Data"]; ok {
			switch dd := d.(type) {
			case []interface{}:
				if len(dd) == 0 {
					return nil, fmt.Errorf("respuesta sin datos")
				}
				row, ok := dd[0].(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("Data[0] no es objeto")
				}
				return row, nil
			case map[string]interface{}:
				return dd, nil
			default:
				return nil, fmt.Errorf("formato Data no soportado: %T", d)
			}
		}
		return m, nil
	}

	if arr, ok := raw.([]interface{}); ok {
		if len(arr) == 0 {
			return nil, fmt.Errorf("respuesta sin datos")
		}
		row, ok := arr[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("[0] no es objeto")
		}
		return row, nil
	}

	return nil, fmt.Errorf("formato de respuesta no soportado: %T", raw)
}

func getBaseURL(appKey, envKey string) (string, error) {
	v := strings.TrimSpace(beego.AppConfig.String(appKey))
	if v == "" || strings.Contains(v, "${") {
		v = strings.TrimSpace(os.Getenv(envKey))
	}
	if v == "" {
		return "", fmt.Errorf("no está configurado %s ni %s", appKey, envKey)
	}
	if !strings.HasPrefix(v, "http://") && !strings.HasPrefix(v, "https://") {
		v = "http://" + v
	}
	return strings.TrimRight(strings.TrimSpace(v), "/"), nil
}

func joinURL(base, path string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
}

func validateAbsoluteURL(u string) error {
	parsed, err := url.Parse(u)
	if err != nil {
		return fmt.Errorf("URL inválida: %v", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("URL inválida (sin scheme/host): %s", u)
	}
	return nil
}

func unwrapDataToMap(resp map[string]interface{}) map[string]interface{} {
	if resp == nil {
		return nil
	}
	if raw, ok := resp["Data"]; ok {
		switch d := raw.(type) {
		case []interface{}:
			if len(d) > 0 {
				if m, ok := d[0].(map[string]interface{}); ok {
					return m
				}
			}
		case map[string]interface{}:
			return d
		}
	}
	if _, ok := resp["Id"]; ok {
		return resp
	}
	return nil
}

func extractId(resp map[string]interface{}) int {
	if resp == nil {
		return 0
	}
	if raw, ok := resp["Data"]; ok {
		switch d := raw.(type) {
		case map[string]interface{}:
			if id, err := strconv.Atoi(fmt.Sprintf("%v", d["Id"])); err == nil {
				return id
			}
		case []interface{}:
			if len(d) > 0 {
				if m, ok := d[0].(map[string]interface{}); ok {
					if id, err := strconv.Atoi(fmt.Sprintf("%v", m["Id"])); err == nil {
						return id
					}
				}
			}
		}
	}
	if id, err := strconv.Atoi(fmt.Sprintf("%v", resp["Id"])); err == nil {
		return id
	}
	return 0
}
