package services_test

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
// TEST VALIDAR PUEDE CREAR SOLICITUD CIERRE
// =====================================================
func TestValidarPuedeCrearSolicitudCierre(t *testing.T) {

	t.Run("Caso 1: no existen solicitudes y puede crear", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

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
			services.ValidarPuedeCrearSolicitudCierre(10)

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

	t.Run("Caso 2: existe solicitud en estado aprobada y no puede crear", func(t *testing.T) {

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

				case strings.Contains(
					rawURL,
					"solicitud?query=",
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
				}

				return errors.New("url no esperada")
			},
		)

		puedeCrear, estado, err :=
			services.ValidarPuedeCrearSolicitudCierre(10)

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

	t.Run("Caso 3: todas las solicitudes están no aprobadas", func(t *testing.T) {

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

				case strings.Contains(
					rawURL,
					"solicitud?query=",
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
				}

				return errors.New("url no esperada")
			},
		)

		puedeCrear, estado, err :=
			services.ValidarPuedeCrearSolicitudCierre(10)

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

		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				return errors.New(
					"servicio no disponible",
				)
			},
		)

		puedeCrear, estado, err :=
			services.ValidarPuedeCrearSolicitudCierre(10)

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

		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				if strings.Contains(
					rawURL,
					"solicitud?query=",
				) {

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
			services.ValidarPuedeCrearSolicitudCierre(10)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if !strings.Contains(
			err.Error(),
			"error consultando solicitudes de cierre",
		) {
			t.Errorf(
				"mensaje inesperado: %s",
				err.Error(),
			)
		}
	})

	t.Run("Caso 6: respuesta inesperada consultando histórico", func(t *testing.T) {

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
					"historico_estado_solicitud",
				):

					*(target.(*models.ResponseListaHistoricoEstadoSolicitud)) =
						models.ResponseListaHistoricoEstadoSolicitud{
							Success: true,
							Status:  "500",
						}

					return nil

				case strings.Contains(
					rawURL,
					"solicitud?query=",
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
				}

				return errors.New("url no esperada")
			},
		)

		_, _, err :=
			services.ValidarPuedeCrearSolicitudCierre(10)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if !strings.Contains(
			err.Error(),
			"respuesta inesperada consultando histórico de solicitudes de cierre",
		) {
			t.Errorf(
				"mensaje inesperado: %s",
				err.Error(),
			)
		}
	})

	t.Run("Caso 7: histórico vacío permite crear", func(t *testing.T) {

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
					"historico_estado_solicitud",
				):

					*(target.(*models.ResponseListaHistoricoEstadoSolicitud)) =
						models.ResponseListaHistoricoEstadoSolicitud{
							Success: true,
							Status:  "200",
							Data:    []models.HistoricoEstadoSolicitud{},
						}

					return nil

				case strings.Contains(
					rawURL,
					"solicitud?query=",
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
				}

				return errors.New("url no esperada")
			},
		)

		puedeCrear, estado, err :=
			services.ValidarPuedeCrearSolicitudCierre(10)

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
// TEST CONSULTAR HISTORICO DE SOLICITUDES DE CIERRE
// =====================================================

