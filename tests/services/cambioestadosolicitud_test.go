package services_test

import (
	"errors"
	"flag"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/comisiones_mid/services"
	"github.com/udistrital/utils_oas/request"
)

var parameters struct {
	UrlComisionesCrud string
	UrlTercerosCrud   string
}

func TestMain(m *testing.M) {
	parameters.UrlComisionesCrud = os.Getenv("UrlComisionesCrud")
	parameters.UrlTercerosCrud = os.Getenv("UrlTercerosCrud")

	_, file, _, _ := runtime.Caller(0)
	apppath, _ := filepath.Abs(filepath.Join(filepath.Dir(file), "../.."))
	beego.TestBeegoInit(apppath)

	if parameters.UrlComisionesCrud != "" {
		_ = beego.AppConfig.Set("UrlComisionesCrud", parameters.UrlComisionesCrud)
	}
	if parameters.UrlTercerosCrud != "" {
		_ = beego.AppConfig.Set("UrlTercerosCrud", parameters.UrlTercerosCrud)
	}

	flag.Parse()
	os.Exit(m.Run())
}

func TestCambiarEstadoSolicitud(t *testing.T) {
	t.Run("Caso 1: cambio de estado exitoso con observacion y documentos", func(t *testing.T) {
		defer monkey.UnpatchAll()

		req := models.CambioEstadoSolicitudRequest{
			NuevoEstado:          "APROB_JEFE",
			RolUsuario:           "JEFE",
			NumeroIdentificacion: "123456789",
			Observacion:          "Cambio aprobado",
			Documentos: []models.DocumentoCambioEstadoRequest{
				{
					IdTipoDocumento: 10,
					TipoDocumento:   "ACTA",
					EstadoDocumento: "CARGADO",
					Nombre:          "acta.pdf",
					Descripcion:     "Documento soporte",
					File:            "archivo-base64",
				},
			},
		}

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			u, err := url.Parse(rawURL)
			if err != nil {
				return err
			}

			switch {
			case strings.Contains(rawURL, "/estado_solicitud"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{"Id": float64(7)},
					},
				}
				return nil

			case strings.Contains(rawURL, "/datos_identificacion"):
				*(target.(*interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"TerceroId": map[string]interface{}{
								"Id": float64(99),
							},
						},
					},
				}
				return nil

			case strings.Contains(rawURL, "/historico_estado_solicitud") && u.Query().Get("query") == "SolicitudId:10,Activo:true":
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"Id": float64(44),
							"EstadoSolicitudId": map[string]interface{}{
								"Id": float64(3),
							},
						},
					},
				}
				return nil

			case strings.HasSuffix(u.Path, "/historico_estado_solicitud/44"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": map[string]interface{}{
						"Id": float64(44),
						"EstadoSolicitudId": map[string]interface{}{
							"Id": float64(3),
						},
						"Activo": true,
					},
				}
				return nil

			case strings.Contains(rawURL, "/tipo_documento_solicitud"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{"Id": float64(21)},
					},
				}
				return nil

			case strings.Contains(rawURL, "/estado_documento"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{"Id": float64(31)},
					},
				}
				return nil
			}

			return errors.New("URL no esperada en GetJson: " + rawURL)
		})

		monkey.Patch(request.SendJson, func(rawURL, method string, target interface{}, body interface{}) error {
			switch {
			case strings.HasSuffix(rawURL, "/historico_estado_solicitud/44") && method == "PUT":
				*(target.(*map[string]interface{})) = map[string]interface{}{"Id": float64(44)}
				return nil

			case strings.HasSuffix(rawURL, "/historico_estado_solicitud") && method == "POST":
				*(target.(*map[string]interface{})) = map[string]interface{}{"Id": float64(55)}
				return nil

			case strings.HasSuffix(rawURL, "/observacion") && method == "POST":
				*(target.(*map[string]interface{})) = map[string]interface{}{"Id": float64(66)}
				return nil

			case strings.HasSuffix(rawURL, "/documento_solicitud") && method == "POST":
				*(target.(*map[string]interface{})) = map[string]interface{}{"Id": float64(88)}
				return nil
			}

			return errors.New("URL no esperada en SendJson: " + rawURL)
		})

		monkey.Patch(helpers.CrearDocumento, func(documentos []models.CrearDocumentoGestorDocumental) ([]map[string]interface{}, map[string]interface{}) {
			return []map[string]interface{}{
				{"id": 501},
			}, nil
		})

		monkey.Patch(services.CrearComision, func(baseCrud string, solicitudId int, terceroId int, rolUsuario string) (int, error) {
			return 0, nil
		})

		resp, err := services.CambiarEstadoSolicitud(10, req)
		if err != nil {
			t.Fatalf("no se esperaba error y se obtuvo: %v", err)
		}

		if resp.SolicitudId != 10 {
			t.Errorf("Se esperaba SolicitudId 10 y se obtuvo %d", resp.SolicitudId)
		}
		if resp.EstadoAnteriorId != 3 {
			t.Errorf("Se esperaba EstadoAnteriorId 3 y se obtuvo %d", resp.EstadoAnteriorId)
		}
		if resp.EstadoDestinoId != 7 {
			t.Errorf("Se esperaba EstadoDestinoId 7 y se obtuvo %d", resp.EstadoDestinoId)
		}
		if resp.TerceroId != 99 {
			t.Errorf("Se esperaba TerceroId 99 y se obtuvo %d", resp.TerceroId)
		}
		if resp.HistoricoAnteriorId != 44 {
			t.Errorf("Se esperaba HistoricoAnteriorId 44 y se obtuvo %d", resp.HistoricoAnteriorId)
		}
		if resp.HistoricoNuevoId != 55 {
			t.Errorf("Se esperaba HistoricoNuevoId 55 y se obtuvo %d", resp.HistoricoNuevoId)
		}
		if resp.ObservacionId != 66 {
			t.Errorf("Se esperaba ObservacionId 66 y se obtuvo %d", resp.ObservacionId)
		}
		if resp.DocumentoId != 501 {
			t.Errorf("Se esperaba DocumentoId 501 y se obtuvo %d", resp.DocumentoId)
		}
		if resp.DocumentoSolicitudId != 88 {
			t.Errorf("Se esperaba DocumentoSolicitudId 88 y se obtuvo %d", resp.DocumentoSolicitudId)
		}
	})

	t.Run("Caso 2: error al consultar el estado destino", func(t *testing.T) {
		defer monkey.UnpatchAll()

		req := models.CambioEstadoSolicitudRequest{
			NuevoEstado:          "APROB_JEFE",
			RolUsuario:           "JEFE",
			NumeroIdentificacion: "123456789",
		}

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			if strings.Contains(rawURL, "/estado_solicitud") {
				return errors.New("servicio no disponible")
			}
			return errors.New("URL no esperada en GetJson: " + rawURL)
		})

		resp, err := services.CambiarEstadoSolicitud(10, req)
		if err == nil {
			t.Fatal("Se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "no se pudo resolver EstadoDestino") {
			t.Errorf("Se esperaba error de EstadoDestino y se obtuvo: %v", err)
		}
		if !reflect.DeepEqual(resp, models.CambioEstadoSolicitudResponse{}) {
			t.Errorf("Se esperaba respuesta vacia y se obtuvo: %+v", resp)
		}
	})

	t.Run("Caso 3: error al crear documentos", func(t *testing.T) {
		defer monkey.UnpatchAll()

		req := models.CambioEstadoSolicitudRequest{
			NuevoEstado:          "APROB_JEFE",
			RolUsuario:           "JEFE",
			NumeroIdentificacion: "123456789",
			Documentos: []models.DocumentoCambioEstadoRequest{
				{
					IdTipoDocumento: 10,
					TipoDocumento:   "ACTA",
					EstadoDocumento: "CARGADO",
					Nombre:          "acta.pdf",
					File:            "archivo-base64",
				},
			},
		}

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			u, err := url.Parse(rawURL)
			if err != nil {
				return err
			}

			switch {
			case strings.Contains(rawURL, "/estado_solicitud"):
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{"Id": float64(7)},
					},
				}
				return nil

			case strings.Contains(rawURL, "/datos_identificacion"):
				*(target.(*interface{})) = map[string]interface{}{
					"Data": []interface{}{
						map[string]interface{}{
							"TerceroId": map[string]interface{}{
								"Id": float64(99),
							},
						},
					},
				}
				return nil

			case strings.Contains(rawURL, "/historico_estado_solicitud") && u.Query().Get("query") == "SolicitudId:10,Activo:true":
				*(target.(*map[string]interface{})) = map[string]interface{}{
					"Data": []interface{}{},
				}
				return nil
			}

			return errors.New("URL no esperada en GetJson: " + rawURL)
		})

		monkey.Patch(request.SendJson, func(rawURL, method string, target interface{}, body interface{}) error {
			if strings.HasSuffix(rawURL, "/historico_estado_solicitud") && method == "POST" {
				*(target.(*map[string]interface{})) = map[string]interface{}{"Id": float64(55)}
				return nil
			}
			if strings.HasSuffix(rawURL, "/historico_estado_solicitud/55") && method == "DELETE" {
				*(target.(*map[string]interface{})) = map[string]interface{}{}
				return nil
			}
			return errors.New("URL no esperada en SendJson: " + rawURL)
		})

		monkey.Patch(helpers.CrearDocumento, func(documentos []models.CrearDocumentoGestorDocumental) ([]map[string]interface{}, map[string]interface{}) {
			return nil, map[string]interface{}{
				"status":  "500",
				"message": "fallo gestor documental",
			}
		})

		resp, err := services.CambiarEstadoSolicitud(10, req)
		if err == nil {
			t.Fatal("Se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "error creando documentos") {
			t.Errorf("Se esperaba error creando documentos y se obtuvo: %v", err)
		}
		if !reflect.DeepEqual(resp, models.CambioEstadoSolicitudResponse{}) {
			t.Errorf("Se esperaba respuesta vacia y se obtuvo: %+v", resp)
		}
	})
}

