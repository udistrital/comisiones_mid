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

	id_tercero := int(terceroMap["Id"].(float64))

	req := models.SolicitudCreateRequest{
		TerceroId: id_tercero,
		Activo:    true,
		TipoSolicitudId: models.IdReference{
			Id: solicitud.TipoSolicitudId,
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
	id_estado := int(dataEstado[0].(map[string]interface{})["Id"].(float64))
	historico := models.HistoricoEstadoSolicitud{
		SolicitudId:       &solicitudTemp,
		EstadoSolicitudId: &models.EstadoSolicitud{Id: id_estado},
		RolUsuario:        solicitud.CodigoAbreviacionRol,
		TerceroId:         id_tercero,
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
			fmt.Println("SE CREA BIEN EN COMISIONES")
			fmt.Println(respDoc)
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

	documentosADesactivar := documentosADesactivar(req)

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

	if len(documentosADesactivar) > 0 {
		if err := desactivarDocumentosSolicitudAsociados(baseCrud, solicitudId, documentosADesactivar); err != nil {
			return respuesta, err
		}
		respuesta.DocumentosDesactivados = documentosADesactivar
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

	if req.TipoSolicitudId > 0 {
		obj["TipoSolicitudId"] = map[string]interface{}{"Id": req.TipoSolicitudId}
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
	historicoObj, err := getHistoricoActivoActual(baseCrud, solicitudId)
	if err != nil {
		return 0, err
	}

	if historicoObj == nil {
		return 0, fmt.Errorf("no se encontró histórico activo para la solicitud %d", solicitudId)
	}

	historicoId := extraerIdRelacionado(historicoObj["Id"])
	if historicoId <= 0 {
		return 0, fmt.Errorf("no se pudo obtener el id del histórico activo para la solicitud %d", solicitudId)
	}

	return historicoId, nil
}

func documentosADesactivar(req models.EditarSolicitud) []int {
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

		historicoId := extraerIdRelacionado(obj["HistoricoEstadoSolicitudId"])
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

	solicitudAsociadaId := extraerIdRelacionado(obj["SolicitudId"])
	if solicitudAsociadaId != solicitudId {
		return fmt.Errorf("el documento_solicitud asociado al histórico %d no pertenece a la solicitud %d", historicoId, solicitudId)
	}

	return nil
}

func extraerIdRelacionado(obj interface{}) int {
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

	//Busca el tercero
	var persona map[string]interface{}
	var tercero []map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("UrlTercerosCrud")+"datos_identificacion?query=Numero:"+fmt.Sprintf("%d", identificacion), &tercero); err == nil {
		if len(tercero) > 0 && len(tercero[0]) > 0 {
			if tercero_comprobacion, ok := tercero[0]["TerceroId"].(map[string]interface{}); ok {
				if id_tercero, ok := tercero_comprobacion["Id"].(float64); ok {
					id_tercero_busqueda := int(id_tercero)

					fmt.Println("ENTRA A SERVICIO ", beego.AppConfig.String("UrlComisionesCrud")+"solicitud?query=TerceroId:"+fmt.Sprintf("%d", id_tercero_busqueda)+"&limit=-1")
					if err := request.GetJson(beego.AppConfig.String("UrlComisionesCrud")+"solicitud?limit=-1&query=TerceroId:"+fmt.Sprintf("%d", id_tercero_busqueda),
						&persona); err == nil {
						fmt.Println("ENTRA A SERVICIO 2 ", persona)
						if data, ok := persona["Data"].([]interface{}); ok && len(data) > 0 {
							for _, item := range data {
								var detalleSolicitud map[string]interface{}

								if itemMap, ok := item.(map[string]interface{}); ok {
									var sol models.SolicitudResumen
									idStr := fmt.Sprintf("%v", itemMap["Id"])
									fmt.Println("ID SOLICITUD ", idStr)
									/*if err := request.GetJson(
										beego.AppConfig.String("UrlComisionesCrud")+"historial_solicitud?query=SolicitudId:"+idStr,
										&detalleSolicitud,
									)*/
									sol.FechaCreacion = fmt.Sprintf("%v", itemMap["FechaCreacion"])
									if err := request.GetJson(
										beego.AppConfig.String("UrlComisionesCrud")+"detalle_solicitud?query=solicitud_id:"+idStr,
										&detalleSolicitud,
									); err == nil {
										fmt.Println("DETALLE SOLICITUD ", detalleSolicitud)
										datosFormulario, err := helpers.ObtenerDatosFormulario(detalleSolicitud)
										if err == nil {
											fmt.Println("Programa: ", datosFormulario.Solicitante.Q7Proyecto)
											fmt.Println("Nombre: ", datosFormulario.Solicitante.Q3NombresApellidos)
											if id, ok := itemMap["Id"].(float64); ok {
												sol.Id = int(id)
											}
											if activo, ok := itemMap["Activo"].(bool); ok {
												sol.Activo = activo
											}
											sol.Programa = datosFormulario.Solicitante.Q7Proyecto
											sol.Nombre = datosFormulario.Solicitante.Q3NombresApellidos
										}
										var respuesta_historico_estado_solicitud_actual map[string]interface{}
										var id_estado_solicitud int
										if err := request.GetJson(
											beego.AppConfig.String("UrlComisionesCrud")+"historico_estado_solicitud?query=solicitudId__Id:"+idStr+",Activo:true&sortby=FechaCreacion&order=desc&limit=1",
											&respuesta_historico_estado_solicitud_actual,
										); err == nil {
											if data, ok := respuesta_historico_estado_solicitud_actual["Data"].([]interface{}); ok && len(data) > 0 {

												if primerRegistro, ok := data[0].(map[string]interface{}); ok {
													if estado, ok := primerRegistro["EstadoSolicitudId"].(map[string]interface{}); ok {
														switch v := estado["Id"].(type) {
														case float64:
															id_estado_solicitud = int(v)
														case int:
															id_estado_solicitud = v
														default:
															fmt.Println("Tipo inesperado en Id")
														}

														nombreEstado, ok := estado["Nombre"].(string)
														if !ok {
															fmt.Println("Nombre no válido")
															nombreEstado = "" // o algún valor por defecto
														}

														sol.EstadoSolicitud = &models.EstadoSolicitud{
															Id:     id_estado_solicitud,
															Nombre: nombreEstado,
														}
													}
												}
											}
										}
									}
									respuesta = append(respuesta, sol)
								}
							}

							return respuesta, nil
						}
					}
				}
			}
		}
	}

	return nil, map[string]interface{}{
		"error":  "no se encontró solicitud",
		"status": 404,
	}
}

func BuscarDetallesSolicitud(id_solicitud int) (respuesta models.SolicitudDetalles, outputError map[string]interface{}) {

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

	var respuesta_historico map[string]interface{}

	err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"historico_estado_solicitud?query=SolicitudId__Id:"+fmt.Sprintf("%d", id_solicitud)+
			",Activo:true&sortby=FechaCreacion&order=desc&limit=-1",
		&respuesta_historico,
	)

	if err != nil {
		return respuesta, nil
	}

	data, ok := respuesta_historico["Data"].([]interface{})
	if !ok || len(data) == 0 {
		return models.SolicitudDetalles{}, map[string]interface{}{
			"error":  "no se encontró solicitud",
			"status": 404,
		}
	}

	primer_registro, ok := data[0].(map[string]interface{})
	if !ok {
		return respuesta, nil
	}

	info_solicitud, ok := primer_registro["SolicitudId"].(map[string]interface{})
	if !ok {
		return respuesta, nil
	}

	if registro_tipo_solicitud, ok := info_solicitud["TipoSolicitudId"].(map[string]interface{}); ok {
		tipo_solicitud_historico := models.TipoSolicitud{
			Id:                int(registro_tipo_solicitud["Id"].(float64)),
			Nombre:            fmt.Sprintf("%v", registro_tipo_solicitud["Nombre"]),
			CodigoAbreviacion: fmt.Sprintf("%v", registro_tipo_solicitud["CodigoAbreviacion"]),
		}

		if estado_solicitud_actual, ok := primer_registro["EstadoSolicitudId"].(map[string]interface{}); ok {
			estado_solicitud_info := models.EstadoSolicitud{
				Id:                int(estado_solicitud_actual["Id"].(float64)),
				Nombre:            fmt.Sprintf("%v", estado_solicitud_actual["Nombre"]),
				Descripcion:       fmt.Sprintf("%v", estado_solicitud_actual["Descripcion"]),
				CodigoAbreviacion: fmt.Sprintf("%v", estado_solicitud_actual["CodigoAbreviacion"]),
			}
			respuesta.EstadoSolicitud = &estado_solicitud_info
		}

		solicitud_historico := models.Solicitud{
			Id:                int(info_solicitud["Id"].(float64)),
			TerceroId:         int(info_solicitud["TerceroId"].(float64)),
			TipoSolicitudId:   &tipo_solicitud_historico,
			ObservacionCierre: fmt.Sprintf("%v", info_solicitud["ObservacionCierre"]),
			Activo:            info_solicitud["Activo"].(bool),
		}
		respuesta.Solicitud = &solicitud_historico
	}

	var respuesta_detalle_formulario map[string]interface{}
	if err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+"detalle_solicitud?query=SolicitudId__Id:"+fmt.Sprintf("%d", id_solicitud),
		&respuesta_detalle_formulario,
	); err == nil {

		if data_formulario, ok := respuesta_detalle_formulario["Data"].([]interface{}); ok && len(data_formulario) > 0 {
			if registro_formulario, ok := data_formulario[0].(map[string]interface{}); ok {
				respuesta.Formulario = registro_formulario["Formulario"]
			}
		}
	}

	var respuesta_documentos map[string]interface{}

	if err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+"documento_solicitud?query=HistoricoEstadoSolicitudId__SolicitudId__Id:"+fmt.Sprintf("%d", id_solicitud),
		&respuesta_documentos,
	); err == nil {

		if data_documentos, ok := respuesta_documentos["Data"].([]interface{}); ok && len(data_documentos) > 0 {

			for _, doc := range data_documentos {
				if documento, ok := doc.(map[string]interface{}); ok {

					docId := int(documento["DocumentoId"].(float64))

					var detalle_doc map[string]interface{}
					if err := request.GetJson(
						beego.AppConfig.String("UrlDocumentos")+"documento/"+fmt.Sprintf("%d", docId),
						&detalle_doc,
					); err == nil {

						if len(detalle_doc) == 0 {
							continue
						}

						nombre, _ := detalle_doc["Nombre"].(string)
						enlace, _ := detalle_doc["Enlace"].(string)

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
							documento_aux := models.DocumentoDetalle{
								Nombre: nombre,
								Enlace: enlace,
								Tipo:   tipo,
								Estado: estado,
							}

							respuesta.Documentos = append(respuesta.Documentos, documento_aux)
						}
					}
				}
			}
		}
	}

	return respuesta, nil
}