func TestConsultarHistoricoSolicitudesCierre(t *testing.T) {

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
					"historico_estado_solicitud",
				):

					*(target.(*models.ResponseListaHistoricoEstadoSolicitud)) =
						models.ResponseListaHistoricoEstadoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.HistoricoEstadoSolicitud{
								{
									Id: 10,
									SolicitudId: &models.Solicitud{
										Id:            100,
										FechaCreacion: "2024-01-15",
									},
									EstadoSolicitudId: &models.EstadoSolicitud{
										Nombre: "Aprobada",
									},
								},
							},
						}

					return nil

				case strings.Contains(
					rawURL,
					"solicitud?query=",
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
				}

				return errors.New("url no esperada")
			},
		)

		resultado, err :=
			services.ConsultarHistoricoSolicitudesCierre(1)

		if err != nil {
			t.Fatalf(
				"no se esperaba error y se obtuvo %v",
				err,
			)
		}

		if len(resultado) != 1 {
			t.Fatalf(
				"se esperaba 1 resultado y llegaron %d",
				len(resultado),
			)
		}

		if resultado[0].SolicitudId != 100 {
			t.Errorf(
				"se esperaba solicitud 100 y llegó %d",
				resultado[0].SolicitudId,
			)
		}

		if resultado[0].HistoricoId != 10 {
			t.Errorf(
				"se esperaba histórico 10 y llegó %d",
				resultado[0].HistoricoId,
			)
		}

		if resultado[0].Estado != "Aprobada" {
			t.Errorf(
				"se esperaba estado Aprobada y llegó %s",
				resultado[0].Estado,
			)
		}
	})

	t.Run("Caso 2: no existen solicitudes", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				if strings.Contains(
					rawURL,
					"solicitud?query=",
				) {

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

		resultado, err :=
			services.ConsultarHistoricoSolicitudesCierre(1)

		if err != nil {
			t.Fatalf(
				"no se esperaba error y se obtuvo %v",
				err,
			)
		}

		if len(resultado) != 0 {
			t.Errorf(
				"se esperaba lista vacía y llegaron %d registros",
				len(resultado),
			)
		}
	})

	t.Run("Caso 3: error consultando solicitudes", func(t *testing.T) {

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

		_, err :=
			services.ConsultarHistoricoSolicitudesCierre(1)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if err.Error() != "servicio no disponible" {
			t.Errorf(
				"error inesperado: %s",
				err.Error(),
			)
		}
	})

	t.Run("Caso 4: respuesta de solicitudes con Success false", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				if strings.Contains(
					rawURL,
					"solicitud?query=",
				) {

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

		_, err :=
			services.ConsultarHistoricoSolicitudesCierre(1)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if !strings.Contains(
			err.Error(),
			"error consultando solicitudes de cierre",
		) {
			t.Errorf(
				"mensaje inesperado: %s",
				err.Error(),
			)
		}
	})

	t.Run("Caso 5: respuesta de solicitudes con status diferente de 200", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				if strings.Contains(
					rawURL,
					"solicitud?query=",
				) {

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: true,
							Status:  "500",
						}

					return nil
				}

				return errors.New("url no esperada")
			},
		)

		_, err :=
			services.ConsultarHistoricoSolicitudesCierre(1)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if !strings.Contains(
			err.Error(),
			"respuesta inesperada consultando solicitudes de cierre",
		) {
			t.Errorf(
				"mensaje inesperado: %s",
				err.Error(),
			)
		}
	})

	t.Run("Caso 6: error consultando histórico", func(t *testing.T) {

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
					"historico_estado_solicitud",
				):
					return errors.New("error histórico")

				case strings.Contains(
					rawURL,
					"solicitud?query=",
				):

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.Solicitud{
								{Id: 100},
							},
						}

					return nil
				}

				return errors.New("url no esperada")
			},
		)

		_, err :=
			services.ConsultarHistoricoSolicitudesCierre(1)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if err.Error() != "error histórico" {
			t.Errorf(
				"error inesperado: %s",
				err.Error(),
			)
		}
	})

	t.Run("Caso 7: histórico con Success false", func(t *testing.T) {

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
					"historico_estado_solicitud",
				):

					*(target.(*models.ResponseListaHistoricoEstadoSolicitud)) =
						models.ResponseListaHistoricoEstadoSolicitud{
							Success: false,
							Status:  "500",
						}

					return nil

				case strings.Contains(
					rawURL,
					"solicitud?query=",
				):

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.Solicitud{
								{Id: 100},
							},
						}

					return nil
				}

				return errors.New("url no esperada")
			},
		)

		_, err :=
			services.ConsultarHistoricoSolicitudesCierre(1)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if !strings.Contains(
			err.Error(),
			"error consultando histórico de solicitudes de cierre",
		) {
			t.Errorf(
				"mensaje inesperado: %s",
				err.Error(),
			)
		}
	})

	t.Run("Caso 8: histórico con status diferente de 200", func(t *testing.T) {

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
					"historico_estado_solicitud",
				):

					*(target.(*models.ResponseListaHistoricoEstadoSolicitud)) =
						models.ResponseListaHistoricoEstadoSolicitud{
							Success: true,
							Status:  "500",
						}

					return nil

				case strings.Contains(
					rawURL,
					"solicitud?query=",
				):

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.Solicitud{
								{Id: 100},
							},
						}

					return nil
				}

				return errors.New("url no esperada")
			},
		)

		_, err :=
			services.ConsultarHistoricoSolicitudesCierre(1)

		if err == nil {
			t.Fatal("se esperaba error y no se obtuvo")
		}

		if !strings.Contains(
			err.Error(),
			"respuesta inesperada consultando histórico de solicitudes de cierre",
		) {
			t.Errorf(
				"mensaje inesperado: %s",
				err.Error(),
			)
		}
	})

	t.Run("Caso 9: histórico vacío no agrega registros", func(t *testing.T) {

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
					"historico_estado_solicitud",
				):

					*(target.(*models.ResponseListaHistoricoEstadoSolicitud)) =
						models.ResponseListaHistoricoEstadoSolicitud{
							Success: true,
							Status:  "200",
							Data:    []models.HistoricoEstadoSolicitud{},
						}

					return nil

				case strings.Contains(
					rawURL,
					"solicitud?query=",
				):

					*(target.(*models.ResponseListaSolicitud)) =
						models.ResponseListaSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.Solicitud{
								{Id: 100},
							},
						}

					return nil
				}

				return errors.New("url no esperada")
			},
		)

		resultado, err :=
			services.ConsultarHistoricoSolicitudesCierre(1)

		if err != nil {
			t.Fatalf(
				"no se esperaba error y se obtuvo %v",
				err,
			)
		}

		if len(resultado) != 0 {
			t.Errorf(
				"se esperaba lista vacía y llegaron %d registros",
				len(resultado),
			)
		}
	})
}

