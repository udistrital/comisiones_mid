package services_test

// Pruebas de caja negra para services/solicitud.go:
// ExtraerIdRelacionado, DocumentosADesactivar, BuscarSolicitudIdentificacion, BuscarDetallesSolicitud.

import (
	"errors"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/comisiones_mid/services"
	"github.com/udistrital/utils_oas/request"
)

// TestExtraerIdRelacionado verifica la extracción del Id numérico desde distintos tipos de valor
func TestExtraerIdRelacionado(t *testing.T) {
	t.Run("Caso 1: objeto map con clave Id en float64", func(t *testing.T) {
		obj := map[string]interface{}{"Id": float64(99)}
		resultado := services.ExtraerIdRelacionado(obj)
		if resultado != 99 {
			t.Errorf("se esperaba 99 y se obtuvo %d", resultado)
		}
	})

	t.Run("Caso 2: valor float64 directo retorna el entero correspondiente", func(t *testing.T) {
		resultado := services.ExtraerIdRelacionado(float64(42))
		if resultado != 42 {
			t.Errorf("se esperaba 42 y se obtuvo %d", resultado)
		}
	})

	t.Run("Caso 3: valor nil o tipo desconocido retorna 0", func(t *testing.T) {
		if services.ExtraerIdRelacionado(nil) != 0 {
			t.Error("se esperaba 0 para nil")
		}
		if services.ExtraerIdRelacionado("cadena") != 0 {
			t.Error("se esperaba 0 para string")
		}
	})
}

// TestDocumentosADesactivar verifica que se filtran ceros y se eliminan IDs duplicados
func TestDocumentosADesactivar(t *testing.T) {
	t.Run("Caso 1: lista limpia sin duplicados ni ceros devuelve todos los IDs", func(t *testing.T) {
		req := models.EditarSolicitud{
			DocumentosDesactivar: []int{1, 2, 3},
		}
		resultado := services.DocumentosADesactivar(req)
		if len(resultado) != 3 {
			t.Errorf("se esperaban 3 elementos y se obtuvo %d", len(resultado))
		}
	})

	t.Run("Caso 2: lista con duplicados y ceros devuelve solo IDs válidos únicos", func(t *testing.T) {
		req := models.EditarSolicitud{
			DocumentosDesactivar: []int{5, 0, 5, 3, 0},
		}
		resultado := services.DocumentosADesactivar(req)
		if len(resultado) != 2 {
			t.Errorf("se esperaban 2 elementos y se obtuvo %d: %v", len(resultado), resultado)
		}
		if resultado[0] != 5 || resultado[1] != 3 {
			t.Errorf("se esperaba [5, 3] y se obtuvo %v", resultado)
		}
	})
}

// TestBuscarSolicitudIdentificacion cubre el flujo: tercero → solicitudes → detalle → historico
func TestBuscarSolicitudIdentificacion(t *testing.T) {
	t.Run("Caso 1: retorna solicitudes del docente con nombre y programa", func(t *testing.T) {
		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlTercerosCrud", "http://terceros/")
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			switch {
			case strings.Contains(rawURL, "datos_identificacion"):
				*(target.(*[]map[string]interface{})) = []map[string]interface{}{
					{"TerceroId": map[string]interface{}{"Id": float64(42)}},
				}
				return nil

			case strings.Contains(rawURL, "TerceroId"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"Id":            float64(10),
							"Activo":        true,
							"FechaCreacion": "2024-01-15T00:00:00Z",
						},
					},
				}
				return nil

			case strings.Contains(rawURL, "detalle_solicitud"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"Formulario": `{"solicitante":{"q3_nombres_apellidos":"Juan Perez","q7_proyecto":"Ingenieria de Sistemas"}}`,
						},
					},
				}
				return nil

			case strings.Contains(rawURL, "historico_estado_solicitud"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"EstadoSolicitudId": map[string]interface{}{
								"Id":     float64(3),
								"Nombre": "En revisión",
							},
						},
					},
				}
				return nil
			}

			return errors.New("URL no esperada: " + rawURL)
		})

		resultado, errMap := services.BuscarSolicitudIdentificacion(123456)
		if errMap != nil {
			t.Fatalf("no se esperaba error y se obtuvo: %v", errMap)
		}
		if len(resultado) == 0 {
			t.Fatal("se esperaba al menos una solicitud en el resultado")
		}
		if resultado[0].Id != 10 {
			t.Errorf("se esperaba Id 10 y se obtuvo %d", resultado[0].Id)
		}
		if resultado[0].Nombre != "Juan Perez" {
			t.Errorf("se esperaba nombre Juan Perez y se obtuvo %s", resultado[0].Nombre)
		}
		if resultado[0].Programa != "Ingenieria de Sistemas" {
			t.Errorf("se esperaba programa Ingenieria de Sistemas y se obtuvo %s", resultado[0].Programa)
		}
	})

	t.Run("Caso 2: error en terceros retorna error de no encontrado", func(t *testing.T) {
		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlTercerosCrud", "http://terceros/")
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			return errors.New("terceros no disponible")
		})

		resultado, errMap := services.BuscarSolicitudIdentificacion(123456)
		if errMap == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if resultado != nil {
			t.Errorf("se esperaba resultado nil y se obtuvo: %v", resultado)
		}
	})
}

