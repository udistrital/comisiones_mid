package services_test

// Pruebas de caja negra para services/detalle_comision.go:
// ObtenerDetalleComision.

import (
	"errors"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/services"
	"github.com/udistrital/utils_oas/request"
)

// =====================================================
// TEST OBTENER DETALLE COMISION
// =====================================================

func TestObtenerDetalleComision(t *testing.T) {

	t.Run("Caso 1: UrlComisionesCrud no configurado retorna error", func(t *testing.T) {
		_ = beego.AppConfig.Set("UrlComisionesCrud", "")
		_, err := services.ObtenerDetalleComision(1)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "UrlComisionesCrud") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 2: solicitud no encontrada retorna error", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			// buscarSolicitudPorComisionId: Data vacío → retorna nil desde la función
			if strings.Contains(rawURL, "/solicitud?") {
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{},
				}
				return nil
			}
			return errors.New("url no esperada: " + rawURL)
		})

		_, err := services.ObtenerDetalleComision(10)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "solicitud no encontrada") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 3: error al buscar solicitud propaga error", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			if strings.Contains(rawURL, "/solicitud?") {
				return errors.New("servicio solicitud no disponible")
			}
			return errors.New("url no esperada: " + rawURL)
		})

		_, err := services.ObtenerDetalleComision(10)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "solicitud no encontrada") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 4: éxito con formulario válido retorna todos los campos del docente", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		formularioJSON := `{
			"solicitante": {
				"q3_nombres_apellidos": "Maria Garcia",
				"q4_documento_identificacion": "CC 12345678",
				"q6_correo": "maria@udistrital.edu.co",
				"q2_facultad": "Facultad de Ingenieria"
			},
			"solicitud": {
				"q14_nombre_programa": "Maestria en Ingenieria",
				"q13_tipo_estudio": "Maestria",
				"q16_universidad": "Universidad Nacional",
				"q17_pais": "Colombia",
				"q18_ciudad": "Bogota",
				"q20_num_semestres": "4"
			}
		}`

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			switch {
			case strings.Contains(rawURL, "/solicitud?"):
				// buscarSolicitudPorComisionId
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"Id":            float64(100),
							"FechaCreacion": "2024-01-15T00:00:00Z",
							"ComisionId": map[string]interface{}{
								"FechaInicio": "2024-06-01T00:00:00Z",
								"FechaFinal":  "2025-06-01T00:00:00Z",
							},
						},
					},
				}
				return nil
			case strings.Contains(rawURL, "detalle_solicitud"):
				// obtenerDetalleSolicitud
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"Formulario": formularioJSON,
						},
					},
				}
				return nil
			case strings.Contains(rawURL, "historico_estado_comision"):
				// obtenerEstadoComisionActivo
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"EstadoComisionId": map[string]interface{}{
								"CodigoAbreviacion": "COM_INI",
								"Nombre":            "Inicio",
							},
						},
					},
				}
				return nil
			}
			return errors.New("url no esperada: " + rawURL)
		})

		detalle, err := services.ObtenerDetalleComision(10)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo: %v", err)
		}

		if detalle.ComisionId != 10 {
			t.Errorf("se esperaba ComisionId=10 y se obtuvo %d", detalle.ComisionId)
		}
		if detalle.SolicitudId != 100 {
			t.Errorf("se esperaba SolicitudId=100 y se obtuvo %d", detalle.SolicitudId)
		}
		if detalle.Radicado != "SOL-100" {
			t.Errorf("se esperaba Radicado='SOL-100' y se obtuvo %q", detalle.Radicado)
		}
		if detalle.Docente != "Maria Garcia" {
			t.Errorf("se esperaba Docente='Maria Garcia' y se obtuvo %q", detalle.Docente)
		}
		// extraerSoloNumero("CC 12345678") → "12345678"
		if detalle.IdDocente != "12345678" {
			t.Errorf("se esperaba IdDocente='12345678' y se obtuvo %q", detalle.IdDocente)
		}
		if detalle.CorreoDocente != "maria@udistrital.edu.co" {
			t.Errorf("se esperaba CorreoDocente='maria@udistrital.edu.co' y se obtuvo %q", detalle.CorreoDocente)
		}
		if detalle.Facultad != "Facultad de Ingenieria" {
			t.Errorf("se esperaba Facultad='Facultad de Ingenieria' y se obtuvo %q", detalle.Facultad)
		}
		if detalle.Programa != "Maestria en Ingenieria" {
			t.Errorf("se esperaba Programa='Maestria en Ingenieria' y se obtuvo %q", detalle.Programa)
		}
		if detalle.TipoEstudio != "Maestria" {
			t.Errorf("se esperaba TipoEstudio='Maestria' y se obtuvo %q", detalle.TipoEstudio)
		}
		if detalle.UniversidadDestino != "Universidad Nacional" {
			t.Errorf("se esperaba UniversidadDestino='Universidad Nacional' y se obtuvo %q", detalle.UniversidadDestino)
		}
		if detalle.PaisDestino != "Colombia" {
			t.Errorf("se esperaba PaisDestino='Colombia' y se obtuvo %q", detalle.PaisDestino)
		}
		if detalle.CiudadDestino != "Bogota" {
			t.Errorf("se esperaba CiudadDestino='Bogota' y se obtuvo %q", detalle.CiudadDestino)
		}
		if detalle.Duracion != "4" {
			t.Errorf("se esperaba Duracion='4' y se obtuvo %q", detalle.Duracion)
		}
		if detalle.EstadoComision != "COM_INI" {
			t.Errorf("se esperaba EstadoComision='COM_INI' y se obtuvo %q", detalle.EstadoComision)
		}
		if detalle.FechaInicio == "" {
			t.Error("se esperaba FechaInicio no vacía")
		}
	})

	t.Run("Caso 5: sin detalle_solicitud disponible retorna campos docente vacíos pero sin error", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			switch {
			case strings.Contains(rawURL, "/solicitud?"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"Id":            float64(200),
							"FechaCreacion": "2024-03-10T00:00:00Z",
						},
					},
				}
				return nil
			case strings.Contains(rawURL, "detalle_solicitud"):
				// Data vacío → ObtenerDatosFormulario retorna error → campos docente quedan vacíos
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{},
				}
				return nil
			case strings.Contains(rawURL, "historico_estado_comision"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"EstadoComisionId": map[string]interface{}{
								"CodigoAbreviacion": "COM_INI",
							},
						},
					},
				}
				return nil
			}
			return errors.New("url no esperada: " + rawURL)
		})

		detalle, err := services.ObtenerDetalleComision(10)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo: %v", err)
		}
		if detalle.SolicitudId != 200 {
			t.Errorf("se esperaba SolicitudId=200 y se obtuvo %d", detalle.SolicitudId)
		}
		if detalle.Docente != "" {
			t.Errorf("se esperaba Docente vacío y se obtuvo %q", detalle.Docente)
		}
		if detalle.EstadoComision != "COM_INI" {
			t.Errorf("se esperaba EstadoComision='COM_INI' y se obtuvo %q", detalle.EstadoComision)
		}
	})
}
