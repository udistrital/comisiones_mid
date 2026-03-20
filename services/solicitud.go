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

								historico_solicitud := models.HistoricoEstadoSolicitud{
									SolicitudId:       &solicitud_temp,
									EstadoSolicitudId: &models.EstadoSolicitud{Id: 2},
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
								}
								if err := request.SendJson(beego.AppConfig.String("UrlComisionesCrud")+"historico_estado_solicitud", "POST", &respuesta_historico_estado_solicitud, &historico_solicitud); err == nil {
									fmt.Println("SE CREA EL HISTORICO")
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

	var persona map[string]interface{}

	fmt.Println("ENTRA A SERVICIO ", beego.AppConfig.String("UrlComisionesCrud")+"solicitud?query=TerceroId:"+fmt.Sprintf("%d", identificacion))
	if err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+"solicitud?query=TerceroId:"+fmt.Sprintf("%d", identificacion),
		&persona,
	); err == nil {
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
					}
					respuesta = append(respuesta, sol)
				}
			}

			return respuesta, nil
		}
	}

	return nil, map[string]interface{}{
		"error":  "no se encontró solicitud",
		"status": 404,
	}
}
