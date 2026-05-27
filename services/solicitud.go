package services

import (
	"fmt"
	"net/url"

	"encoding/json"

	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/utils_oas/request"
)

func resolverTipoSolicitudId(tipoSolicitudId int, codigo string) (int, error) {
	if tipoSolicitudId > 0 {
		return tipoSolicitudId, nil
	}

	if codigo == "" {
		return 0, fmt.Errorf("debe enviar tipo_solicitud_id o cod_abreviacion_tipo_solicitud")
	}

	var resp map[string]interface{}
	err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+"tipo_solicitud?query=CodigoAbreviacion:"+url.QueryEscape(codigo),
		&resp,
	)
	if err != nil {
		return 0, fmt.Errorf("error consultando tipo_solicitud por código %s: %v", codigo, err)
	}

	data, ok := resp["Data"].([]interface{})
	if !ok || len(data) == 0 {
		return 0, fmt.Errorf("no se encontró tipo_solicitud para código %s", codigo)
	}

	id, ok := data[0].(map[string]interface{})["Id"].(float64)
	if !ok {
		return 0, fmt.Errorf("respuesta inválida consultando tipo_solicitud para código %s", codigo)
	}

	return int(id), nil
}

func CrearSolicitud(solicitud models.CrearSolicitudEntrada) (respuesta models.Solicitud, outputError map[string]interface{}) {

	defer func() {
		if err := recover(); err != nil {
			outputError = map[string]interface{}{
				"funcion": "CrearSolicitud",
				"error":   err,
				"status":  500,
			}
		}
	}()

	var persona []map[string]interface{}
	err := request.GetJson(beego.AppConfig.String("UrlTercerosCrud")+
		"datos_identificacion?query=Numero:"+fmt.Sprintf("%d", solicitud.Identificacion), &persona)

	if err != nil {
		return respuesta, map[string]interface{}{"error": "Error consultando tercero", "detalle": err.Error()}
	}

	if len(persona) == 0 {
		return respuesta, map[string]interface{}{"error": "No se encontró el tercero"}
	}

	terceroMap, ok := persona[0]["TerceroId"].(map[string]interface{})
	if !ok {
		return respuesta, map[string]interface{}{"error": "Estructura inválida de TerceroId"}
	}

	idTercero := int(terceroMap["Id"].(float64))

	tipoSolicitudId, err := resolverTipoSolicitudId(solicitud.TipoSolicitudId, solicitud.CodigoAbreviacionTipo)
	if err != nil {
		return respuesta, map[string]interface{}{"error": err.Error()}
	}

	req := models.SolicitudCreateRequest{
		TerceroId: idTercero,
		Activo:    true,
		TipoSolicitudId: models.IdReference{
			Id: tipoSolicitudId,
		},
		ObservacionCierre: solicitud.Observacion,
	}
	var respSolicitud map[string]interface{}

	err = request.SendJson(beego.AppConfig.String("UrlComisionesCrud")+"solicitud", "POST", &respSolicitud, &req)
	if err != nil {
		return respuesta, map[string]interface{}{
			"error":   "Error en request creando solicitud",
			"detalle": err.Error(),
		}
	}

	var errorCreacionSolicitud map[string]interface{}
	dataSolicitud, errorCreacionSolicitud := helpers.ValidarRespuesta(respSolicitud)
	if errorCreacionSolicitud != nil {
		return respuesta, errorCreacionSolicitud
	}

	idRaw, ok := dataSolicitud["Id"]
	if !ok {
		return respuesta, map[string]interface{}{
			"error": "No se encontró Id en la respuesta",
		}
	}

	idSolicitudFloat, ok := idRaw.(float64)
	if !ok {
		return respuesta, map[string]interface{}{
			"error": "Id con tipo inválido",
		}
	}

	idSolicitud := int(idSolicitudFloat)
	solicitudTemp := models.Solicitud{Id: idSolicitud, ObservacionCierre: solicitud.Observacion}

	formularioBytes, _ := json.Marshal(solicitud.Formulario)

	detalle := models.DetalleSolicitud{
		SolicitudId: &solicitudTemp,
		Formulario:  string(formularioBytes),
		Activo:      true,
	}

	var respDetalle map[string]interface{}
	err = request.SendJson(beego.AppConfig.String("UrlComisionesCrud")+"detalle_solicitud", "POST", &respDetalle, &detalle)

	if err != nil {
		return respuesta, map[string]interface{}{"error": "Error creando detalle", "detalle": err.Error()}
	}

	var respEstado map[string]interface{}
	err = request.GetJson(beego.AppConfig.String("UrlComisionesCrud")+"estado_solicitud?query=CodigoAbreviacion:NO_ENV", &respEstado)
	if err != nil {
		return respuesta, map[string]interface{}{"error": "Error consultando estado"}
	}
	dataEstado := respEstado["Data"].([]interface{})
	idEstado := int(dataEstado[0].(map[string]interface{})["Id"].(float64))
	historico := models.HistoricoEstadoSolicitud{
		SolicitudId:       &solicitudTemp,
		EstadoSolicitudId: &models.EstadoSolicitud{Id: idEstado},
		RolUsuario:        solicitud.CodigoAbreviacionRol,
		TerceroId:         idTercero,
		Activo:            true,
	}

	var respHistorico map[string]interface{}
	err = request.SendJson(beego.AppConfig.String("UrlComisionesCrud")+"historico_estado_solicitud", "POST", &respHistorico, &historico)

	if err != nil {
		return respuesta, map[string]interface{}{"error": "Error creando histórico"}
	}

	idHistorico := int(respHistorico["Data"].(map[string]interface{})["Id"].(float64))

	if len(solicitud.DocumentoSolicitud) > 0 {
		docs, errDoc := helpers.CrearDocumento(solicitud.DocumentoSolicitud)
		if errDoc != nil {
			return respuesta, map[string]interface{}{"error": "Error creando documentos"}
		}
		for _, doc := range docs {
			idDoc := int(doc["id"].(int))
			documento := models.DocumentoSolicitud{
				DocumentoId: idDoc,
				HistoricoEstadoSolicitudId: &models.HistoricoEstadoSolicitud{
					Id: idHistorico,
				},
				TipoDocumentoId: &models.TipoDocumentoSolicitud{
					Id: 1,
				},
				EstadoDocumentoId: &models.EstadoDocumento{
					Id: 1,
				},
				Activo: true,
			}

			var respDoc map[string]interface{}
			err = request.SendJson(beego.AppConfig.String("UrlComisionesCrud")+"documento_solicitud", "POST", &respDoc, &documento)
			if err != nil {
				return respuesta, map[string]interface{}{
					"error":   "Error vinculando documento",
					"idDoc":   idDoc,
					"detalle": err.Error(),
				}
			}
		}
	}

	return solicitudTemp, nil
}

