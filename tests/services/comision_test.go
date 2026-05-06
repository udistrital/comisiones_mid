package services_test

// Pruebas de las funciones exportadas en services/comision.go usadas en el flujo APROB_EJEC:
// ExtraerComisionIdDesdeSolicitud, ExtraerIdRelacion, ConfirmarHistoricoEstadoComision, GetFechaCreacionSolicitud.

import (
	"errors"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/udistrital/comisiones_mid/services"
	"github.com/udistrital/utils_oas/request"
)

// TestExtraerComisionIdDesdeSolicitud verifica que se extrae el ComisionId del objeto solicitud
func TestExtraerComisionIdDesdeSolicitud(t *testing.T) {
	t.Run("Caso 1: extrae el id cuando ComisionId es un objeto anidado", func(t *testing.T) {
		obj := map[string]interface{}{
			"ComisionId": map[string]interface{}{"Id": float64(77)},
		}
		resultado := services.ExtraerComisionIdDesdeSolicitud(obj)
		if resultado != 77 {
			t.Errorf("se esperaba 77 y se obtuvo %d", resultado)
		}
	})

	t.Run("Caso 2: retorna 0 si ComisionId es nil (solicitud sin comision aun)", func(t *testing.T) {
		obj := map[string]interface{}{
			"ComisionId": nil,
		}
		resultado := services.ExtraerComisionIdDesdeSolicitud(obj)
		if resultado != 0 {
			t.Errorf("se esperaba 0 para ComisionId nil y se obtuvo %d", resultado)
		}
	})
}

// TestExtraerIdRelacion verifica la extracción de Id desde relaciones anidadas o valores directos
func TestExtraerIdRelacion(t *testing.T) {
	t.Run("Caso 1: extrae el id desde map con clave Id", func(t *testing.T) {
		v := map[string]interface{}{"Id": float64(55)}
		resultado := services.ExtraerIdRelacion(v)
		if resultado != 55 {
			t.Errorf("se esperaba 55 y se obtuvo %d", resultado)
		}
	})

	t.Run("Caso 2: retorna 0 si el valor es nil", func(t *testing.T) {
		resultado := services.ExtraerIdRelacion(nil)
		if resultado != 0 {
			t.Errorf("se esperaba 0 para nil y se obtuvo %d", resultado)
		}
	})
}

// TestConfirmarHistoricoEstadoComision verifica que el histórico creado coincide con la comisión esperada
func TestConfirmarHistoricoEstadoComision(t *testing.T) {
	t.Run("Caso 1: confirma correctamente cuando el historico consultado coincide", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			*(target.(*map[string]interface{})) = map[string]interface{}{
				"Data": map[string]interface{}{
					"Id":         float64(33),
					"ComisionId": map[string]interface{}{"Id": float64(10)},
				},
			}
			return nil
		})

		err := services.ConfirmarHistoricoEstadoComision("http://comisiones/", 33, 10)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo %v", err)
		}
	})

	t.Run("Caso 2: error al consultar el historico devuelve error descriptivo", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			return errors.New("servicio caido")
		})

		err := services.ConfirmarHistoricoEstadoComision("http://comisiones/", 33, 10)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "error consultando histórico de comisión") {
			t.Errorf("se esperaba error de consulta y se obtuvo: %v", err)
		}
	})
}

// TestGetFechaCreacionSolicitud verifica que se extrae la FechaCreacion desde el CRUD
func TestGetFechaCreacionSolicitud(t *testing.T) {
	t.Run("Caso 1: retorna la fecha de creacion cuando el CRUD responde", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			*(target.(*map[string]interface{})) = map[string]interface{}{
				"Data": []interface{}{
					map[string]interface{}{
						"Id":            float64(10),
						"FechaCreacion": "2024-01-15T10:00:00Z",
					},
				},
			}
			return nil
		})

		fecha, err := services.GetFechaCreacionSolicitud("http://comisiones/", 10)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo %v", err)
		}
		if fecha != "2024-01-15T10:00:00Z" {
			t.Errorf("se esperaba '2024-01-15T10:00:00Z' y se obtuvo '%s'", fecha)
		}
	})

	t.Run("Caso 2: Data vacio retorna error de solicitud no encontrada", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			*(target.(*map[string]interface{})) = map[string]interface{}{
				"Data": []interface{}{},
			}
			return nil
		})

		fecha, err := services.GetFechaCreacionSolicitud("http://comisiones/", 999)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "no se encontró la solicitud") {
			t.Errorf("se esperaba error de no encontrado y se obtuvo: %v", err)
		}
		if fecha != "" {
			t.Errorf("se esperaba fecha vacia y se obtuvo '%s'", fecha)
		}
	})
}