// =====================================================
// TEST CREAR SOLICITUD CIERRE
// =====================================================
func TestCrearSolicitudCierre(t *testing.T) {

	t.Run("Caso 1: creación exitosa", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

		monkey.Patch(
			services.ValidarPuedeCrearSolicitudCierre,
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
									Id:        100,
									TerceroId: 999,
								},
							},
						}

					return nil

				case strings.Contains(
					rawURL,
					"detalle_solicitud",
				):

					*(target.(*models.ResponseDetalleSolicitud)) =
						models.ResponseDetalleSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.DetalleSolicitud{
								{
									Formulario: `{
										"solicitante":{
											"q2_facultad":"Ingenieria",
											"q3_nombres_apellidos":"Juan Perez",
											"q4_documento_identificacion":"123"
										}
									}`,
								},
							},
						}

					return nil

				case strings.Contains(
					rawURL,
					"tipo_solicitud",
				):

					*(target.(*models.ResponseListaTipoSolicitud)) =
						models.ResponseListaTipoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.TipoSolicitud{
								{
									Id: 5,
								},
							},
						}

					return nil

				case strings.Contains(
					rawURL,
					"estado_solicitud",
				):

					*(target.(*models.ResponseListaEstadoSolicitud)) =
						models.ResponseListaEstadoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.EstadoSolicitud{
								{
									Id: 8,
								},
							},
						}

					return nil
				}

				return errors.New("url no esperada")
			},
		)

		monkey.Patch(
			request.SendJson,
			func(
				url string,
				method string,
				target interface{},
				data interface{},
			) error {

				switch resp := target.(type) {

				case *models.ResponseCreateSolicitud:

					*resp = models.ResponseCreateSolicitud{
						Success: true,
						Status:  "201",
						Data: models.Solicitud{
							Id: 500,
						},
					}

				case *models.ResponseCreateDetalleSolicitud:

					*resp = models.ResponseCreateDetalleSolicitud{
						Success: true,
						Status:  "201",
						Data: models.Solicitud{
							Id: 600,
						},
					}

				case *models.ResponseCreateHistoricoEstadoSolicitud:

					*resp = models.ResponseCreateHistoricoEstadoSolicitud{
						Success: true,
						Status:  "201",
						Data: models.HistoricoEstadoSolicitud{
							Id: 700,
						},
					}

				default:
					return errors.New("tipo no esperado")
				}

				return nil
			},
		)

		resp, err := services.CrearSolicitudCierre(
			models.CrearSolicitudCierreEntrada{
				ComisionId:           10,
				Observacion:          "Observacion",
				CodigoAbreviacionRol: "DOC",
			},
		)

		if err != nil {
			t.Fatalf(
				"no se esperaba error y se obtuvo %v",
				err,
			)
		}

		if resp.ComisionId != 10 {
			t.Errorf(
				"se esperaba comision 10 y llegó %d",
				resp.ComisionId,
			)
		}

		if resp.SolicitudCierreId != 500 {
			t.Errorf(
				"se esperaba solicitud 500 y llegó %d",
				resp.SolicitudCierreId,
			)
		}
	})
}


