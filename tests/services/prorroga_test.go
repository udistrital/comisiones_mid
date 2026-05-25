package services_test

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
// TEST VALIDAR PUEDE CREAR SOLICITUD PRÓRROGA
// =====================================================
func TestValidarPuedeCrearSolicitudProrroga(t *testing.T) {

	t.Run("Caso 1: no existen solicitudes y puede crear", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				if strings.Contains(rawURL, "solicitud?query=") {

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: true,
							Status:  "200",
							Data:    []models.Solicitud{},
						}

					return nil
				}

				return errors.New("url no esperada")
			},
		)

		puedeCrear, estado, err :=
			services.ValidarPuedeCrearSolicitudProrroga(10)

		if err != nil {
			t.Fatalf(
				"no se esperaba error y se obtuvo %v",
				err,
			)
		}

		if !puedeCrear {
			t.Error("se esperaba true y se obtuvo false")
		}

		if estado != "" {
			t.Errorf(
				"se esperaba estado vacío y se obtuvo %s",
				estado,
			)
		}
	})

	t.Run("Caso 2: existe solicitud aprobada y no puede crear", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				switch {

				case strings.Contains(
					rawURL,
					"historico_estado_solicitud",
				):

					*(target.(*models.ResponseListaHistoricoEstadoSolicitud)) =
						models.ResponseListaHistoricoEstadoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.HistoricoEstadoSolicitud{
								{
									Id: 1,

									EstadoSolicitudId: &models.EstadoSolicitud{
										Nombre: "Aprobada",
									},
								},
							},
						}

					return nil

				case strings.Contains(rawURL, "solicitud?query="):

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.Solicitud{
								{
									Id: 100,
								},
							},
						}

					return nil
				}

				return errors.New("url no esperada")
			},
		)

		puedeCrear, estado, err :=
			services.ValidarPuedeCrearSolicitudProrroga(10)

		if err != nil {
			t.Fatalf(
				"no se esperaba error y se obtuvo %v",
				err,
			)
		}

		if puedeCrear {
			t.Error("se esperaba false y se obtuvo true")
		}

		if estado != "Aprobada" {
			t.Errorf(
				"se esperaba estado Aprobada y se obtuvo %s",
				estado,
			)
		}
	})

	t.Run("Caso 3: todas las solicitudes están rechazadas", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				switch {

				case strings.Contains(
					rawURL,
					"historico_estado_solicitud",
				):

					*(target.(*models.ResponseListaHistoricoEstadoSolicitud)) =
						models.ResponseListaHistoricoEstadoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.HistoricoEstadoSolicitud{
								{
									Id: 1,

									EstadoSolicitudId: &models.EstadoSolicitud{
										Nombre: "No aprobada",
									},
								},
							},
						}

					return nil

				case strings.Contains(rawURL, "solicitud?query="):

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.Solicitud{
								{
									Id: 100,
								},
							},
						}

					return nil
				}

				return errors.New("url no esperada")
			},
		)

		puedeCrear, estado, err :=
			services.ValidarPuedeCrearSolicitudProrroga(10)

		if err != nil {
			t.Fatalf(
				"no se esperaba error y se obtuvo %v",
				err,
			)
		}

		if !puedeCrear {
			t.Error("se esperaba true y se obtuvo false")
		}

		if estado != "" {
			t.Errorf(
				"se esperaba estado vacío y se obtuvo %s",
				estado,
			)
		}
	})

	t.Run("Caso 4: error consultando solicitudes", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				return errors.New("servicio no disponible")
			},
		)

		puedeCrear, estado, err :=
			services.ValidarPuedeCrearSolicitudProrroga(10)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if err.Error() != "servicio no disponible" {
			t.Errorf(
				"se esperaba error 'servicio no disponible' y se obtuvo %s",
				err.Error(),
			)
		}

		if puedeCrear {
			t.Error("se esperaba false y se obtuvo true")
		}

		if estado != "" {
			t.Errorf(
				"se esperaba estado vacío y se obtuvo %s",
				estado,
			)
		}
	})

	t.Run("Caso 5: error en respuesta de solicitudes", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				if strings.Contains(rawURL, "solicitud?query=") {

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: false,
							Status:  "500",
						}

					return nil
				}

				return errors.New("url no esperada")
			},
		)

		_, _, err :=
			services.ValidarPuedeCrearSolicitudProrroga(10)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if !strings.Contains(
			err.Error(),
			"error consultando solicitudes de prórroga",
		) {
			t.Errorf(
				"mensaje inesperado: %s",
				err.Error(),
			)
		}
	})

	t.Run("Caso 6: respuesta inesperada consultando histórico", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				switch {

				case strings.Contains(
					rawURL,
					"historico_estado_solicitud",
				):

					*(target.(*models.ResponseListaHistoricoEstadoSolicitud)) =
						models.ResponseListaHistoricoEstadoSolicitud{
							Success: true,
							Status:  "500",
						}

					return nil

				case strings.Contains(rawURL, "solicitud?query="):

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.Solicitud{
								{
									Id: 100,
								},
							},
						}

					return nil
				}

				return errors.New("url no esperada")
			},
		)

		_, _, err :=
			services.ValidarPuedeCrearSolicitudProrroga(10)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if !strings.Contains(
			err.Error(),
			"respuesta inesperada consultando histórico",
		) {
			t.Errorf(
				"mensaje inesperado: %s",
				err.Error(),
			)
		}
	})

	t.Run("Caso 7: histórico vacío permite crear", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				switch {

				case strings.Contains(
					rawURL,
					"historico_estado_solicitud",
				):

					*(target.(*models.ResponseListaHistoricoEstadoSolicitud)) =
						models.ResponseListaHistoricoEstadoSolicitud{
							Success: true,
							Status:  "200",
							Data:    []models.HistoricoEstadoSolicitud{},
						}

					return nil

				case strings.Contains(rawURL, "solicitud?query="):

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.Solicitud{
								{
									Id: 100,
								},
							},
						}

					return nil
				}

				return errors.New("url no esperada")
			},
		)

		puedeCrear, estado, err :=
			services.ValidarPuedeCrearSolicitudProrroga(10)

		if err != nil {
			t.Fatalf(
				"no se esperaba error y se obtuvo %v",
				err,
			)
		}

		if !puedeCrear {
			t.Error("se esperaba true y se obtuvo false")
		}

		if estado != "" {
			t.Errorf(
				"se esperaba estado vacío y se obtuvo %s",
				estado,
			)
		}
	})
}