func TestCrearDocumentosCambioEstado(t *testing.T) {
	t.Run("Caso 1: creacion exitosa de documentos", func(t *testing.T) {
		defer monkey.UnpatchAll()

		documentosReq := []models.DocumentoCambioEstadoRequest{
			{
				IdTipoDocumento: 10,
				TipoDocumento:   "ACTA",
				EstadoDocumento: "CARGADO",
				Nombre:          "acta.pdf",
				Descripcion:     "Documento soporte",
				File:            "archivo-base64",
			},
		}

		monkey.Patch(helpers.CrearDocumento, func(documentos []models.CrearDocumentoGestorDocumental) ([]map[string]interface{}, map[string]interface{}) {
			if len(documentos) != 1 {
				t.Fatalf("Se esperaba 1 documento y se obtuvo %d", len(documentos))
			}
			return []map[string]interface{}{
				{"id": 501},
			}, nil
		})

		monkey.Patch(services.GetIdByCodigoAbreviacion, func(base, recurso, codigo string) (int, error) {
			switch {
			case recurso == "tipo_documento_solicitud" && codigo == "ACTA":
				return 21, nil
			case recurso == "estado_documento" && codigo == "CARGADO":
				return 31, nil
			default:
				return 0, errors.New("recurso no esperado")
			}
		})

		monkey.Patch(services.CrearDocumentoSolicitud, func(baseCrud string, historicoId int, id int, tipoDocumentoId int, estadoDocumentoId int) (int, error) {
			if historicoId != 44 {
				t.Fatalf("Se esperaba historicoId 44 y se obtuvo %d", historicoId)
			}
			if id != 501 {
				t.Fatalf("Se esperaba documentoId 501 y se obtuvo %d", id)
			}
			if tipoDocumentoId != 21 {
				t.Fatalf("Se esperaba tipoDocumentoId 21 y se obtuvo %d", tipoDocumentoId)
			}
			if estadoDocumentoId != 31 {
				t.Fatalf("Se esperaba estadoDocumentoId 31 y se obtuvo %d", estadoDocumentoId)
			}
			return 801, nil
		})

		documentoIds, documentoSolicitudIds, err := services.CrearDocumentosCambioEstado(parameters.UrlComisionesCrud, 44, documentosReq)
		if err != nil {
			t.Fatalf("No se esperaba error y se obtuvo %v", err)
		}
		if len(documentoIds) != 1 || documentoIds[0] != 501 {
			t.Errorf("Se esperaba documentoIds [501] y se obtuvo %v", documentoIds)
		}
		if len(documentoSolicitudIds) != 1 || documentoSolicitudIds[0] != 801 {
			t.Errorf("Se esperaba documentoSolicitudIds [801] y se obtuvo %v", documentoSolicitudIds)
		}
	})

	t.Run("Caso 2: error al crear documentos en gestor documental", func(t *testing.T) {
		defer monkey.UnpatchAll()

		documentosReq := []models.DocumentoCambioEstadoRequest{
			{
				IdTipoDocumento: 10,
				Nombre:          "acta.pdf",
				File:            "archivo-base64",
			},
		}

		monkey.Patch(helpers.CrearDocumento, func(documentos []models.CrearDocumentoGestorDocumental) ([]map[string]interface{}, map[string]interface{}) {
			return nil, map[string]interface{}{
				"status":  "500",
				"message": "fallo gestor documental",
			}
		})

		documentoIds, documentoSolicitudIds, err := services.CrearDocumentosCambioEstado(parameters.UrlComisionesCrud, 44, documentosReq)
		if err == nil {
			t.Fatal("Se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "error creando documentos en gestor documental") {
			t.Errorf("Se esperaba error de gestor documental y se obtuvo %v", err)
		}
		if documentoIds != nil {
			t.Errorf("Se esperaba documentoIds nil y se obtuvo %v", documentoIds)
		}
		if documentoSolicitudIds != nil {
			t.Errorf("Se esperaba documentoSolicitudIds nil y se obtuvo %v", documentoSolicitudIds)
		}
	})
}

func TestGetIdByCodigoAbreviacion(t *testing.T) {
	t.Run("Caso 1: consulta exitosa del id", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			u, err := url.Parse(rawURL)
			if err != nil {
				return err
			}
			if u.Path != "/estado_solicitud" {
				t.Fatalf("Se esperaba path /estado_solicitud y se obtuvo %s", u.Path)
			}
			if u.Query().Get("query") != "CodigoAbreviacion:APROB_JEFE,Activo:true" {
				t.Fatalf("Query no esperada: %s", u.Query().Get("query"))
			}

			*(target.(*map[string]interface{})) = map[string]interface{}{
				"Data": []interface{}{
					map[string]interface{}{"Id": float64(17)},
				},
			}
			return nil
		})

		id, err := services.GetIdByCodigoAbreviacion(parameters.UrlComisionesCrud, "estado_solicitud", "APROB_JEFE")
		if err != nil {
			t.Fatalf("No se esperaba error y se obtuvo %v", err)
		}
		if id != 17 {
			t.Errorf("Se esperaba id 17 y se obtuvo %d", id)
		}
	})

	t.Run("Caso 2: error al consultar el id", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			return errors.New("servicio no disponible")
		})

		id, err := services.GetIdByCodigoAbreviacion(parameters.UrlComisionesCrud, "estado_solicitud", "APROB_JEFE")
		if err == nil {
			t.Fatal("Se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "servicio no disponible") {
			t.Errorf("Se esperaba error del servicio y se obtuvo %v", err)
		}
		if id != 0 {
			t.Errorf("Se esperaba id 0 y se obtuvo %d", id)
		}
	})
}

