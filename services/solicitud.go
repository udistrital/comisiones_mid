package services

import (
	"fmt"

	"encoding/json"
	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
	// "github.com/udistrital/utils_oas/request"
)

func CrearSolicitud(solicitud models.CrearSolicitudEntrada) (respuesta models.SolicitudInicial, outputError map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			outputError = map[string]interface{}{"funcion": "/CrearSolicitudService", "err": err, "status": "404"}
			panic(outputError)
		}
	}()
	terceros_url := "http://pruebasapi.intranetoas.udistrital.edu.co:8121/v1/"
	comision_crud := "http://localhost:8080/v1/"
	// Llamar terceros para buscar datos
	//Busqueda de datos del tercero por numero de documento
	var persona []map[string]interface{}
	if response, err := helpers.GetJsonTest(terceros_url+"datos_identificacion?query=Numero:"+fmt.Sprintf("%d", solicitud.Identificacion), &persona); (err == nil) && (response == 200) {
		if len(persona) > 0 && len(persona[0]) > 0 {
			if tercero, ok := persona[0]["TerceroId"].(map[string]interface{}); ok {
				if id, ok := tercero["Id"].(float64); ok {
					respuesta.TerceroId = int(id)
					fmt.Println("ENTRA A ASGINAR")
					respuesta.Activo = true
					respuesta.TipoSolicitudId = &models.TipoSolicitud{
						Id: solicitud.TipoSolicitudId,
					}
					fmt.Println(respuesta)
					var respuesta_creacion models.ResponseSolicitud
					if response, err := helpers.PostJsonTest(comision_crud+"solicitud", &respuesta, &respuesta_creacion); (err == nil) && (response == 201) {
						print(response)
						if solicitud.Formulario != nil {
							formularioBytes, err := json.Marshal(solicitud.Formulario)
							if err != nil {
								fmt.Println("Error convirtiendo formulario a JSON:", err)
								return respuesta, map[string]interface{}{"error": err.Error()}
							}
							fmt.Println(respuesta_creacion.Data.Id)
							var solicitud_temp models.SolicitudInicial
							solicitud_temp.Id = respuesta_creacion.Data.Id
							detalles_solicitud := models.DetalleSolicitud{
								SolicitudId: &solicitud_temp,
								Formulario:  string(formularioBytes),
								Activo:      true,
							}
							var respuesta_detalle_solicitud models.DetalleSolicitud
							fmt.Println("CREO LA SOLICITUD")
							if response, err := helpers.PostJsonTest(comision_crud+"detalle_solicitud", &detalles_solicitud, &respuesta_detalle_solicitud); (err == nil) && (response == 201) {
								fmt.Println("SE CREA LA SOLICITUD CON FORMULARIO")
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

	comision_crud := "http://localhost:8080/v1/"

	var persona map[string]interface{}

	if response, err := helpers.GetJsonTest(
		comision_crud+"solicitud?query=TerceroId:"+fmt.Sprintf("%d", identificacion),
		&persona,
	); err == nil && response == 200 {

		// 🔥 acceder a Data
		if data, ok := persona["Data"].([]interface{}); ok && len(data) > 0 {

			for _, item := range data {

				if itemMap, ok := item.(map[string]interface{}); ok {

					var sol models.SolicitudResumen

					if id, ok := itemMap["Id"].(float64); ok {
						sol.Id = int(id)
					}

					if activo, ok := itemMap["Activo"].(bool); ok {
						sol.Activo = activo
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
