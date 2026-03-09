package helpers

import (
	"github.com/udistrital/comisiones_mid/models"
)

//Crear funcion de documentos, con entrada la estructura del documento y sale el id de creación

func CrearDocumento(documento models.Documento) (id_documento int) {
	url_documentos := "http://pruebasapi.intranetoas.udistrital.edu.co:8094/v1/"
	var respuesta_creacion map[string]interface{}	
	if status, err := PostJsonTest(url_documentos+"documento", documento, &respuesta_creacion); err == nil && (status == 200 || status == 201) {
		if id, ok := respuesta_creacion["Id"]; ok {
			switch v := id.(type) {
			case float64:
				return int(v)
			case int:
				return v
			}
		}
	}
	return 0
}
