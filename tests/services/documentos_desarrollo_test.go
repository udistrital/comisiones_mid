package services_test

// Pruebas de caja negra para services/documentos_desarrollo.go:
// ObtenerDocumentosDesarrollo, SubirDocumentoDesarrollo, DesactivarDocumentoDesarrollo.

import (
	"errors"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/comisiones_mid/services"
	"github.com/udistrital/utils_oas/request"
)

// =====================================================
// TEST OBTENER DOCUMENTOS DESARROLLO
// =====================================================

func TestObtenerDocumentosDesarrollo(t *testing.T) {

	t.Run("Caso 1: comision_id inválido retorna error", func(t *testing.T) {
		_, err := services.ObtenerDocumentosDesarrollo(0)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "comision_id es obligatorio") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 2: UrlComisionesCrud no configurado retorna error", func(t *testing.T) {
		_ = beego.AppConfig.Set("UrlComisionesCrud", "")
		_, err := services.ObtenerDocumentosDesarrollo(1)
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
			if strings.Contains(rawURL, "historico_estado_comision") {
				return errors.New("servicio no disponible")
			}
			return errors.New("url no esperada: " + rawURL)
		})

		_, err := services.ObtenerDocumentosDesarrollo(10)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
	})

	t.Run("Caso 4: historico activo sin registros retorna error", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			if strings.Contains(rawURL, "historico_estado_comision") {
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{},
				}
				return nil
			}
			return errors.New("url no esperada: " + rawURL)
		})

		_, err := services.ObtenerDocumentosDesarrollo(10)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "no se encontro historico activo") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 5: error al consultar tipo_documento_comision propaga error", func(t *testing.T) {
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
			case strings.Contains(rawURL, "tipo_documento_comision"):
				return errors.New("tipo_documento_comision no disponible")
			}
			return errors.New("url no esperada: " + rawURL)
		})

		_, err := services.ObtenerDocumentosDesarrollo(10)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
	})

	t.Run("Caso 6: sin documentos subidos retorna grupos con items vacíos", func(t *testing.T) {
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
			case strings.Contains(rawURL, "tipo_documento_comision"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"Id":                float64(1),
							"CodigoAbreviacion": "INI_CARTA",
							"Nombre":            "Carta de aceptación",
						},
					},
				}
				return nil
			case strings.Contains(rawURL, "documento_comision"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{},
				}
				return nil
			}
			return errors.New("url no esperada: " + rawURL)
		})

		grupos, err := services.ObtenerDocumentosDesarrollo(10)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo: %v", err)
		}
		if len(grupos) == 0 {
			t.Fatal("se esperaban grupos y se obtuvo lista vacía")
		}

		encontrado := false
		for _, g := range grupos {
			if g.Prefijo == "INI_" {
				encontrado = true
				if len(g.Documentos) != 1 {
					t.Errorf("se esperaba 1 documento en grupo INI_ y se obtuvo %d", len(g.Documentos))
					break
				}
				doc := g.Documentos[0]
				if doc.Codigo != "INI_CARTA" {
					t.Errorf("se esperaba codigo INI_CARTA y se obtuvo %s", doc.Codigo)
				}
				if doc.DocumentoComisionId != 0 {
					t.Errorf("se esperaba DocumentoComisionId=0 (sin doc subido) y se obtuvo %d", doc.DocumentoComisionId)
				}
				if doc.Enlace != "" {
					t.Errorf("se esperaba enlace vacío y se obtuvo %q", doc.Enlace)
				}
			}
		}
		if !encontrado {
			t.Error("no se encontró el grupo con prefijo INI_ en los resultados")
		}
	})

	t.Run("Caso 7: con documentos subidos retorna items con id y enlace", func(t *testing.T) {
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
			// tipo_documento_comision debe ir antes que documento_comision
			case strings.Contains(rawURL, "tipo_documento_comision"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"Id":                float64(1),
							"CodigoAbreviacion": "INI_CARTA",
							"Nombre":            "Carta de aceptación",
						},
					},
				}
				return nil
			case strings.Contains(rawURL, "documento_comision"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"Id":          float64(55),
							"DocumentoId": float64(200),
							"TipoDocumentoId": map[string]interface{}{
								"Id": float64(1),
							},
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
					"Enlace": "http://gestor/doc/200",
				}
				return nil
			}
			return errors.New("url no esperada: " + rawURL)
		})

		grupos, err := services.ObtenerDocumentosDesarrollo(10)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo: %v", err)
		}

		encontrado := false
		for _, g := range grupos {
			for _, doc := range g.Documentos {
				if doc.Codigo == "INI_CARTA" {
					encontrado = true
					if doc.DocumentoComisionId != 55 {
						t.Errorf("se esperaba DocumentoComisionId=55 y se obtuvo %d", doc.DocumentoComisionId)
					}
					if doc.DocumentoId != 200 {
						t.Errorf("se esperaba DocumentoId=200 y se obtuvo %d", doc.DocumentoId)
					}
					if doc.Estado != "CARG" {
						t.Errorf("se esperaba estado=CARG y se obtuvo %q", doc.Estado)
					}
					if doc.EstadoNombre != "Cargado" {
						t.Errorf("se esperaba estado_nombre=Cargado y se obtuvo %q", doc.EstadoNombre)
					}
					if doc.Enlace != "http://gestor/doc/200" {
						t.Errorf("se esperaba enlace 'http://gestor/doc/200' y se obtuvo %q", doc.Enlace)
					}
				}
			}
		}
		if !encontrado {
			t.Error("no se encontró el documento INI_CARTA en ningún grupo")
		}
	})
}