func EditarSolicitud(solicitudId int, req models.EditarSolicitud) (models.EditarSolicitudResponse, error) {
	var respuesta models.EditarSolicitudResponse
	respuesta.SolicitudId = solicitudId

	if solicitudId <= 0 {
		return respuesta, fmt.Errorf("solicitudId es obligatorio")
	}

	baseCrud := beego.AppConfig.String("UrlComisionesCrud")

	if err := actualizarSolicitud(baseCrud, solicitudId, req); err != nil {
		return respuesta, err
	}

	detalleId, err := edicionDetalleSolicitud(baseCrud, solicitudId, req.Formulario)
	if err != nil {
		return respuesta, err
	}
	respuesta.DetalleSolicitudId = detalleId

	docsADesactivar := DocumentosADesactivar(req)

	if len(req.DocumentosNuevos) > 0 {
		historicoActivoId, err := obtenerHistoricoActivo(baseCrud, solicitudId)
		if err != nil {
			return respuesta, err
		}

		respuesta.HistoricoEstadoSolicitudId = historicoActivoId

		documentosIds, documentoSolicitudIds, err := CrearDocumentosSolicitud(baseCrud, historicoActivoId, req.DocumentosNuevos)
		if err != nil {
			return respuesta, err
		}

		respuesta.DocumentoIds = documentosIds
		respuesta.DocumentoSolicitudIds = documentoSolicitudIds
	}

	if len(docsADesactivar) > 0 {
		if err := desactivarDocumentosSolicitudAsociados(baseCrud, solicitudId, docsADesactivar); err != nil {
			return respuesta, err
		}
		respuesta.DocumentosDesactivados = docsADesactivar
	}

	respuesta.Mensaje = "Solicitud actualizada correctamente"
	return respuesta, nil
}