// =====================================================
// TEST RECHAZAR SOLICITUD CIERRE
// =====================================================
func TestRechazarSolicitudCierre(t *testing.T) {

	t.Run("Caso 1: rechazo exitoso con observación", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				switch {

				// IMPORTANTE:
				// debe ir primero porque contiene la cadena
				// "estado_solicitud"
				case strings.Contains(
					rawURL,
					"historico_estado_solicitud",
				):

					*(target.(*models.ResponseListaHistoricoEstadoSolicitudPUT)) =
						models.ResponseListaHistoricoEstadoSolicitudPUT{
							Success: true,
							Status:  "200",
							Data: []models.HistoricoEstadoSolicitudPUT{
								{
									Id: 50,

									SolicitudId: &models.Solicitud{
										Id: 100,
									},

									EstadoSolicitudId: &models.EstadoSolicitud{
										Id: 1,
									},

									RolUsuario:    "DEC",
									TerceroId:     999,
									FechaCreacion: "2025-01-01",
									Activo:        true,
								},
							},
						}

					return nil

				case strings.Contains(
					rawURL,
					"estado_solicitud",
				):

					*(target.(*models.ResponseListaEstadoSolicitud)) =
						models.ResponseListaEstadoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.EstadoSolicitud{
								{
									Id:     10,
									Nombre: "No aprobada",
								},
							},
						}

					return nil
				}

				return errors.New("url no esperada")
			},
		)

		monkey.Patch(
			request.SendJson,
			func(
				url string,
				method string,
				target interface{},
				data interface{},
			) error {

				switch resp := target.(type) {

				// PUT histórico
				case *map[string]interface{}:

					*resp = map[string]interface{}{
						"Success": true,
					}

					return nil

				// POST nuevo histórico
				case *models.ResponseCreateHistoricoEstadoSolicitud:

					*resp =
						models.ResponseCreateHistoricoEstadoSolicitud{
							Success: true,
							Status:  "201",
							Data: models.HistoricoEstadoSolicitud{
								Id: 200,
							},
						}

					return nil
				}

				return errors.New("tipo no esperado")
			},
		)

		resp, err := services.RechazarSolicitudCierre(
			models.CierreAprobacionSolicitud{
				SolicitudId: 100,
				ComisionId:  10,
				Observacion: "Observación de rechazo",
				TerceroId:   999,
				RolUsuario:  "DEC",
			},
		)

		if err != nil {
			t.Fatalf(
				"no se esperaba error y se obtuvo %v",
				err,
			)
		}

		if resp.SolicitudCierreId != 100 {

			t.Errorf(
				"se esperaba solicitud 100 y llegó %d",
				resp.SolicitudCierreId,
			)
		}
	})
}