// =====================================================
// TEST SUBIR DOCUMENTO DESARROLLO
// =====================================================

func TestSubirDocumentoDesarrollo(t *testing.T) {

	t.Run("Caso 1: ComisionId inválido retorna error", func(t *testing.T) {
		req := models.SubirDocumentoDesarrolloRequest{
			ComisionId: 0, TipoDocumentoCodigo: "INI_CARTA", Nombre: "doc.pdf", File: "base64",
		}
		_, err := services.SubirDocumentoDesarrollo(req)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "comision_id es obligatorio") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 2: TipoDocumentoCodigo vacío retorna error", func(t *testing.T) {
		req := models.SubirDocumentoDesarrolloRequest{
			ComisionId: 1, TipoDocumentoCodigo: "", Nombre: "doc.pdf", File: "base64",
		}
		_, err := services.SubirDocumentoDesarrollo(req)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "tipo_documento_codigo es obligatorio") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 3: Nombre vacío retorna error", func(t *testing.T) {
		req := models.SubirDocumentoDesarrolloRequest{
			ComisionId: 1, TipoDocumentoCodigo: "INI_CARTA", Nombre: "", File: "base64",
		}
		_, err := services.SubirDocumentoDesarrollo(req)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "nombre es obligatorio") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 4: File vacío retorna error", func(t *testing.T) {
		req := models.SubirDocumentoDesarrolloRequest{
			ComisionId: 1, TipoDocumentoCodigo: "INI_CARTA", Nombre: "doc.pdf", File: "",
		}
		_, err := services.SubirDocumentoDesarrollo(req)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "file es obligatorio") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 5: UrlComisionesCrud no configurado retorna error", func(t *testing.T) {
		_ = beego.AppConfig.Set("UrlComisionesCrud", "")
		req := models.SubirDocumentoDesarrolloRequest{
			ComisionId: 1, TipoDocumentoCodigo: "INI_CARTA", Nombre: "doc.pdf", File: "base64",
		}
		_, err := services.SubirDocumentoDesarrollo(req)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
	})

	t.Run("Caso 6: error al consultar historico activo retorna error", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			return errors.New("historico no disponible")
		})

		req := models.SubirDocumentoDesarrolloRequest{
			ComisionId: 10, TipoDocumentoCodigo: "INI_CARTA", Nombre: "doc.pdf", File: "base64",
		}
		_, err := services.SubirDocumentoDesarrollo(req)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
	})

	t.Run("Caso 7: error en helpers.CrearDocumento retorna error", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			switch {
			case strings.Contains(rawURL, "historico_estado_comision"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{map[string]interface{}{"Id": float64(99)}},
				}
				return nil
			case strings.Contains(rawURL, "CodigoAbreviacion"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{map[string]interface{}{"Id": float64(5)}},
				}
				return nil
			}
			return errors.New("url no esperada: " + rawURL)
		})

		monkey.Patch(helpers.CrearDocumento, func(docs []models.CrearDocumentoGestorDocumental) ([]map[string]interface{}, map[string]interface{}) {
			return nil, map[string]interface{}{"error": "gestor documental no disponible"}
		})

		req := models.SubirDocumentoDesarrolloRequest{
			ComisionId: 10, TipoDocumentoCodigo: "INI_CARTA", Nombre: "doc.pdf", File: "base64",
		}
		_, err := services.SubirDocumentoDesarrollo(req)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "error subiendo documento al gestor documental") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 8: gestor documental retorna lista vacía retorna error", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			switch {
			case strings.Contains(rawURL, "historico_estado_comision"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{map[string]interface{}{"Id": float64(99)}},
				}
				return nil
			case strings.Contains(rawURL, "CodigoAbreviacion"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{map[string]interface{}{"Id": float64(5)}},
				}
				return nil
			}
			return errors.New("url no esperada: " + rawURL)
		})

		monkey.Patch(helpers.CrearDocumento, func(docs []models.CrearDocumentoGestorDocumental) ([]map[string]interface{}, map[string]interface{}) {
			return []map[string]interface{}{}, nil
		})

		req := models.SubirDocumentoDesarrolloRequest{
			ComisionId: 10, TipoDocumentoCodigo: "INI_CARTA", Nombre: "doc.pdf", File: "base64",
		}
		_, err := services.SubirDocumentoDesarrollo(req)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "gestor documental no retornó el documento creado") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 9: éxito completo retorna id del documento_comision creado", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			switch {
			case strings.Contains(rawURL, "historico_estado_comision"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{map[string]interface{}{"Id": float64(99)}},
				}
				return nil
			case strings.Contains(rawURL, "CodigoAbreviacion"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{map[string]interface{}{"Id": float64(5)}},
				}
				return nil
			}
			return errors.New("url no esperada: " + rawURL)
		})

		monkey.Patch(helpers.CrearDocumento, func(docs []models.CrearDocumentoGestorDocumental) ([]map[string]interface{}, map[string]interface{}) {
			return []map[string]interface{}{{"id": 888}}, nil
		})

		monkey.Patch(request.SendJson, func(rawURL, method string, response interface{}, data interface{}) error {
			*(response.(*map[string]interface{})) = map[string]interface{}{
				"Data": map[string]interface{}{"Id": float64(77)},
			}
			return nil
		})

		req := models.SubirDocumentoDesarrolloRequest{
			ComisionId:          10,
			TipoDocumentoCodigo: "INI_CARTA",
			Nombre:              "doc.pdf",
			File:                "base64data",
		}
		id, err := services.SubirDocumentoDesarrollo(req)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo: %v", err)
		}
		if id != 77 {
			t.Errorf("se esperaba id=77 y se obtuvo %d", id)
		}
	})
}