func actualizarSolicitud(baseCrud string, solicitudId int, req models.EditarSolicitud) error {
	getURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/solicitud/%d", solicitudId))
	if err := helpers.ValidateAbsoluteURL(getURL); err != nil {
		return err
	}

	var getResp map[string]interface{}
	if err := request.GetJson(getURL, &getResp); err != nil {
		return fmt.Errorf("error consultando solicitud %d: %v", solicitudId, err)
	}

	obj := helpers.UnwrapDataToMap(getResp)
	if obj == nil {
		return fmt.Errorf("respuesta inválida al consultar solicitud %d", solicitudId)
	}

	tipoSolicitudId, err := resolverTipoSolicitudId(req.TipoSolicitudId, req.CodigoAbreviacionTipo)
	if err != nil && (req.TipoSolicitudId > 0 || req.CodigoAbreviacionTipo != "") {
		return err
	}

	if tipoSolicitudId > 0 {
		obj["TipoSolicitudId"] = map[string]interface{}{"Id": tipoSolicitudId}
	}

	obj["ObservacionCierre"] = req.Observacion

	var putResp map[string]interface{}
	if err := request.SendJson(getURL, "PUT", &putResp, obj); err != nil {
		return fmt.Errorf("error actualizando solicitud %d: %v", solicitudId, err)
	}

	return nil
}

func edicionDetalleSolicitud(baseCrud string, solicitudId int, formulario map[string]interface{}) (int, error) {
	if formulario == nil {
		return 0, nil
	}

	formularioBytes, err := json.Marshal(formulario)
	if err != nil {
		return 0, fmt.Errorf("error convirtiendo formulario a JSON: %v", err)
	}

	detalleId, detalleObj, err := obtenerDetalleSolicitudActivo(baseCrud, solicitudId)
	if err != nil {
		return 0, err
	}

	if detalleObj == nil {
		postURL := helpers.JoinURL(baseCrud, "/detalle_solicitud")
		if err := helpers.ValidateAbsoluteURL(postURL); err != nil {
			return 0, err
		}

		payload := map[string]interface{}{
			"SolicitudId": map[string]interface{}{"Id": solicitudId},
			"Formulario":  string(formularioBytes),
			"Activo":      true,
		}

		var postResp map[string]interface{}
		if err := request.SendJson(postURL, "POST", &postResp, payload); err != nil {
			return 0, fmt.Errorf("error creando detalle_solicitud para la solicitud %d: %v", solicitudId, err)
		}

		return helpers.ExtractIdAtoi(postResp), nil
	}

	detalleObj["Formulario"] = string(formularioBytes)

	putURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/detalle_solicitud/%d", detalleId))
	if err := helpers.ValidateAbsoluteURL(putURL); err != nil {
		return 0, err
	}

	var putResp map[string]interface{}
	if err := request.SendJson(putURL, "PUT", &putResp, detalleObj); err != nil {
		return 0, fmt.Errorf("error actualizando detalle_solicitud %d: %v", detalleId, err)
	}

	return detalleId, nil
}

func obtenerDetalleSolicitudActivo(baseCrud string, solicitudId int) (int, map[string]interface{}, error) {
	u, err := url.Parse(helpers.JoinURL(baseCrud, "/detalle_solicitud"))
	if err != nil {
		return 0, nil, err
	}

	q := u.Query()
	q.Set("query", fmt.Sprintf("SolicitudId:%d,Activo:true", solicitudId))
	q.Set("sortby", "FechaCreacion")
	q.Set("order", "desc")
	q.Set("limit", "1")
	u.RawQuery = q.Encode()

	var resp map[string]interface{}
	if err := request.GetJson(u.String(), &resp); err != nil {
		return 0, nil, fmt.Errorf("error consultando detalle_solicitud activo: %v", err)
	}

	obj := helpers.UnwrapDataToMap(resp)
	if obj == nil {
		return 0, nil, nil
	}

	detalleId := helpers.ExtractIdAtoi(resp)
	if detalleId <= 0 {
		return 0, nil, fmt.Errorf("no se pudo determinar el id del detalle_solicitud activo de la solicitud %d", solicitudId)
	}

	return detalleId, obj, nil
}

