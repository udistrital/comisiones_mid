package services

import (
	"fmt"

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
	fmt.Println(solicitud.TipoSolicitudId)
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
					var respuesta_creacion []map[string]interface{}
					if response, err := helpers.PostJsonTest(comision_crud+"solicitud", &respuesta, respuesta_creacion); (err == nil) && (response == 200) {
						fmt.Println(response)
						fmt.Println(err)
					}
				}
			}
		}
	}

	return respuesta, outputError
}