func TestGetTerceroIdByNumeroIdentificacion(t *testing.T) {
	t.Run("Caso 1: consulta exitosa del tercero", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			u, err := url.Parse(rawURL)
			if err != nil {
				return err
			}
			if u.Path != "/datos_identificacion" {
				t.Fatalf("Se esperaba path /datos_identificacion y se obtuvo %s", u.Path)
			}
			if u.Query().Get("query") != "numero:123456789" {
				t.Fatalf("Query no esperada: %s", u.Query().Get("query"))
			}

			*(target.(*interface{})) = map[string]interface{}{
				"Data": []interface{}{
					map[string]interface{}{
						"TerceroId": map[string]interface{}{
							"Id": float64(99),
						},
					},
				},
			}
			return nil
		})

		id, err := services.GetTerceroIdByNumeroIdentificacion(parameters.UrlTercerosCrud, "123456789")
		if err != nil {
			t.Fatalf("No se esperaba error y se obtuvo %v", err)
		}
		if id != 99 {
			t.Errorf("Se esperaba id 99 y se obtuvo %d", id)
		}
	})

	t.Run("Caso 2: error al consultar tercero", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			return errors.New("servicio terceros no disponible")
		})

		id, err := services.GetTerceroIdByNumeroIdentificacion(parameters.UrlTercerosCrud, "123456789")
		if err == nil {
			t.Fatal("Se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "servicio terceros no disponible") {
			t.Errorf("Se esperaba error del servicio y se obtuvo %v", err)
		}
		if id != 0 {
			t.Errorf("Se esperaba id 0 y se obtuvo %d", id)
		}
	})
}