func obtenerHistoricoActivo(baseCrud string, solicitudId int) (int, error) {
	historicoObj, err := GetHistoricoActivoActual(baseCrud, solicitudId)
	if err != nil {
		return 0, err
	}

	if historicoObj == nil {
		return 0, fmt.Errorf("no se encontró histórico activo para la solicitud %d", solicitudId)
	}

	historicoId := ExtraerIdRelacionado(historicoObj["Id"])
	if historicoId <= 0 {
		return 0, fmt.Errorf("no se pudo obtener el id del histórico activo para la solicitud %d", solicitudId)
	}

	return historicoId, nil
}

func DocumentosADesactivar(req models.EditarSolicitud) []int {
	ids := make([]int, 0)
	seen := make(map[int]struct{})

	for _, id := range req.DocumentosDesactivar {
		if id > 0 {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				ids = append(ids, id)
			}
		}
	}

	return ids
}

func desactivarDocumentosSolicitudAsociados(baseCrud string, solicitudId int, documentoSolicitudIds []int) error {
	for _, documentoSolicitudId := range documentoSolicitudIds {
		getURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/documento_solicitud/%d", documentoSolicitudId))
		if err := helpers.ValidateAbsoluteURL(getURL); err != nil {
			return err
		}

		var getResp map[string]interface{}
		if err := request.GetJson(getURL, &getResp); err != nil {
			return fmt.Errorf("error consultando documento_solicitud %d: %v", documentoSolicitudId, err)
		}

		obj := helpers.UnwrapDataToMap(getResp)
		if obj == nil {
			return fmt.Errorf("respuesta inválida al consultar documento_solicitud %d", documentoSolicitudId)
		}

		historicoId := ExtraerIdRelacionado(obj["HistoricoEstadoSolicitudId"])
		if historicoId <= 0 {
			return fmt.Errorf("no se pudo obtener el histórico asociado al documento_solicitud %d", documentoSolicitudId)
		}

		if err := validarHistorico(baseCrud, historicoId, solicitudId); err != nil {
			return err
		}

		obj["Activo"] = false

		var putResp map[string]interface{}
		if err := request.SendJson(getURL, "PUT", &putResp, obj); err != nil {
			return fmt.Errorf("error desactivando documento_solicitud %d: %v", documentoSolicitudId, err)
		}
	}
	return nil
}

func validarHistorico(baseCrud string, historicoId int, solicitudId int) error {
	getURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/historico_estado_solicitud/%d", historicoId))
	if err := helpers.ValidateAbsoluteURL(getURL); err != nil {
		return err
	}

	var getResp map[string]interface{}
	if err := request.GetJson(getURL, &getResp); err != nil {
		return fmt.Errorf("error consultando histórico %d: %v", historicoId, err)
	}

	obj := helpers.UnwrapDataToMap(getResp)
	if obj == nil {
		return fmt.Errorf("respuesta inválida al consultar histórico %d", historicoId)
	}

	solicitudAsociadaId := ExtraerIdRelacionado(obj["SolicitudId"])
	if solicitudAsociadaId != solicitudId {
		return fmt.Errorf("el documento_solicitud asociado al histórico %d no pertenece a la solicitud %d", historicoId, solicitudId)
	}

	return nil
}

func ExtraerIdRelacionado(obj interface{}) int {
	switch valor := obj.(type) {
	case map[string]interface{}:
		if id, ok := valor["Id"]; ok {
			switch v := id.(type) {
			case float64:
				return int(v)
			case int:
				return v
			}
		}
	case float64:
		return int(valor)
	case int:
		return valor
	}

	return 0
}

