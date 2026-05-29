package services_test

// Pruebas de caja negra para services/documentos_pagos.go:
// ObtenerDocumentosPago.

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
// TEST OBTENER DOCUMENTOS PAGO
// =====================================================

func TestObtenerDocumentosPago(t *testing.T) {

	t.Run("Caso 1: comision_id inválido retorna error", func(t *testing.T) {
		_, err := services.ObtenerDocumentosPago(0)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "comision_id es obligatorio") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 2: UrlComisionesCrud no configurado retorna error", func(t *testing.T) {
		_ = beego.AppConfig.Set("UrlComisionesCrud", "")
		_, err := services.ObtenerDocumentosPago(1)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "UrlComisionesCrud") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 3: error al consultar historico activo propaga error", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			return errors.New("servicio no disponible")
		})

		_, err := services.ObtenerDocumentosPago(10)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
	})

	t.Run("Caso 4: error resolviendo tipo_documento_comision PAG_SOPORTE retorna error", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			switch {
			case strings.Contains(rawURL, "historico_estado_comision"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{"Id": float64(99)},
					},
				}
				return nil
			case strings.Contains(rawURL, "tipo_documento_comision") && strings.Contains(rawURL, "CodigoAbreviacion"):
				return errors.New("tipo_documento_comision no disponible")
			}
			return errors.New("url no esperada: " + rawURL)
		})

		_, err := services.ObtenerDocumentosPago(10)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "PAG_SOPORTE") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 5: sin documentos subidos retorna lista vacía", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			switch {
			case strings.Contains(rawURL, "historico_estado_comision"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{"Id": float64(99)},
					},
				}
				return nil
			case strings.Contains(rawURL, "tipo_documento_comision") && strings.Contains(rawURL, "CodigoAbreviacion"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{"Id": float64(7)},
					},
				}
				return nil
			// documento_comision con TipoDocumentoId.Id en la URL (no URL-encoded)
			case strings.Contains(rawURL, "documento_comision"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{},
				}
				return nil
			}
			return errors.New("url no esperada: " + rawURL)
		})

		docs, err := services.ObtenerDocumentosPago(10)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo: %v", err)
		}
		if len(docs) != 0 {
			t.Errorf("se esperaba lista vacía y se obtuvo %d elementos", len(docs))
		}
	})

	t.Run("Caso 6: con documentos retorna items con enlace y metadatos del subidor", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")
		_ = beego.AppConfig.Set("UrlDocumentos", "http://documentos/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			switch {
			case strings.Contains(rawURL, "historico_estado_comision"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{"Id": float64(99)},
					},
				}
				return nil
			case strings.Contains(rawURL, "tipo_documento_comision") && strings.Contains(rawURL, "CodigoAbreviacion"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{"Id": float64(7)},
					},
				}
				return nil
			case strings.Contains(rawURL, "documento_comision"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"Id":          float64(33),
							"DocumentoId": float64(500),
							"EstadoDocumentoComisionId": map[string]interface{}{
								"CodigoAbreviacion": "CARG",
								"Nombre":            "Cargado",
							},
						},
					},
				}
				return nil
			case strings.HasPrefix(rawURL, "http://documentos/"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Enlace":      "http://gestor/pago/500",
					"Nombre":      "soporte_pago.pdf",
					"Descripcion": `{"rol":"SECRETARIA","nombre":"Ana Gomez"}`,
				}
				return nil
			}
			return errors.New("url no esperada: " + rawURL)
		})

		docs, err := services.ObtenerDocumentosPago(10)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo: %v", err)
		}
		if len(docs) != 1 {
			t.Fatalf("se esperaba 1 documento y se obtuvo %d", len(docs))
		}

		doc := docs[0]
		if doc.DocumentoComisionId != 33 {
			t.Errorf("se esperaba DocumentoComisionId=33 y se obtuvo %d", doc.DocumentoComisionId)
		}
		if doc.DocumentoId != 500 {
			t.Errorf("se esperaba DocumentoId=500 y se obtuvo %d", doc.DocumentoId)
		}
		if doc.Estado != "CARG" {
			t.Errorf("se esperaba estado=CARG y se obtuvo %q", doc.Estado)
		}
		if doc.Enlace != "http://gestor/pago/500" {
			t.Errorf("se esperaba enlace 'http://gestor/pago/500' y se obtuvo %q", doc.Enlace)
		}
		if doc.Nombre != "soporte_pago.pdf" {
			t.Errorf("se esperaba nombre 'soporte_pago.pdf' y se obtuvo %q", doc.Nombre)
		}
		if doc.SubidoPorRol != "SECRETARIA" {
			t.Errorf("se esperaba SubidoPorRol=SECRETARIA y se obtuvo %q", doc.SubidoPorRol)
		}
		if doc.SubidoPorNombre != "Ana Gomez" {
			t.Errorf("se esperaba SubidoPorNombre='Ana Gomez' y se obtuvo %q", doc.SubidoPorNombre)
		}
	})
}
