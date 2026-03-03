package services

import (
	"github.com/udistrital/comisiones_mid/models"
)

func CrearSolicitud(solicitud models.CrearComisionEntrada) (respuesta map[string]interface{}, outputError map[string]interface{}){
	defer func() {
		if err := recover(); err != nil {
			outputError = map[string]interface{}{"funcion": "/CrearSolicitudService", "err": err, "status": "404"}
			panic(outputError)
		}
	}()
	// Llamar terceros para buscar datos
	// Llamar a la api de documentos 
	return 
}