func BuscarSolicitudIdentificacion(identificacion int) (respuesta []models.SolicitudResumen, outputError map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			outputError = map[string]interface{}{
				"funcion": "/BuscarSolicitudIdentificacion",
				"err":     err,
				"status":  "404",
			}
			panic(outputError)
		}
	}()

	terceroId, err := obtenerTerceroIdPorIdentificacion(identificacion)
	if err != nil {
		return nil, map[string]interface{}{
			"error":  "no se encontró solicitud",
			"status": 404,
		}
	}

	solicitudes, err := obtenerSolicitudesPorTercero(terceroId)
	if err != nil || len(solicitudes) == 0 {
		return nil, map[string]interface{}{
			"error":  "no se encontró solicitud",
			"status": 404,
		}
	}

	respuesta = make([]models.SolicitudResumen, 0, len(solicitudes))
	for _, item := range solicitudes {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		sol := construirSolicitudResumen(itemMap)
		respuesta = append(respuesta, sol)
	}

	if len(respuesta) == 0 {
		return nil, map[string]interface{}{
			"error":  "no se encontró solicitud",
			"status": 404,
		}
	}

	return respuesta, nil
}

func BuscarDetallesSolicitud(idSolicitud int) (respuesta models.SolicitudDetalles, outputError map[string]interface{}) {

	defer func() {
		if err := recover(); err != nil {
			outputError = map[string]interface{}{
				"funcion": "/BuscarSolicitudIdentificacion",
				"err":     err,
				"status":  "404",
			}
			panic(outputError)
		}
	}()

	var respuestaHistorico map[string]interface{}

	err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"historico_estado_solicitud?query=SolicitudId__Id:"+fmt.Sprintf("%d", idSolicitud)+
			",Activo:true&sortby=FechaCreacion&order=desc&limit=-1",
		&respuestaHistorico,
	)

	if err != nil {
		return respuesta, nil
	}
	data, ok := respuestaHistorico["Data"].([]interface{})
	if !ok || len(data) == 0 {
		return models.SolicitudDetalles{}, map[string]interface{}{
			"error":  "no se encontró solicitud",
			"status": 404,
		}
	}

	primerRegistro, ok := data[0].(map[string]interface{})
	if !ok {
		return respuesta, nil
	}

	infoSolicitud, ok := primerRegistro["SolicitudId"].(map[string]interface{})
	if !ok {
		return respuesta, nil
	}

	if registroTipoSolicitud, ok := infoSolicitud["TipoSolicitudId"].(map[string]interface{}); ok {
		tipoSolicitudHistorico := models.TipoSolicitud{
			Id:                int(registroTipoSolicitud["Id"].(float64)),
			Nombre:            fmt.Sprintf("%v", registroTipoSolicitud["Nombre"]),
			CodigoAbreviacion: fmt.Sprintf("%v", registroTipoSolicitud["CodigoAbreviacion"]),
		}

		if estadoSolicitudActual, ok := primerRegistro["EstadoSolicitudId"].(map[string]interface{}); ok {
			estadoSolicitudInfo := models.EstadoSolicitud{
				Id:                int(estadoSolicitudActual["Id"].(float64)),
				Nombre:            fmt.Sprintf("%v", estadoSolicitudActual["Nombre"]),
				Descripcion:       fmt.Sprintf("%v", estadoSolicitudActual["Descripcion"]),
				CodigoAbreviacion: fmt.Sprintf("%v", estadoSolicitudActual["CodigoAbreviacion"]),
			}
			respuesta.EstadoSolicitud = &estadoSolicitudInfo
		}

		solicitudHistorico := models.Solicitud{
			Id:                int(infoSolicitud["Id"].(float64)),
			TerceroId:         int(infoSolicitud["TerceroId"].(float64)),
			TipoSolicitudId:   &tipoSolicitudHistorico,
			ObservacionCierre: fmt.Sprintf("%v", infoSolicitud["ObservacionCierre"]),
			Activo:            infoSolicitud["Activo"].(bool),
		}
		respuesta.Solicitud = &solicitudHistorico
	}

	var respuestaDetalleFormulario map[string]interface{}
	if err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+"detalle_solicitud?query=SolicitudId__Id:"+fmt.Sprintf("%d", idSolicitud),
		&respuestaDetalleFormulario,
	); err == nil {

		if data_formulario, ok := respuestaDetalleFormulario["Data"].([]interface{}); ok && len(data_formulario) > 0 {
			if registro_formulario, ok := data_formulario[0].(map[string]interface{}); ok {
				respuesta.Formulario = registro_formulario["Formulario"]
			}
		}
	}

	var respuestaDocumentos map[string]interface{}

	if err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+"documento_solicitud?query=HistoricoEstadoSolicitudId__SolicitudId__Id:"+fmt.Sprintf("%d", idSolicitud)+",Activo:true&limit=-1",
		&respuestaDocumentos,
	); err == nil {

		if data_documentos, ok := respuestaDocumentos["Data"].([]interface{}); ok && len(data_documentos) > 0 {
			for _, doc := range data_documentos {
				if documento, ok := doc.(map[string]interface{}); ok {
					idDocumentoComision := int(documento["Id"].(float64))
					var rolDocumentoComision string
					if hist, ok := documento["HistoricoEstadoSolicitudId"].(map[string]interface{}); ok {
						if rol, ok := hist["RolUsuario"].(string); ok {
							rolDocumentoComision = rol
						}
					}
					docId := int(documento["DocumentoId"].(float64))
					var detalleDoc map[string]interface{}
					if err := request.GetJson(
						beego.AppConfig.String("UrlDocumentos")+"documento/"+fmt.Sprintf("%d", docId),
						&detalleDoc,
					); err == nil {

						if len(detalleDoc) == 0 {
							continue
						}
						idDocumento := int(detalleDoc["Id"].(float64))
						nombre, _ := detalleDoc["Nombre"].(string)
						enlace, _ := detalleDoc["Enlace"].(string)

						// TipoDocumento
						var tipo *models.TipoDocumentoSolicitud
						if tipoDoc, ok := documento["TipoDocumentoId"].(map[string]interface{}); ok {
							tipo = &models.TipoDocumentoSolicitud{
								Id:                int(tipoDoc["Id"].(float64)),
								Nombre:            fmt.Sprintf("%v", tipoDoc["Nombre"]),
								Descripcion:       fmt.Sprintf("%v", tipoDoc["Descripcion"]),
								CodigoAbreviacion: fmt.Sprintf("%v", tipoDoc["CodigoAbreviacion"]),
							}
						}

						// EstadoDocumento
						var estado *models.EstadoDocumento
						if estadoDoc, ok := documento["EstadoDocumentoId"].(map[string]interface{}); ok {
							estado = &models.EstadoDocumento{
								Id:                int(estadoDoc["Id"].(float64)),
								Nombre:            fmt.Sprintf("%v", estadoDoc["Nombre"]),
								Descripcion:       fmt.Sprintf("%v", estadoDoc["Descripcion"]),
								CodigoAbreviacion: fmt.Sprintf("%v", estadoDoc["CodigoAbreviacion"]),
							}
						}

						if nombre != "" && enlace != "" {
							documentoAux := models.DocumentoDetalle{
								Id:          idDocumentoComision,
								Rol:         rolDocumentoComision,
								IdDocumento: idDocumento,
								Nombre:      nombre,
								Enlace:      enlace,
								Tipo:        tipo,
								Estado:      estado,
							}

							respuesta.Documentos = append(respuesta.Documentos, documentoAux)
						}
					}
				}
			}
		} else {
			respuesta.Documentos = []models.DocumentoDetalle{}
		}
	}

	var respuestaObservaciones map[string]interface{}

	if err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"observacion?query=HistoricoEstadoSolicitudId__SolicitudId__Id:"+
			fmt.Sprintf("%d", idSolicitud)+",Activo:true&limit=-1",
		&respuestaObservaciones,
	); err == nil {

		if data, ok := respuestaObservaciones["Data"].([]interface{}); ok && len(data) > 0 {

			for _, item := range data {

				observacionMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				var idObservacion int
				if v, ok := observacionMap["Id"].(float64); ok {
					idObservacion = int(v)
				}

				var rolObservacion string
				if hist, ok := observacionMap["HistoricoEstadoSolicitudId"].(map[string]interface{}); ok {
					if rol, ok := hist["RolUsuario"].(string); ok {
						rolObservacion = rol
					}
				}

				descripcion, _ := observacionMap["Descripcion"].(string)

				observacionAux := models.ObservacionDetalle{
					Id:          idObservacion,
					Rol:         rolObservacion,
					Descripcion: descripcion,
				}

				respuesta.Observaciones = append(respuesta.Observaciones, observacionAux)
			}
		} else {
			respuesta.Observaciones = []models.ObservacionDetalle{}
		}
	}

	return respuesta, nil
}

