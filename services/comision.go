package services

import (
	"fmt"

	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
	// "github.com/udistrital/utils_oas/request"
)

func CrearSolicitud(solicitud models.CrearComisionEntrada) (respuesta map[string]interface{}, outputError map[string]interface{}){
	defer func() {
		if err := recover(); err != nil {
			outputError = map[string]interface{}{"funcion": "/CrearSolicitudService", "err": err, "status": "404"}
			panic(outputError)
		}
	}()
	terceros_url := "http://pruebasapi.intranetoas.udistrital.edu.co:8121/v1/"
	// Llamar terceros para buscar datos
	//Busqueda de datos del tercero por numero de documento
	var persona []map[string]interface{}
	fmt.Println("ENTRA A BUSCAR")
	url := terceros_url + "datos_identificacion?query=Numero:"+ solicitud.Tercero_id
	fmt.Println(url)
	if response, err := helpers.GetJsonTest(url, &persona); (err == nil) && (response == 200){
		fmt.Println("encuentra")
		fmt.Println(response)
		fmt.Println(persona)
	}
	// Llamar a la api de documentos 
	return 
}