// =====================================================
// TEST DESACTIVAR DOCUMENTO DESARROLLO
// =====================================================

func TestDesactivarDocumentoDesarrollo(t *testing.T) {

	t.Run("Caso 1: id inválido retorna error", func(t *testing.T) {
		err := services.DesactivarDocumentoDesarrollo(0)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "documento_comision_id es obligatorio") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 2: UrlComisionesCrud no configurado retorna error", func(t *testing.T) {
		_ = beego.AppConfig.Set("UrlComisionesCrud", "")
		err := services.DesactivarDocumentoDesarrollo(1)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
	})

	t.Run("Caso 3: error en GET documento_comision retorna error", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			return errors.New("servicio no disponible")
		})

		err := services.DesactivarDocumentoDesarrollo(5)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "error obteniendo documento_comision") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 4: documento_comision no encontrado retorna error", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			// Respuesta sin clave Data → UnwrapDataToMap retorna nil
			*(target.(*map[string]interface{})) = map[string]interface{}{}
			return nil
		})

		err := services.DesactivarDocumentoDesarrollo(5)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "no encontrado") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 5: error en PUT retorna error", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			*(target.(*map[string]interface{})) = map[string]interface{}{
				"Data": map[string]interface{}{"Id": float64(5), "Activo": true},
			}
			return nil
		})

		monkey.Patch(request.SendJson, func(rawURL, method string, response interface{}, data interface{}) error {
			return errors.New("error al actualizar")
		})

		err := services.DesactivarDocumentoDesarrollo(5)
		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "error desactivando documento_comision") {
			t.Errorf("mensaje inesperado: %s", err.Error())
		}
	})

	t.Run("Caso 6: éxito envía Activo=false en el PUT", func(t *testing.T) {
		defer monkey.UnpatchAll()
		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			*(target.(*map[string]interface{})) = map[string]interface{}{
				"Data": map[string]interface{}{"Id": float64(5), "Activo": true},
			}
			return nil
		})

		activoEnviado := true
		monkey.Patch(request.SendJson, func(rawURL, method string, response interface{}, data interface{}) error {
			if payload, ok := data.(map[string]interface{}); ok {
				activoEnviado, _ = payload["Activo"].(bool)
			}
			*(response.(*map[string]interface{})) = map[string]interface{}{
				"Data": map[string]interface{}{"Id": float64(5)},
			}
			return nil
		})

		err := services.DesactivarDocumentoDesarrollo(5)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo: %v", err)
		}
		if activoEnviado {
			t.Error("se esperaba Activo=false en el payload del PUT")
		}
	})
}
