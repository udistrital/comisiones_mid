package services_test

// Pruebas de caja negra para services/cumplimiento.go:
// ObtenerEstadosCumplimiento, ObtenerHistorialCumplimiento
// y CrearRegistroCumplimiento.

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

// =====================================================
// TEST OBTENER ESTADOS CUMPLIMIENTO
// =====================================================

func TestObtenerEstadosCumplimiento(t *testing.T) {
	t.Run("Caso 1: exito retorna solamente estados de cumplimiento", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {
				if !strings.Contains(rawURL, "estado_comision") {
					return errors.New("url no esperada: " + rawURL)
				}

				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"Id":                float64(1),
							"Nombre":            "Cumplimiento total",
							"CodigoAbreviacion": "CUMP_TOTAL",
						},
						map[string]interface{}{
							"Id":                float64(2),
							"Nombre":            "Incumplimiento parcial",
							"CodigoAbreviacion": "INCUMP_PARCIAL",
						},
						map[string]interface{}{
							"Id":                float64(3),
							"Nombre":            "Comision iniciada",
							"CodigoAbreviacion": "COM_INI",
						},
					},
				}
				return nil
			},
		)

		estados, err := services.ObtenerEstadosCumplimiento()
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo: %v", err)
		}

		if len(estados) != 2 {
			t.Fatalf(
				"se esperaban 2 estados de cumplimiento y se obtuvieron %d",
				len(estados),
			)
		}

		if estados[0].Id != 1 ||
			estados[0].Codigo != "CUMP_TOTAL" ||
			estados[0].Nombre != "Cumplimiento total" {
			t.Errorf("primer estado inesperado: %+v", estados[0])
		}

		if estados[1].Id != 2 ||
			estados[1].Codigo != "INCUMP_PARCIAL" {
			t.Errorf("segundo estado inesperado: %+v", estados[1])
		}
	})

	t.Run("Caso 2: fallo del CRUD retorna error esperado", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {
				return errors.New("servicio no disponible")
			},
		)

		_, err := services.ObtenerEstadosCumplimiento()
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if !strings.Contains(
			err.Error(),
			"error consultando estado_comision",
		) {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})
}

// =====================================================
// TEST OBTENER HISTORIAL CUMPLIMIENTO
// =====================================================

func TestObtenerHistorialCumplimiento(t *testing.T) {
	t.Run("Caso 1: exito retorna solamente registros de cumplimiento", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {
				if !strings.Contains(
					rawURL,
					"historico_estado_comision",
				) {
					return errors.New("url no esperada: " + rawURL)
				}

				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"Id":            float64(40),
							"Descripcion":   "El docente cumplio los compromisos",
							"FechaCreacion": "2026-06-10T10:00:00Z",
							"Activo":        true,
							"EstadoComisionId": map[string]interface{}{
								"Id":                float64(7),
								"Nombre":            "Cumplimiento total",
								"CodigoAbreviacion": "CUMP_TOTAL",
							},
						},
						map[string]interface{}{
							"Id":            float64(39),
							"Descripcion":   "Comision iniciada",
							"FechaCreacion": "2026-06-01T10:00:00Z",
							"Activo":        false,
							"EstadoComisionId": map[string]interface{}{
								"Id":                float64(2),
								"Nombre":            "Comision iniciada",
								"CodigoAbreviacion": "COM_INI",
							},
						},
					},
				}
				return nil
			},
		)

		historial, err := services.ObtenerHistorialCumplimiento(10)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo: %v", err)
		}

		if len(historial) != 1 {
			t.Fatalf(
				"se esperaba 1 registro de cumplimiento y se obtuvieron %d",
				len(historial),
			)
		}

		registro := historial[0]

		if registro.Id != 40 {
			t.Errorf("se esperaba Id=40 y se obtuvo %d", registro.Id)
		}

		if registro.EstadoId != 7 {
			t.Errorf(
				"se esperaba EstadoId=7 y se obtuvo %d",
				registro.EstadoId,
			)
		}

		if registro.EstadoCodigo != "CUMP_TOTAL" {
			t.Errorf(
				"se esperaba EstadoCodigo=CUMP_TOTAL y se obtuvo %q",
				registro.EstadoCodigo,
			)
		}

		if !registro.Activo {
			t.Error("se esperaba que el registro estuviera activo")
		}
	})

	t.Run("Caso 2: comision_id invalido retorna error esperado", func(t *testing.T) {
		_, err := services.ObtenerHistorialCumplimiento(0)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if !strings.Contains(
			err.Error(),
			"comision_id es obligatorio",
		) {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})
}

// =====================================================
// TEST CREAR REGISTRO CUMPLIMIENTO
// =====================================================

