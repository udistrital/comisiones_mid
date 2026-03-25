package services

import (
	"fmt"

	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/utils_oas/request"
)

func CrearSolicitud(solicitud models.CrearSolicitudEntrada) (respuesta models.Solicitud, outputError map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			outputError = map[string]interface{}{"funcion": "/CrearSolicitudService", "err": err, "status": "404"}
			panic(outputError)
		}
	}()
	// Llamar terceros para buscar datos
	//Busqueda de datos del tercero por numero de documento
	var persona []map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("UrlTercerosCrud")+"datos_identificacion?query=Numero:"+fmt.Sprintf("%d", solicitud.Identificacion), &persona); err == nil {
		if len(persona) > 0 && len(persona[0]) > 0 {
			if tercero, ok := persona[0]["TerceroId"].(map[string]interface{}); ok {
				if id_tercero, ok := tercero["Id"].(float64); ok {
					var respuesta_request models.SolicitudCreateRequest
					respuesta_request.TerceroId = int(id_tercero)
					fmt.Println("ENTRA A ASGINAR")
					respuesta_request.Activo = true
					respuesta_request.TipoSolicitudId = models.IdReference{
						Id: solicitud.TipoSolicitudId,
					}
					fmt.Println("respuesta ", respuesta_request)
					fmt.Println("tipo solicitud ", respuesta_request.TipoSolicitudId)
					var respuesta_creacion map[string]interface{}
					if err := request.SendJson(beego.AppConfig.String("UrlComisionesCrud")+"/solicitud", "POST", &respuesta_creacion, &respuesta_request); err == nil {
						fmt.Println("response", respuesta_creacion)
						if solicitud.Formulario != nil {
							formularioBytes, err := json.Marshal(solicitud.Formulario)
							if err != nil {
								fmt.Println("Error convirtiendo formulario a JSON:", err)
								return respuesta, map[string]interface{}{"error": err.Error()}
							}
							var solicitud_temp models.Solicitud
							if data, ok := respuesta_creacion["Data"].(map[string]interface{}); ok {
								solicitud_temp.Id = int(data["Id"].(float64))
								detalles_solicitud := models.DetalleSolicitud{
									SolicitudId: &solicitud_temp,
									Formulario:  string(formularioBytes),
									Activo:      true,
								}
								var id_estado int
								if solicitud.Formulario["formulario_completado"] == false || len(solicitud.DocumentoSolicitud) == 0 || solicitud.Observacion == "" {
									id_estado = 1
								} else {
									id_estado = 2
								}

								historico_solicitud := models.HistoricoEstadoSolicitud{
									SolicitudId:       &solicitud_temp,
									EstadoSolicitudId: &models.EstadoSolicitud{Id: id_estado},
									RolUsuario:        solicitud.CodigoAbreviacionRol,
									TerceroId:         int(id_tercero),
									Activo:            true,
								}
								fmt.Println("historico_solicitud", historico_solicitud)
								var respuesta_detalle_solicitud map[string]interface{}
								var respuesta_historico_estado_solicitud map[string]interface{}
								fmt.Println("CREO LA SOLICITUD")
								if err := request.SendJson(beego.AppConfig.String("UrlComisionesCrud")+"detalle_solicitud", "POST", &respuesta_detalle_solicitud, &detalles_solicitud); err == nil {
									fmt.Println("SE CREA LA SOLICITUD CON FORMULARIO")
									if err := request.SendJson(beego.AppConfig.String("UrlComisionesCrud")+"historico_estado_solicitud", "POST", &respuesta_historico_estado_solicitud, &historico_solicitud); err == nil {
										if solicitud.Formulario["formulario_completado"] == false || len(solicitud.DocumentoSolicitud) == 0 || solicitud.Observacion == "" {
											if data_estado_historico, ok := respuesta_historico_estado_solicitud["Data"].(map[string]interface{}); ok {
												var id_historico_estado int
												if id, ok := data_estado_historico["Id"].(float64); ok {
													id_historico_estado = int(id)
												}
												if len(solicitud.DocumentoSolicitud) != 0 {
													var resultado_documentos []map[string]interface{}
													var errDoc map[string]interface{}
													if resultado_documentos, errDoc = helpers.CrearDocumento(solicitud.DocumentoSolicitud); errDoc == nil {
														for _, doc := range resultado_documentos {
															fmt.Println("DOCUMENTOS CREADOS")
															fmt.Println(doc)
															var idDoc int
															switch v := doc["id"].(type) {
															case float64:
																idDoc = int(v)
															case int:
																idDoc = v
															default:
																fmt.Println("ERROR: tipo inesperado en id", v)
																continue
															}
															documento_solicitud := models.DocumentoSolicitud{
																DocumentoId:           idDoc,
																SolicitudEstadoEvento: &models.HistoricoEstadoSolicitud{Id: id_historico_estado},
																TipoDocumento:         &models.TipoDocumentoSolicitud{Id: 1},
																EstadoDocumento:       &models.EstadoDocumento{Id: 1},
																Activo:                true,
															}
															var respuesta_documento_solicitud map[string]interface{}
															if err := request.SendJson(beego.AppConfig.String("UrlComisionesCrud")+"documento_solicitud", "POST", &respuesta_documento_solicitud, &documento_solicitud); err == nil {
																fmt.Println("documento creado y anexado a la solicitud")
															}

														}
													}
												}
											}
										}
										fmt.Println("SE CREA EL HISTORICO")
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return respuesta, outputError
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
					if err := request.GetJson(beego.AppConfig.String("UrlComisionesCrud")+"solicitud?query=TerceroId:"+fmt.Sprintf("%d", id_tercero_busqueda),
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