// =====================================================
// TEST CONSULTAR HISTÓRICO SOLICITUDES PRÓRROGA
// =====================================================

func TestConsultarHistoricoSolicitudesProrroga(t *testing.T) {

	t.Run("Caso 1: retorna histórico correctamente", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				switch {

				case strings.Contains(
					rawURL,
					"solicitud?query=TipoSolicitudId__CodigoAbreviacion:SOL_PRORROGA",
				):

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.Solicitud{
								{
									Id: 100,
								},
							},
						}

					return nil

				case strings.Contains(
					rawURL,
					"historico_estado_solicitud",
				):

					*(target.(*models.ResponseListaHistoricoEstadoSolicitud)) =
						models.ResponseListaHistoricoEstadoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.HistoricoEstadoSolicitud{
								{
									Id: 500,

									SolicitudId: &models.Solicitud{
										Id:            100,
										FechaCreacion: "2026-01-01",
									},

									EstadoSolicitudId: &models.EstadoSolicitud{
										Nombre: "No aprobada",
									},
								},
							},
						}

					return nil
				}

				return errors.New(
					"url no esperada: " + rawURL,
				)
			},
		)

		resp, err :=
			services.ConsultarHistoricoSolicitudesProrroga(10)

		if err != nil {
			t.Fatalf(
				"no se esperaba error y se obtuvo %v",
				err,
			)
		}

		if len(resp) != 1 {
			t.Fatalf(
				"se esperaba 1 elemento y se obtuvo %d",
				len(resp),
			)
		}

		if resp[0].SolicitudId != 100 {
			t.Errorf(
				"se esperaba SolicitudId 100 y se obtuvo %d",
				resp[0].SolicitudId,
			)
		}

		if resp[0].HistoricoId != 500 {
			t.Errorf(
				"se esperaba HistoricoId 500 y se obtuvo %d",
				resp[0].HistoricoId,
			)
		}

		if resp[0].Estado != "No aprobada" {
			t.Errorf(
				"se esperaba estado No aprobada y se obtuvo %s",
				resp[0].Estado,
			)
		}
	})

	t.Run("Caso 2: error consultando solicitudes", func(t *testing.T) {

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

		resp, err :=
			services.ConsultarHistoricoSolicitudesProrroga(10)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if err.Error() != "servicio no disponible" {
			t.Errorf(
				"se esperaba 'servicio no disponible' y se obtuvo %s",
				err.Error(),
			)
		}

		if resp != nil {
			t.Errorf(
				"se esperaba nil y se obtuvo %v",
				resp,
			)
		}
	})

	t.Run("Caso 3: respuesta unsuccess consultando solicitudes", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				*(target.(*models.ResponseListaSolicitud)) =
					models.ResponseListaSolicitud{
						Success: false,
						Status:  "500",
					}

				return nil
			},
		)

		resp, err :=
			services.ConsultarHistoricoSolicitudesProrroga(10)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if !strings.Contains(
			err.Error(),
			"error consultando solicitudes de prórroga",
		) {
			t.Errorf(
				"mensaje inesperado: %s",
				err.Error(),
			)
		}

		if resp != nil {
			t.Errorf(
				"se esperaba nil y se obtuvo %v",
				resp,
			)
		}
	})

	t.Run("Caso 4: status inesperado consultando solicitudes", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				*(target.(*models.ResponseListaSolicitud)) =
					models.ResponseListaSolicitud{
						Success: true,
						Status:  "404",
					}

				return nil
			},
		)

		resp, err :=
			services.ConsultarHistoricoSolicitudesProrroga(10)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if !strings.Contains(
			err.Error(),
			"respuesta inesperada consultando solicitudes de prórroga",
		) {
			t.Errorf(
				"mensaje inesperado: %s",
				err.Error(),
			)
		}

		if resp != nil {
			t.Errorf(
				"se esperaba nil y se obtuvo %v",
				resp,
			)
		}
	})

	t.Run("Caso 5: histórico vacío continúa correctamente", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				switch {

				case strings.Contains(
					rawURL,
					"solicitud?query=TipoSolicitudId__CodigoAbreviacion:SOL_PRORROGA",
				):

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.Solicitud{
								{
									Id: 100,
								},
							},
						}

					return nil

				case strings.Contains(
					rawURL,
					"historico_estado_solicitud",
				):

					*(target.(*models.ResponseListaHistoricoEstadoSolicitud)) =
						models.ResponseListaHistoricoEstadoSolicitud{
							Success: true,
							Status:  "200",
							Data:    []models.HistoricoEstadoSolicitud{},
						}

					return nil
				}

				return errors.New(
					"url no esperada: " + rawURL,
				)
			},
		)

		resp, err :=
			services.ConsultarHistoricoSolicitudesProrroga(10)

		if err != nil {
			t.Fatalf(
				"no se esperaba error y se obtuvo %v",
				err,
			)
		}

		if len(resp) != 0 {
			t.Errorf(
				"se esperaba lista vacía y se obtuvo %d elementos",
				len(resp),
			)
		}
	})
}