// =====================================================
// TEST APROBAR SOLICITUD CIERRE
// =====================================================
func TestAprobarSolicitudCierre(t *testing.T) {

	t.Run("Caso 1: aprobación exitosa con observación", func(t *testing.T) {

		defer monkey.UnpatchAll()

		_ = beego.AppConfig.Set(
			"UrlComisionesCrud",
			"http://comisiones/",
		)

		monkey.Patch(
			request.GetJson,
			func(rawURL string, target interface{}) error {

				switch {

				// IMPORTANTE:
				// historico_estado_solicitud contiene estado_solicitud
				case strings.Contains(
					rawURL,
					"historico_estado_solicitud",
				):

					*(target.(*models.ResponseListaHistoricoEstadoSolicitudPUT)) =
						models.ResponseListaHistoricoEstadoSolicitudPUT{
							Success: true,
							Status:  "200",
							Data: []models.HistoricoEstadoSolicitudPUT{
								{
									Id: 50,

									SolicitudId: &models.Solicitud{
										Id: 100,
									},

									EstadoSolicitudId: &models.EstadoSolicitud{
										Id: 1,
									},

									RolUsuario:    "DEC",
									TerceroId:     999,
									FechaCreacion: "2025-01-01",
									Activo:        true,
								},
							},
						}

					return nil

				case strings.Contains(
					rawURL,
					"historico_estado_comision",
				):

					*(target.(*models.ResponseListaHistoricoEstadoComisionPUT)) =
						models.ResponseListaHistoricoEstadoComisionPUT{
							Success: true,
							Status:  "200",
							Data: []models.HistoricoEstadoComisionPUT{
								{
									Id: 60,

									ComisionId: &models.Comision{
										Id: 10,
									},

									EstadoComisionId: &models.EstadoComision{
										Id: 2,
									},

									RolUsuario:    "DEC",
									TerceroId:     999,
									FechaCreacion: "2025-01-01",
									Activo:        true,
								},
							},
						}

					return nil

				case strings.Contains(
					rawURL,
					"estado_comision",
				):

					*(target.(*models.ResponseListaEstadoComision)) =
						models.ResponseListaEstadoComision{
							Success: true,
							Status:  "200",
							Data: []models.EstadoComision{
								{
									Id:     20,
									Nombre: "Comisión Finalizada",
								},
							},
						}

					return nil

				case strings.Contains(
					rawURL,
					"estado_solicitud",
				):

					*(target.(*models.ResponseListaEstadoSolicitud)) =
						models.ResponseListaEstadoSolicitud{
							Success: true,
							Status:  "200",
							Data: []models.EstadoSolicitud{
								{
									Id:     10,
									Nombre: "Aprobada",
								},
							},
						}

					return nil
				}

				return errors.New("url no esperada")
			},
		)

		monkey.Patch(
			request.SendJson,
			func(
				url string,
				method string,
				target interface{},
				data interface{},
			) error {

				switch resp := target.(type) {

				// PUT historico solicitud
				// PUT historico comision
				// POST observacion
				case *map[string]interface{}:

					*resp = map[string]interface{}{
						"Success": true,
					}

					return nil

				case *models.ResponseCreateHistoricoEstadoSolicitud:

					*resp =
						models.ResponseCreateHistoricoEstadoSolicitud{
							Success: true,
							Status:  "201",
							Data: models.HistoricoEstadoSolicitud{
								Id: 200,
							},
						}

					return nil

				case *models.ResponseCreateHistoricoEstadoComision:

					*resp =
						models.ResponseCreateHistoricoEstadoComision{
							Success: true,
							Status:  "201",
							Data: models.HistoricoEstadoComision{
								Id: 300,
							},
						}

					return nil
				}

				return errors.New("tipo no esperado")
			},
		)

		resp, err := services.AprobarSolicitudCierre(
			models.CierreAprobacionSolicitud{
				SolicitudId: 100,
				ComisionId:  10,
				Observacion: "Aprobada correctamente",
				TerceroId:   999,
				RolUsuario:  "DEC",
			},
		)

		if err != nil {
			t.Fatalf(
				"no se esperaba error y se obtuvo %v",
				err,
			)
		}

		if resp.SolicitudCierreId != 100 {

			t.Errorf(
				"se esperaba solicitud 100 y llegó %d",
				resp.SolicitudCierreId,
			)
		}
	})
}