func TestCrearDocumentoSolicitud(t *testing.T) {
	t.Run("Caso 1: creacion exitosa de documento solicitud", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.SendJson, func(rawURL, method string, target interface{}, body interface{}) error {
			if !strings.HasSuffix(rawURL, "/documento_solicitud") {
				t.Fatalf("URL no esperada: %s", rawURL)
			}
			if method != "POST" {
				t.Fatalf("Metodo no esperado: %s", method)
			}

			payload := body.(map[string]interface{})
			if payload["DocumentoId"] != 501 {
				t.Fatalf("Se esperaba DocumentoId 501 y se obtuvo %v", payload["DocumentoId"])
			}

			*(target.(*map[string]interface{})) = map[string]interface{}{
				"Id": float64(77),
			}
			return nil
		})

		id, err := services.CrearDocumentoSolicitud(parameters.UrlComisionesCrud, 44, 501, 21, 31)
		if err != nil {
			t.Fatalf("No se esperaba error y se obtuvo %v", err)
		}
		if id != 77 {
			t.Errorf("Se esperaba id 77 y se obtuvo %d", id)
		}
	})

	t.Run("Caso 2: error al crear documento solicitud", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.SendJson, func(rawURL, method string, target interface{}, body interface{}) error {
			return errors.New("error en documento_solicitud")
		})

		id, err := services.CrearDocumentoSolicitud(parameters.UrlComisionesCrud, 44, 501, 21, 31)
		if err == nil {
			t.Fatal("Se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "error creando documento_solicitud") {
			t.Errorf("Se esperaba error de documento_solicitud y se obtuvo %v", err)
		}
		if id != 0 {
			t.Errorf("Se esperaba id 0 y se obtuvo %d", id)
		}
	})
}

