package services_test

// Pruebas de las 4 funciones exportadas en services/cancelarsolicitud.go:
// ObtenerIdsPorQuery, DesactivarRecursoPorId, DesactivarSolicitud, CancelarSolicitud.

import (
	"errors"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/services"
	"github.com/udistrital/utils_oas/request"
)

// TestObtenerIdsPorQuery verifica que se extraen IDs de la respuesta del CRUD
func TestObtenerIdsPorQuery(t *testing.T) {
	t.Run("Caso 1: retorna lista de IDs cuando el CRUD responde correctamente", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			*(target.(*map[string]interface{})) = map[string]interface{}{
				"Data": []interface{}{
					map[string]interface{}{"Id": float64(11)},
					map[string]interface{}{"Id": float64(22)},
				},
			}
			return nil
		})

		ids, err := services.ObtenerIdsPorQuery("http://comisiones/", "detalle_solicitud", "SolicitudId:5,Activo:true")
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo %v", err)
		}
		if len(ids) != 2 {
			t.Errorf("se esperaban 2 ids y se obtuvieron %d", len(ids))
		}
		if ids[0] != 11 || ids[1] != 22 {
			t.Errorf("se esperaban [11, 22] y se obtuvieron %v", ids)
		}
	})

	t.Run("Caso 2: error de GetJson retorna error envuelto", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			return errors.New("servicio no disponible")
		})

		ids, err := services.ObtenerIdsPorQuery("http://comisiones/", "detalle_solicitud", "SolicitudId:5,Activo:true")
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "error consultando detalle_solicitud") {
			t.Errorf("se esperaba mensaje de error de consulta y se obtuvo: %v", err)
		}
		if ids != nil {
			t.Errorf("se esperaba ids nil y se obtuvo %v", ids)
		}
	})
}

// TestDesactivarRecursoPorId verifica el patrón GET + PUT con Activo=false
func TestDesactivarRecursoPorId(t *testing.T) {
	t.Run("Caso 1: desactiva el recurso con PUT Activo=false", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			*(target.(*map[string]interface{})) = map[string]interface{}{
				"Data": map[string]interface{}{
					"Id":     float64(7),
					"Activo": true,
				},
			}
			return nil
		})

		monkey.Patch(request.SendJson, func(rawURL, method string, target interface{}, body interface{}) error {
			if method != "PUT" {
				t.Fatalf("se esperaba PUT y se obtuvo %s", method)
			}
			payload := body.(map[string]interface{})
			if payload["Activo"] != false {
				t.Fatalf("se esperaba Activo false y se obtuvo %v", payload["Activo"])
			}
			*(target.(*map[string]interface{})) = map[string]interface{}{}
			return nil
		})

		err := services.DesactivarRecursoPorId("http://comisiones/", "observacion", 7)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo %v", err)
		}
	})

	t.Run("Caso 2: error en GET retorna error con nombre del recurso", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			return errors.New("servicio caido")
		})

		err := services.DesactivarRecursoPorId("http://comisiones/", "observacion", 7)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "error GET observacion") {
			t.Errorf("se esperaba 'error GET observacion' y se obtuvo: %v", err)
		}
	})
}

// TestDesactivarSolicitud verifica el soft delete de la solicitud principal
func TestDesactivarSolicitud(t *testing.T) {
	t.Run("Caso 1: desactiva la solicitud correctamente via GET + PUT", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			*(target.(*map[string]interface{})) = map[string]interface{}{
				"Data": map[string]interface{}{
					"Id":     float64(10),
					"Activo": true,
				},
			}
			return nil
		})

		monkey.Patch(request.SendJson, func(rawURL, method string, target interface{}, body interface{}) error {
			payload := body.(map[string]interface{})
			if payload["Activo"] != false {
				t.Fatalf("se esperaba Activo false y se obtuvo %v", payload["Activo"])
			}
			*(target.(*map[string]interface{})) = map[string]interface{}{}
			return nil
		})

		err := services.DesactivarSolicitud("http://comisiones/", 10)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo %v", err)
		}
	})

	t.Run("Caso 2: error en PUT retorna error descriptivo", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			*(target.(*map[string]interface{})) = map[string]interface{}{
				"Data": map[string]interface{}{"Id": float64(10), "Activo": true},
			}
			return nil
		})

		monkey.Patch(request.SendJson, func(rawURL, method string, target interface{}, body interface{}) error {
			return errors.New("fallo de red")
		})

		err := services.DesactivarSolicitud("http://comisiones/", 10)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "error PUT solicitud") {
			t.Errorf("se esperaba 'error PUT solicitud' y se obtuvo: %v", err)
		}
	})
}

// TestCancelarSolicitud cubre la orquestación completa de cancelación
func TestCancelarSolicitud(t *testing.T) {
	t.Run("Caso 1: solicitudId <= 0 retorna error de validacion", func(t *testing.T) {
		resp, err := services.CancelarSolicitud(0)
		if err == nil {
			t.Fatal("se esperaba error por solicitudId invalido")
		}
		if !strings.Contains(err.Error(), "solicitudId es obligatorio") {
			t.Errorf("se esperaba error de obligatorio y se obtuvo: %v", err)
		}
		if resp.SolicitudId != 0 {
			t.Errorf("se esperaba respuesta vacia y se obtuvo: %+v", resp)
		}
	})

	t.Run("Caso 2: cancelacion exitosa sin detalles ni historicos asociados", func(t *testing.T) {
		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(services.DesactivarSolicitud, func(baseCrud string, solicitudId int) error {
			return nil
		})
		monkey.Patch(services.ObtenerIdsPorQuery, func(baseCrud, recurso, query string) ([]int, error) {
			return []int{}, nil
		})

		resp, err := services.CancelarSolicitud(5)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo: %v", err)
		}
		if !resp.SolicitudDesactivada {
			t.Error("se esperaba SolicitudDesactivada = true")
		}
		if resp.SolicitudId != 5 {
			t.Errorf("se esperaba SolicitudId 5 y se obtuvo %d", resp.SolicitudId)
		}
		if len(resp.HistoricosDesactivados) != 0 {
			t.Errorf("se esperaban 0 historicos desactivados y se obtuvieron %d", len(resp.HistoricosDesactivados))
		}
	})
}