// TestBuscarDetallesSolicitud cubre: historico → formulario → documentos → observaciones
func TestBuscarDetallesSolicitud(t *testing.T) {
	t.Run("Caso 1: retorna detalle completo con solicitud y estado", func(t *testing.T) {
		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")
		_ = beego.AppConfig.Set("UrlDocumentos", "http://documentos/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			switch {
			case strings.Contains(rawURL, "historico_estado_solicitud"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"SolicitudId": map[string]interface{}{
								"Id":                float64(5),
								"TerceroId":         float64(42),
								"ObservacionCierre": "",
								"Activo":            true,
								"TipoSolicitudId": map[string]interface{}{
									"Id":                float64(1),
									"Nombre":            "Comisión de Estudio",
									"CodigoAbreviacion": "COM_EST",
								},
							},
							"EstadoSolicitudId": map[string]interface{}{
								"Id":                float64(3),
								"Nombre":            "En revisión",
								"Descripcion":       "Solicitud en revisión por coordinador",
								"CodigoAbreviacion": "REV_PROY",
							},
						},
					},
				}
				return nil

			case strings.Contains(rawURL, "detalle_solicitud"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{"Formulario": `{}`},
					},
				}
				return nil

			case strings.Contains(rawURL, "documento_solicitud"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{},
				}
				return nil

			case strings.Contains(rawURL, "observacion"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{},
				}
				return nil
			}

			return errors.New("URL no esperada: " + rawURL)
		})

		respuesta, errMap := services.BuscarDetallesSolicitud(5)
		if errMap != nil {
			t.Fatalf("no se esperaba error y se obtuvo: %v", errMap)
		}
		if respuesta.Solicitud == nil {
			t.Fatal("se esperaba Solicitud no nil")
		}
		if respuesta.Solicitud.Id != 5 {
			t.Errorf("se esperaba Solicitud.Id 5 y se obtuvo %d", respuesta.Solicitud.Id)
		}
		if respuesta.EstadoSolicitud == nil {
			t.Fatal("se esperaba EstadoSolicitud no nil")
		}
		if respuesta.EstadoSolicitud.Id != 3 {
			t.Errorf("se esperaba EstadoSolicitud.Id 3 y se obtuvo %d", respuesta.EstadoSolicitud.Id)
		}
	})

	t.Run("Caso 2: historico sin datos retorna error de solicitud no encontrada", func(t *testing.T) {
		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			if strings.Contains(rawURL, "historico_estado_solicitud") {
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{},
				}
				return nil
			}
			return errors.New("URL no esperada: " + rawURL)
		})

		respuesta, errMap := services.BuscarDetallesSolicitud(999)
		if errMap == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if errMap["error"] != "no se encontró solicitud" {
			t.Errorf("se esperaba 'no se encontró solicitud' y se obtuvo: %v", errMap)
		}
		if respuesta.Solicitud != nil {
			t.Errorf("se esperaba Solicitud nil y se obtuvo: %v", respuesta.Solicitud)
		}
	})
}