// =====================================================
// TEST CREAR SOLICITUDES PRÓRROGA
// =====================================================

func TestCrearSolicitudProrroga(t *testing.T) {

	t.Run("Caso 1: crea correctamente la solicitud de prórroga", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(
			services.ValidarPuedeCrearSolicitudProrroga,
			func(comisionId int) (bool, string, error) {
				return true, "", nil
			},
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				switch {

				case strings.Contains(
					rawURL,
					"TipoSolicitudId__CodigoAbreviacion:SOL_INI",
				):

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.Solicitud{
								{
									Id:        1,
									TerceroId: 999,
									ComisionId: &models.Comision{
										Id: 10,
									},
								},
							},
						}

					return nil

				case strings.Contains(rawURL, "tipo_solicitud"):

					*(target.(*models.ResponseListaTipoSolicitud)) =
						models.ResponseListaTipoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.TipoSolicitud{
								{
									Id:                2,
									CodigoAbreviacion: "SOL_PRORROGA",
								},
							},
						}

					return nil

				case strings.Contains(rawURL, "estado_documento"):

					*(target.(*models.ResponseListaEstadoDocumento)) =
						models.ResponseListaEstadoDocumento{
							Success: true,
							Status:  "200",
							Data: []models.EstadoDocumento{
								{
									Id: 3,
								},
							},
						}

					return nil

				case strings.Contains(rawURL, "estado_solicitud"):

					*(target.(*models.ResponseListaEstadoSolicitud)) =
						models.ResponseListaEstadoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.EstadoSolicitud{
								{
									Id:     4,
									Nombre: "En revisión",
								},
							},
						}

					return nil

				case strings.Contains(rawURL, "tipo_documento_solicitud"):

					*(target.(*models.ResponseListaTipoDocumentoSolicitud)) =
						models.ResponseListaTipoDocumentoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.TipoDocumentoSolicitud{
								{
									Id:                11,
									CodigoAbreviacion: "SOL_PRO_CARTA",
								},
							},
						}

					return nil
				}

				return errors.New("url no esperada: " + rawURL)
			},
		)

		monkey.Patch(
			request.SendJson,
			func(
				rawURL string,
				method string,
				response interface{},
				data interface{},
			) error {

				switch {

				case strings.Contains(
					rawURL,
					"historico_estado_solicitud",
				):

					*(response.(*models.ResponseCreateHistoricoEstadoSolicitud)) =
						models.ResponseCreateHistoricoEstadoSolicitud{
							Success: true,
							Status:  "201",
							Data: models.HistoricoEstadoSolicitud{
								Id: 200,
							},
						}

					return nil

				case strings.Contains(
					rawURL,
					"documento_solicitud",
				):

					*(response.(*map[string]interface{})) =
						map[string]interface{}{
							"Success": true,
						}

					return nil

				case strings.Contains(rawURL, "solicitud") &&
					!strings.Contains(rawURL, "historico") &&
					!strings.Contains(rawURL, "documento"):

					*(response.(*models.ResponseCreateSolicitud)) =
						models.ResponseCreateSolicitud{
							Success: true,
							Status:  "201",
							Data: models.Solicitud{
								Id: 100,
							},
						}

					return nil
				}

				return errors.New("url no esperada")
			},
		)

		monkey.Patch(
			helpers.CrearDocumento,
			func(
				documentos []models.CrearDocumentoGestorDocumental,
			) ([]map[string]interface{}, map[string]interface{}) {

				return []map[string]interface{}{
					{
						"id": 900,
					},
				}, nil
			},
		)

		entrada := models.CrearSolicitudProrrogaEntrada{
			ComisionId:           10,
			Observacion:          "Solicitud de prórroga",
			CodigoAbreviacionRol: "DOCENTE",
			DocumentosSolicitudProrroga: []models.DocumentoProrroga{
				{
					CodigoAbreviacionDoc: "SOL_PRO_CARTA",
					DocumentoSolicitud: models.CrearDocumentoGestorDocumental{
						Nombre: "carta.pdf",
					},
				},
			},
		}

		resultado, err := services.CrearSolicitudProrroga(entrada)

		if err != nil {
			t.Fatalf(
				"no se esperaba error y se obtuvo: %v",
				err,
			)
		}

		if resultado.ComisionId != 10 {
			t.Errorf(
				"se esperaba ComisionId 10 y se obtuvo %d",
				resultado.ComisionId,
			)
		}

		if resultado.SolicitudProrrogaId != 100 {
			t.Errorf(
				"se esperaba SolicitudProrrogaId 100 y se obtuvo %d",
				resultado.SolicitudProrrogaId,
			)
		}
	})

	t.Run("Caso 2: retorna error cuando no puede crear solicitud", func(t *testing.T) {

		defer monkey.UnpatchAll()

		monkey.Patch(
			services.ValidarPuedeCrearSolicitudProrroga,
			func(comisionId int) (bool, string, error) {
				return false, "En revisión", nil
			},
		)

		entrada := models.CrearSolicitudProrrogaEntrada{
			ComisionId: 10,
		}

		resultado, err := services.CrearSolicitudProrroga(entrada)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if !strings.Contains(
			err.Error(),
			"El maestro ya tiene una solicitud de prórroga",
		) {
			t.Errorf(
				"mensaje inesperado: %v",
				err.Error(),
			)
		}

		if resultado.SolicitudProrrogaId != 0 {
			t.Errorf(
				"se esperaba SolicitudProrrogaId 0 y se obtuvo %d",
				resultado.SolicitudProrrogaId,
			)
		}
	})

	t.Run("Caso 3: error consultando solicitud base", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(
			services.ValidarPuedeCrearSolicitudProrroga,
			func(comisionId int) (bool, string, error) {
				return true, "", nil
			},
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {
				return errors.New("servicio no disponible")
			},
		)

		entrada := models.CrearSolicitudProrrogaEntrada{
			ComisionId: 10,
		}

		resultado, err := services.CrearSolicitudProrroga(entrada)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if err.Error() != "servicio no disponible" {
			t.Errorf(
				"se esperaba 'servicio no disponible' y se obtuvo %v",
				err.Error(),
			)
		}

		if resultado.SolicitudProrrogaId != 0 {
			t.Errorf(
				"se esperaba SolicitudProrrogaId 0 y se obtuvo %d",
				resultado.SolicitudProrrogaId,
			)
		}
	})

	t.Run("Caso 4: documentos enviados no coinciden con los requeridos", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(
			services.ValidarPuedeCrearSolicitudProrroga,
			func(comisionId int) (bool, string, error) {
				return true, "", nil
			},
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				switch {

				case strings.Contains(
					rawURL,
					"TipoSolicitudId__CodigoAbreviacion:SOL_INI",
				):

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.Solicitud{
								{
									Id:        1,
									TerceroId: 999,
								},
							},
						}

					return nil

				case strings.Contains(rawURL, "tipo_solicitud"):

					*(target.(*models.ResponseListaTipoSolicitud)) =
						models.ResponseListaTipoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.TipoSolicitud{
								{
									Id: 2,
								},
							},
						}

					return nil

				case strings.Contains(rawURL, "estado_documento"):

					*(target.(*models.ResponseListaEstadoDocumento)) =
						models.ResponseListaEstadoDocumento{
							Success: true,
							Status:  "200",
							Data: []models.EstadoDocumento{
								{
									Id: 3,
								},
							},
						}

					return nil

				case strings.Contains(rawURL, "estado_solicitud"):

					*(target.(*models.ResponseListaEstadoSolicitud)) =
						models.ResponseListaEstadoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.EstadoSolicitud{
								{
									Id: 4,
								},
							},
						}

					return nil

				case strings.Contains(rawURL, "tipo_documento_solicitud"):

					*(target.(*models.ResponseListaTipoDocumentoSolicitud)) =
						models.ResponseListaTipoDocumentoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.TipoDocumentoSolicitud{
								{
									Id:                11,
									CodigoAbreviacion: "SOL_PRO_CARTA",
								},
							},
						}

					return nil
				}

				return nil
			},
		)

		entrada := models.CrearSolicitudProrrogaEntrada{
			ComisionId: 10,
			DocumentosSolicitudProrroga: []models.DocumentoProrroga{
				{
					CodigoAbreviacionDoc: "DOCUMENTO_INVALIDO",
				},
			},
		}

		resultado, err := services.CrearSolicitudProrroga(entrada)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if err.Error() !=
			"los documentos enviados no coinciden con los requeridos" {

			t.Errorf(
				"mensaje inesperado: %v",
				err.Error(),
			)
		}

		if resultado.SolicitudProrrogaId != 0 {
			t.Errorf(
				"se esperaba SolicitudProrrogaId 0 y se obtuvo %d",
				resultado.SolicitudProrrogaId,
			)
		}
	})

	t.Run("Caso 5: error creando documentos", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set("UrlComisionesCrud", "http://comisiones/")

		monkey.Patch(
			services.ValidarPuedeCrearSolicitudProrroga,
			func(comisionId int) (bool, string, error) {
				return true, "", nil
			},
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				switch {

				case strings.Contains(
					rawURL,
					"TipoSolicitudId__CodigoAbreviacion:SOL_INI",
				):

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.Solicitud{
								{
									Id:        1,
									TerceroId: 999,
								},
							},
						}

					return nil

				case strings.Contains(rawURL, "tipo_solicitud"):

					*(target.(*models.ResponseListaTipoSolicitud)) =
						models.ResponseListaTipoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.TipoSolicitud{
								{
									Id: 2,
								},
							},
						}

					return nil

				case strings.Contains(rawURL, "estado_documento"):

					*(target.(*models.ResponseListaEstadoDocumento)) =
						models.ResponseListaEstadoDocumento{
							Success: true,
							Status:  "200",
							Data: []models.EstadoDocumento{
								{
									Id: 3,
								},
							},
						}

					return nil

				case strings.Contains(rawURL, "estado_solicitud"):

					*(target.(*models.ResponseListaEstadoSolicitud)) =
						models.ResponseListaEstadoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.EstadoSolicitud{
								{
									Id: 4,
								},
							},
						}

					return nil

				case strings.Contains(rawURL, "tipo_documento_solicitud"):

					*(target.(*models.ResponseListaTipoDocumentoSolicitud)) =
						models.ResponseListaTipoDocumentoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.TipoDocumentoSolicitud{
								{
									Id:                11,
									CodigoAbreviacion: "SOL_PRO_CARTA",
								},
							},
						}

					return nil
				}

				return nil
			},
		)

		monkey.Patch(
			request.SendJson,
			func(
				rawURL string,
				method string,
				response interface{},
				data interface{},
			) error {

				switch {

				case strings.Contains(
					rawURL,
					"historico_estado_solicitud",
				):

					*(response.(*models.ResponseCreateHistoricoEstadoSolicitud)) =
						models.ResponseCreateHistoricoEstadoSolicitud{
							Success: true,
							Status:  "201",
							Data: models.HistoricoEstadoSolicitud{
								Id: 200,
							},
						}

					return nil

				case strings.Contains(rawURL, "solicitud") &&
					!strings.Contains(rawURL, "historico") &&
					!strings.Contains(rawURL, "documento"):

					*(response.(*models.ResponseCreateSolicitud)) =
						models.ResponseCreateSolicitud{
							Success: true,
							Status:  "201",
							Data: models.Solicitud{
								Id: 100,
							},
						}

					return nil
				}

				return nil
			},
		)

		monkey.Patch(
			helpers.CrearDocumento,
			func(
				documentos []models.CrearDocumentoGestorDocumental,
			) ([]map[string]interface{}, map[string]interface{}) {

				return nil, map[string]interface{}{
					"error": "error gestor documental",
				}
			},
		)

		entrada := models.CrearSolicitudProrrogaEntrada{
			ComisionId: 10,
			DocumentosSolicitudProrroga: []models.DocumentoProrroga{
				{
					CodigoAbreviacionDoc: "SOL_PRO_CARTA",
				},
			},
		}

		resultado, err := services.CrearSolicitudProrroga(entrada)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if err.Error() != "error creando documentos" {
			t.Errorf(
				"se esperaba 'error creando documentos' y se obtuvo %v",
				err.Error(),
			)
		}

		if resultado.SolicitudProrrogaId != 0 {
			t.Errorf(
				"se esperaba SolicitudProrrogaId 0 y se obtuvo %d",
				resultado.SolicitudProrrogaId,
			)
		}
	})
}
