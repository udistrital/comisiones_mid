package helpers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/models"
	"strconv"
)

func CrearDocumento(documentos []models.CrearDocumentoGestorDocumental) (resultado []map[string]interface{}, outputError map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			outputError = map[string]interface{}{"funcion": "/CrearDocumento", "err": err, "status": "404"}
			panic(outputError)
		}
	}()

	var respuesta_creacion map[string]interface{}
	resultado = []map[string]interface{}{}

	if status, err := PostJsonTest(
		beego.AppConfig.String("UrlGestorDocumental")+"document/upload",
		documentos,
		&respuesta_creacion,
	); err == nil && (status == 200 || status == 201) {
		if res, ok := respuesta_creacion["res"]; ok {
			switch docs := res.(type) {
			// varios documentos
			case []interface{}:
				for _, d := range docs {
					doc := d.(map[string]interface{})

					resultado = append(resultado, map[string]interface{}{
						"id":          int(doc["Id"].(float64)),
						"nombre":      doc["Nombre"].(string),
						"descripcion": doc["Descripcion"].(string),
						"enlace":      doc["Enlace"].(string),
					})
				}

			// un solo documento
			case map[string]interface{}:
				resultado = append(resultado, map[string]interface{}{
					"id":          int(docs["Id"].(float64)),
					"nombre":      docs["Nombre"].(string),
					"descripcion": docs["Descripcion"].(string),
					"enlace":      docs["Enlace"].(string),
				})
			}
		}

		return resultado, outputError
	}

	return resultado, outputError
}

func ObtenerDatosFormulario(detalleSolicitud map[string]interface{}) (datos models.Formulario, outputError map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			outputError = map[string]interface{}{
				"funcion": "/ObtenerDatosFormulario",
				"err":     err,
				"status":  "404",
			}
			panic(outputError)
		}
	}()

	if data, ok := detalleSolicitud["Data"].([]interface{}); ok && len(data) > 0 {
		if itemMap, ok := data[0].(map[string]interface{}); ok {
			if formularioStr, ok := itemMap["Formulario"].(string); ok {

				if err := json.Unmarshal([]byte(formularioStr), &datos); err != nil {
					outputError = map[string]interface{}{
						"funcion": "/ObtenerDatosFormulario",
						"err":     err.Error(),
						"status":  "404",
					}
					return datos, outputError
				}

				return datos, nil
			}
		}
	}

	outputError = map[string]interface{}{
		"funcion": "/ObtenerDatosFormulario",
		"err":     "no se encontró la información de Formulario",
		"status":  "404",
	}
	return datos, outputError
}

func ValidarRespuesta(resp map[string]interface{}) (map[string]interface{}, map[string]interface{}) {

	success, _ := resp["Success"].(bool)

	var status int
	switch v := resp["Status"].(type) {
	case float64:
		status = int(v)
	case string:
		status, _ = strconv.Atoi(v)
	default:
		status = 500
	}

	if !success || status >= 400 {
		return nil, map[string]interface{}{
			"Success": false,
			"Status":  status,
			"Message": "Error en servicio externo",
			"Data":    resp["Data"],
			"Raw":     resp, // 🔥 útil para debug
		}
	}

	data, ok := resp["Data"].(map[string]interface{})
	if !ok || data == nil {
		return nil, map[string]interface{}{
			"Success": false,
			"Status":  502,
			"Message": "Respuesta sin data",
			"Data":    nil,
		}
	}

	return data, nil
}