func obtenerTerceroIdPorIdentificacion(identificacion int) (int, error) {
	var tercero []map[string]interface{}
	err := request.GetJson(
		beego.AppConfig.String("UrlTercerosCrud")+"datos_identificacion?query=Numero:"+fmt.Sprintf("%d", identificacion),
		&tercero,
	)
	if err != nil || len(tercero) == 0 || len(tercero[0]) == 0 {
		return 0, fmt.Errorf("tercero no encontrado")
	}

	terceroComprobacion, ok := tercero[0]["TerceroId"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("TerceroId inválido")
	}

	idTercero, ok := terceroComprobacion["Id"].(float64)
	if !ok {
		return 0, fmt.Errorf("Id de tercero inválido")
	}

	return int(idTercero), nil
}

func obtenerSolicitudesPorTercero(terceroId int) ([]interface{}, error) {
	var persona map[string]interface{}
	err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+"solicitud?limit=-1&sortby=id&order=desc&query=TerceroId:"+fmt.Sprintf("%d", terceroId),
		&persona,
	)
	if err != nil {
		return nil, err
	}

	data, ok := persona["Data"].([]interface{})
	if !ok || len(data) == 0 {
		return nil, fmt.Errorf("sin solicitudes")
	}

	return data, nil
}

func construirSolicitudResumen(itemMap map[string]interface{}) models.SolicitudResumen {
	var sol models.SolicitudResumen

	idStr := fmt.Sprintf("%v", itemMap["Id"])
	sol.FechaCreacion = fmt.Sprintf("%v", itemMap["FechaCreacion"])

	if id, ok := itemMap["Id"].(float64); ok {
		sol.Id = int(id)
	}
	if activo, ok := itemMap["Activo"].(bool); ok {
		sol.Activo = activo
	}

	cargarDetalleSolicitudResumen(idStr, &sol)
	cargarEstadoActualSolicitud(idStr, &sol)

	return sol
}