func TestGetHistoricoActivoActual(t *testing.T) {
	t.Run("Caso 1: consulta exitosa del historico activo", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			u, err := url.Parse(rawURL)
			if err != nil {
				return err
			}
			if u.Path != "/historico_estado_solicitud" {
				t.Fatalf("Path no esperado: %s", u.Path)
			}
			if u.Query().Get("query") != "SolicitudId:10,Activo:true" {
				t.Fatalf("Query no esperada: %s", u.Query().Get("query"))
			}

			*(target.(*map[string]interface{})) = map[string]interface{}{
				"Data": []interface{}{
					map[string]interface{}{
						"Id":     float64(44),
						"Activo": true,
					},
				},
			}
			return nil
		})

		row, err := services.GetHistoricoActivoActual(parameters.UrlComisionesCrud, 10)
		if err != nil {
			t.Fatalf("No se esperaba error y se obtuvo %v", err)
		}
		if row == nil {
			t.Fatal("Se esperaba historico y se obtuvo nil")
		}
		if row["Id"] != float64(44) {
			t.Errorf("Se esperaba Id 44 y se obtuvo %v", row["Id"])
		}
	})

	t.Run("Caso 2: error al consultar historico activo", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			return errors.New("error consultando historico")
		})

		row, err := services.GetHistoricoActivoActual(parameters.UrlComisionesCrud, 10)
		if err == nil {
			t.Fatal("Se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "error consultando histórico actual") && !strings.Contains(err.Error(), "error consultando hist") {
			t.Errorf("Se esperaba error de historico y se obtuvo %v", err)
		}
		if row != nil {
			t.Errorf("Se esperaba row nil y se obtuvo %v", row)
		}
	})
}