func TestCrearRegistroCumplimiento(t *testing.T) {
	t.Run("Caso 1: exito desactiva historico y crea nuevo registro", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)
		_ = beego.AppConfig.Set("UrlTercerosCrud", "")

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {
				switch {
				case strings.Contains(rawURL, "/estado_comision/7"):
					*(target.(*map[string]interface{})) =
						map[string]interface{}{
							"Data": map[string]interface{}{
								"Id":                float64(7),
								"Nombre":            "Cumplimiento total",
								"CodigoAbreviacion": "CUMP_TOTAL",
							},
						}
					return nil

				case strings.Contains(
					rawURL,
					"/historico_estado_comision/90",
				):
					*(target.(*map[string]interface{})) =
						map[string]interface{}{
							"Data": map[string]interface{}{
								"Id":     float64(90),
								"Activo": true,
							},
						}
					return nil

				case strings.Contains(
					rawURL,
					"historico_estado_comision",
				):
					*(target.(*map[string]interface{})) =
						map[string]interface{}{
							"Data": []interface{}{
								map[string]interface{}{
									"Id": float64(90),
								},
							},
						}
					return nil
				}

				return errors.New("url no esperada: " + rawURL)
			},
		)

		historicoDesactivado := false
		registroCreado := false

		monkey.Patch(
			request.SendJson,
			func(
				rawURL string,
				method string,
				response interface{},
				data interface{},
			) error {
				payload, ok := data.(map[string]interface{})
				if !ok {
					return errors.New("payload inesperado")
				}

				switch {
				case method == "PUT" &&
					strings.Contains(
						rawURL,
						"/historico_estado_comision/90",
					):
					activo, existe := payload["Activo"].(bool)
					if !existe || activo {
						return errors.New(
							"el historico anterior no fue desactivado",
						)
					}

					historicoDesactivado = true
					*(response.(*map[string]interface{})) =
						map[string]interface{}{
							"Data": payload,
						}
					return nil

				case method == "POST" &&
					strings.HasSuffix(
						rawURL,
						"/historico_estado_comision",
					):
					comision, ok :=
						payload["ComisionId"].(map[string]interface{})
					if !ok || comision["Id"] != 10 {
						return errors.New(
							"ComisionId inesperado en el POST",
						)
					}

					estado, ok :=
						payload["EstadoComisionId"].(map[string]interface{})
					if !ok || estado["Id"] != 7 {
						return errors.New(
							"EstadoComisionId inesperado en el POST",
						)
					}

					if payload["Descripcion"] != "Cumplimiento verificado" {
						return errors.New(
							"descripcion inesperada en el POST",
						)
					}

					if payload["RolUsuario"] != "DECANO" {
						return errors.New(
							"rol inesperado en el POST",
						)
					}

					registroCreado = true
					*(response.(*map[string]interface{})) =
						map[string]interface{}{
							"Data": map[string]interface{}{
								"Id": float64(123),
							},
						}
					return nil
				}

				return errors.New(
					"operacion no esperada: " + method + " " + rawURL,
				)
			},
		)

		req := models.CrearRegistroCumplimientoRequest{
			ComisionId:  10,
			EstadoId:    7,
			Descripcion: "Cumplimiento verificado",
			Rol:         "DECANO",
			Nombre:      "Ana Gomez",
		}

		id, err := services.CrearRegistroCumplimiento(req)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo: %v", err)
		}

		if id != 123 {
			t.Errorf("se esperaba id=123 y se obtuvo %d", id)
		}

		if !historicoDesactivado {
			t.Error("se esperaba desactivar el historico anterior")
		}

		if !registroCreado {
			t.Error("se esperaba crear el nuevo registro")
		}
	})

	t.Run("Caso 2: estado que no es de cumplimiento retorna error", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)
		_ = beego.AppConfig.Set("UrlTercerosCrud", "")

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {
				if strings.Contains(rawURL, "/estado_comision/5") {
					*(target.(*map[string]interface{})) =
						map[string]interface{}{
							"Data": map[string]interface{}{
								"Id":                float64(5),
								"Nombre":            "Comision iniciada",
								"CodigoAbreviacion": "COM_INI",
							},
						}
					return nil
				}

				return errors.New("url no esperada: " + rawURL)
			},
		)

		req := models.CrearRegistroCumplimientoRequest{
			ComisionId:  10,
			EstadoId:    5,
			Descripcion: "Intento de registro",
			Rol:         "DECANO",
		}

		id, err := services.CrearRegistroCumplimiento(req)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if id != 0 {
			t.Errorf("se esperaba id=0 y se obtuvo %d", id)
		}

		if !strings.Contains(
			err.Error(),
			"no es un estado de cumplimiento",
		) {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})
}