func cargarDetalleSolicitudResumen(idStr string, sol *models.SolicitudResumen) {
	var detalleSolicitud map[string]interface{}
	err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+"detalle_solicitud?query=solicitud_id:"+idStr,
		&detalleSolicitud,
	)
	if err != nil {
		return
	}

	datosFormulario, formError := helpers.ObtenerDatosFormulario(detalleSolicitud)
	if formError != nil {
		return
	}

	sol.Programa = datosFormulario.Solicitante.Q7Proyecto
	sol.Nombre = datosFormulario.Solicitante.Q3NombresApellidos
}

func cargarEstadoActualSolicitud(idStr string, sol *models.SolicitudResumen) {
	var respuestaHistoricoEstadoSolicitudActual map[string]interface{}
	err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+"historico_estado_solicitud?query=solicitudId__Id:"+idStr+",Activo:true&sortby=FechaCreacion&order=desc&limit=1",
		&respuestaHistoricoEstadoSolicitudActual,
	)
	if err != nil {
		return
	}

	data, ok := respuestaHistoricoEstadoSolicitudActual["Data"].([]interface{})
	if !ok || len(data) == 0 {
		return
	}

	primerRegistro, ok := data[0].(map[string]interface{})
	if !ok {
		return
	}

	estado, ok := primerRegistro["EstadoSolicitudId"].(map[string]interface{})
	if !ok {
		return
	}

	sol.EstadoSolicitud = &models.EstadoSolicitud{
		Id:     extraerEstadoID(estado["Id"]),
		Nombre: extraerStringMapa(estado, "Nombre"),
	}
}

func extraerEstadoID(valor interface{}) int {
	switch v := valor.(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

func extraerStringMapa(data map[string]interface{}, key string) string {
	valor, ok := data[key].(string)
	if !ok {
		return ""
	}
	return valor
}