func TestDesActivarHistorico(t *testing.T) {
	t.Run("Caso 1: desactivacion exitosa del historico", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			*(target.(*map[string]interface{})) = map[string]interface{}{
				"Data": map[string]interface{}{
					"Id":     float64(44),
					"Activo": true,
				},
			}
			return nil
		})

		monkey.Patch(request.SendJson, func(rawURL, method string, target interface{}, body interface{}) error {
			if method != "PUT" {
				t.Fatalf("Se esperaba PUT y se obtuvo %s", method)
			}
			payload := body.(map[string]interface{})
			if payload["Activo"] != false {
				t.Fatalf("Se esperaba Activo false y se obtuvo %v", payload["Activo"])
			}
			*(target.(*map[string]interface{})) = map[string]interface{}{}
			return nil
		})

		err := services.DesActivarHistorico(parameters.UrlComisionesCrud, 44)
		if err != nil {
			t.Fatalf("No se esperaba error y se obtuvo %v", err)
		}
	})

	t.Run("Caso 2: error al actualizar el historico", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			*(target.(*map[string]interface{})) = map[string]interface{}{
				"Data": map[string]interface{}{
					"Id":     float64(44),
					"Activo": true,
				},
			}
			return nil
		})

		monkey.Patch(request.SendJson, func(rawURL, method string, target interface{}, body interface{}) error {
			return errors.New("error PUT historico")
		})

		err := services.DesActivarHistorico(parameters.UrlComisionesCrud, 44)
		if err == nil {
			t.Fatal("Se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "error PUT histórico") && !strings.Contains(err.Error(), "error PUT hist") {
			t.Errorf("Se esperaba error de PUT y se obtuvo %v", err)
		}
	})
}

func TestRevertirCambioEstado(t *testing.T) {
	t.Run("Caso 1: rollback exitoso", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(services.EliminarObservacion, func(baseCrud string, observacionId int) error {
			return nil
		})
		monkey.Patch(services.EliminarHistorico, func(baseCrud string, historicoId int) error {
			return nil
		})
		monkey.Patch(services.ActivarHistorico, func(base string, historicoId int) error {
			return nil
		})

		err := services.RevertirCambioEstado(parameters.UrlComisionesCrud, 44, 55, 66)
		if err != nil {
			t.Fatalf("No se esperaba error y se obtuvo %v", err)
		}
	})

	t.Run("Caso 2: error eliminando historico nuevo", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(services.EliminarObservacion, func(baseCrud string, observacionId int) error {
			return nil
		})
		monkey.Patch(services.EliminarHistorico, func(baseCrud string, historicoId int) error {
			return errors.New("fallo eliminando historico")
		})
		monkey.Patch(services.ActivarHistorico, func(base string, historicoId int) error {
			return nil
		})

		err := services.RevertirCambioEstado(parameters.UrlComisionesCrud, 44, 55, 66)
		if err == nil {
			t.Fatal("Se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "no se pudo eliminar el histórico nuevo") && !strings.Contains(err.Error(), "no se pudo eliminar el hist") {
			t.Errorf("Se esperaba error de rollback y se obtuvo %v", err)
		}
	})
}

func TestEliminarObservacion(t *testing.T) {
	t.Run("Caso 1: eliminacion exitosa de observacion", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.SendJson, func(rawURL, method string, target interface{}, body interface{}) error {
			if !strings.HasSuffix(rawURL, "/observacion/66") {
				t.Fatalf("URL no esperada: %s", rawURL)
			}
			if method != "DELETE" {
				t.Fatalf("Metodo no esperado: %s", method)
			}
			*(target.(*map[string]interface{})) = map[string]interface{}{}
			return nil
		})

		err := services.EliminarObservacion(parameters.UrlComisionesCrud, 66)
		if err != nil {
			t.Fatalf("No se esperaba error y se obtuvo %v", err)
		}
	})

	t.Run("Caso 2: error eliminando observacion", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.SendJson, func(rawURL, method string, target interface{}, body interface{}) error {
			return errors.New("fallo eliminando observacion")
		})

		err := services.EliminarObservacion(parameters.UrlComisionesCrud, 66)
		if err == nil {
			t.Fatal("Se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "error eliminando observación") && !strings.Contains(err.Error(), "error eliminando observ") {
			t.Errorf("Se esperaba error eliminando observacion y se obtuvo %v", err)
		}
	})
}

func TestEliminarHistorico(t *testing.T) {
	t.Run("Caso 1: eliminacion exitosa de historico", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.SendJson, func(rawURL, method string, target interface{}, body interface{}) error {
			if !strings.HasSuffix(rawURL, "/historico_estado_solicitud/55") {
				t.Fatalf("URL no esperada: %s", rawURL)
			}
			if method != "DELETE" {
				t.Fatalf("Metodo no esperado: %s", method)
			}
			*(target.(*map[string]interface{})) = map[string]interface{}{}
			return nil
		})

		err := services.EliminarHistorico(parameters.UrlComisionesCrud, 55)
		if err != nil {
			t.Fatalf("No se esperaba error y se obtuvo %v", err)
		}
	})

	t.Run("Caso 2: error eliminando historico", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.SendJson, func(rawURL, method string, target interface{}, body interface{}) error {
			return errors.New("fallo eliminando historico")
		})

		err := services.EliminarHistorico(parameters.UrlComisionesCrud, 55)
		if err == nil {
			t.Fatal("Se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "error eliminando histórico") && !strings.Contains(err.Error(), "error eliminando hist") {
			t.Errorf("Se esperaba error eliminando historico y se obtuvo %v", err)
		}
	})
}

func TestActivarHistorico(t *testing.T) {
	t.Run("Caso 1: activacion exitosa del historico", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			*(target.(*map[string]interface{})) = map[string]interface{}{
				"Data": map[string]interface{}{
					"Id":     float64(44),
					"Activo": false,
				},
			}
			return nil
		})

		monkey.Patch(request.SendJson, func(rawURL, method string, target interface{}, body interface{}) error {
			if method != "PUT" {
				t.Fatalf("Se esperaba PUT y se obtuvo %s", method)
			}
			payload := body.(map[string]interface{})
			if payload["Activo"] != true {
				t.Fatalf("Se esperaba Activo true y se obtuvo %v", payload["Activo"])
			}
			*(target.(*map[string]interface{})) = map[string]interface{}{}
			return nil
		})

		err := services.ActivarHistorico(parameters.UrlComisionesCrud, 44)
		if err != nil {
			t.Fatalf("No se esperaba error y se obtuvo %v", err)
		}
	})

	t.Run("Caso 2: error activando historico", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.Patch(request.GetJson, func(rawURL string, target interface{}) error {
			return errors.New("error GET historico")
		})

		err := services.ActivarHistorico(parameters.UrlComisionesCrud, 44)
		if err == nil {
			t.Fatal("Se esperaba error y no se obtuvo")
		}
		if !strings.Contains(err.Error(), "error GET histórico") && !strings.Contains(err.Error(), "error GET hist") {
			t.Errorf("Se esperaba error GET historico y se obtuvo %v", err)
		}
	})
}
