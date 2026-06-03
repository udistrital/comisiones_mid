package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/utils_oas/request"
)

func ValidarPuedeCrearSolicitudCierre(comisionId int) (bool, string, error) {

	// =========================
	// CONSULTAR SOLICITUDES DE CIERRE
	// =========================

	var responseBusquedaSolicitudCierre models.ResponseListaSolicitud

	err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"solicitud?query=TipoSolicitudId__CodigoAbreviacion:SOL_CIERRE,ComisionId__Id:"+
			fmt.Sprintf("%d", comisionId),
		&responseBusquedaSolicitudCierre,
	)

	if err != nil {
		return false, "", err
	}

	if !responseBusquedaSolicitudCierre.Success {
		return false, "",
			fmt.Errorf(
				"error consultando solicitudes de cierre: status %s",
				responseBusquedaSolicitudCierre.Status,
			)
	}

	if responseBusquedaSolicitudCierre.Status != "200" {
		return false, "",
			fmt.Errorf(
				"respuesta inesperada consultando solicitudes de cierre: %s",
				responseBusquedaSolicitudCierre.Status,
			)
	}

	// =========================
	// SI NO HAY SOLICITUDES
	// =========================

	solicitudes := responseBusquedaSolicitudCierre.Data

	if len(solicitudes) == 0 {
		return true, "", nil
	}

	// =========================
	// VALIDAR ÚLTIMO ESTADO
	// =========================

	for _, solicitud := range solicitudes {

		var responseHistorico models.ResponseListaHistoricoEstadoSolicitud

		err = request.GetJson(
			beego.AppConfig.String("UrlComisionesCrud")+
				"historico_estado_solicitud?query=solicitud_id:"+
				fmt.Sprintf("%d", solicitud.Id)+
				"&sortby=fecha_creacion&order=desc&limit=1",
			&responseHistorico,
		)

		if err != nil {
			return false, "", err
		}

		if !responseHistorico.Success {
			return false, "",
				fmt.Errorf(
					"error consultando histórico de solicitudes de cierre: status %s",
					responseHistorico.Status,
				)
		}

		if responseHistorico.Status != "200" {
			return false, "",
				fmt.Errorf(
					"respuesta inesperada consultando histórico de solicitudes de cierre: %s",
					responseHistorico.Status,
				)
		}

		// SI NO TIENE HISTÓRICO CONTINÚA
		if len(responseHistorico.Data) == 0 {
			continue
		}

		estado := strings.TrimSpace(
			strings.ToLower(
				responseHistorico.Data[0].EstadoSolicitudId.Nombre,
			),
		)

		// =========================
		// SI NO ESTÁ RECHAZADA
		// =========================

		if estado != "no aprobada" {

			return false,
				responseHistorico.Data[0].EstadoSolicitudId.Nombre,
				nil
		}
	}

	// =========================
	// TODAS ESTÁN RECHAZADAS
	// =========================

	return true, "", nil
}

func ConsultarHistoricoSolicitudesCierre(
	comisionId int,
) ([]models.HistoricoSolicitudCierre, error) {

	// =========================
	// CONSULTAR SOLICITUDES DE CIERRE
	// =========================

	var responseBusquedaSolicitudCierre models.ResponseListaSolicitud

	err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"solicitud?query=TipoSolicitudId__CodigoAbreviacion:SOL_CIERRE,ComisionId__Id:"+
			fmt.Sprintf("%d", comisionId),
		&responseBusquedaSolicitudCierre,
	)

	if err != nil {
		return nil, err
	}

	if !responseBusquedaSolicitudCierre.Success {
		return nil,
			fmt.Errorf(
				"error consultando solicitudes de cierre: status %s",
				responseBusquedaSolicitudCierre.Status,
			)
	}

	if responseBusquedaSolicitudCierre.Status != "200" {
		return nil,
			fmt.Errorf(
				"respuesta inesperada consultando solicitudes de cierre: %s",
				responseBusquedaSolicitudCierre.Status,
			)
	}

	// =========================
	// ARMAR HISTÓRICO
	// =========================

	historicoSolicitudes := []models.HistoricoSolicitudCierre{}

	for _, solicitud := range responseBusquedaSolicitudCierre.Data {

		// =========================
		// CONSULTAR ÚLTIMO HISTÓRICO
		// =========================

		var responseHistorico models.ResponseListaHistoricoEstadoSolicitud

		err = request.GetJson(
			beego.AppConfig.String("UrlComisionesCrud")+
				"historico_estado_solicitud?query=solicitud_id:"+
				fmt.Sprintf("%d", solicitud.Id)+
				"&sortby=fecha_creacion&order=desc&limit=1",
			&responseHistorico,
		)

		if err != nil {
			return nil, err
		}

		if !responseHistorico.Success {
			return nil,
				fmt.Errorf(
					"error consultando histórico de solicitudes de cierre: status %s",
					responseHistorico.Status,
				)
		}

		if responseHistorico.Status != "200" {
			return nil,
				fmt.Errorf(
					"respuesta inesperada consultando histórico de solicitudes de cierre: %s",
					responseHistorico.Status,
				)
		}

		// =========================
		// SI NO TIENE HISTÓRICO
		// =========================

		if len(responseHistorico.Data) == 0 {
			continue
		}

		ultimoHistorico := responseHistorico.Data[0]

		// =========================
		// AGREGAR A LISTA
		// =========================

		historicoSolicitudes = append(
			historicoSolicitudes,
			models.HistoricoSolicitudCierre{
				SolicitudId:   solicitud.Id,
				HistoricoId:   ultimoHistorico.Id,
				FechaCreacion: ultimoHistorico.SolicitudId.FechaCreacion,
				Estado:        ultimoHistorico.EstadoSolicitudId.Nombre,
			},
		)
	}

	return historicoSolicitudes, nil
}

func CrearSolicitudCierre(comisionCierre models.CrearSolicitudCierreEntrada) (cierre models.CrearSolicitudCierreSalida, err error) {

	puedeCrear, estadoActual, err := ValidarPuedeCrearSolicitudCierre(
		comisionCierre.ComisionId,
	)

	if err != nil {
		return models.CrearSolicitudCierreSalida{}, err
	}

	if !puedeCrear {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"El maestro ya tiene una solicitud de cierre en estado %s",
				estadoActual,
			)
	}

	// =========================
	// CONSULTAR SOLICITUD BASE
	// =========================

	var responseSolicitud models.ResponseListaSolicitud

	err = request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"solicitud?query=ComisionId__Id:"+
			fmt.Sprintf("%d", comisionCierre.ComisionId)+
			",TipoSolicitudId__CodigoAbreviacion:SOL_INI",
		&responseSolicitud,
	)

	if err != nil {
		return models.CrearSolicitudCierreSalida{}, err
	}

	if !responseSolicitud.Success {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"error consultando solicitud base: status %s",
				responseSolicitud.Status,
			)
	}

	if responseSolicitud.Status != "200" {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"respuesta inesperada consultando solicitud base: %s",
				responseSolicitud.Status,
			)
	}

	if len(responseSolicitud.Data) != 1 {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"se esperaba 1 solicitud y llegaron %d",
				len(responseSolicitud.Data),
			)
	}

	solicitudComision := responseSolicitud.Data[0]

	// =========================
	// CONSULTAR DETALLE SOLICITUD BASE
	// =========================

	var responseDetalleSolicitud models.ResponseDetalleSolicitud

	err = request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"detalle_solicitud?query=SolicitudId__Id:"+
			fmt.Sprintf("%d", solicitudComision.Id),
		&responseDetalleSolicitud,
	)

	if err != nil {
		return models.CrearSolicitudCierreSalida{}, err
	}

	if !responseDetalleSolicitud.Success {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"error consultando detalle solicitud base: status %s",
				responseDetalleSolicitud.Status,
			)
	}

	if responseDetalleSolicitud.Status != "200" {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"respuesta inesperada consultando detalle solicitud base: %s",
				responseDetalleSolicitud.Status,
			)
	}

	if len(responseDetalleSolicitud.Data) != 1 {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"se esperaba 1 detalle solicitud base y llegaron %d",
				len(responseDetalleSolicitud.Data),
			)
	}

	detalleSolicitudBase := responseDetalleSolicitud.Data[0]

	// =========================
	// CONSULTAR TIPO SOLICITUD
	// =========================

	var responseTipo models.ResponseListaTipoSolicitud

	err = request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"tipo_solicitud?query=CodigoAbreviacion:SOL_CIERRE",
		&responseTipo,
	)

	if err != nil {
		return models.CrearSolicitudCierreSalida{}, err
	}

	if !responseTipo.Success {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"error consultando tipo solicitud: status %s",
				responseTipo.Status,
			)
	}

	if responseTipo.Status != "200" {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"respuesta inesperada consultando tipo solicitud: %s",
				responseTipo.Status,
			)
	}

	if len(responseTipo.Data) != 1 {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"se esperaba 1 tipo de solicitud y llegaron %d",
				len(responseTipo.Data),
			)
	}

	tipoSolicitud := responseTipo.Data[0]

	// =========================
	// CONSULTAR ESTADO SOLICITUD
	// =========================

	var responseEstado models.ResponseListaEstadoSolicitud

	err = request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"estado_solicitud?query=CodigoAbreviacion:REV_DEC",
		&responseEstado,
	)

	if err != nil {
		return models.CrearSolicitudCierreSalida{}, err
	}

	if !responseEstado.Success {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"error consultando estado solicitud: status %s",
				responseEstado.Status,
			)
	}

	if responseEstado.Status != "200" {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"respuesta inesperada consultando estado solicitud: %s",
				responseEstado.Status,
			)
	}

	if len(responseEstado.Data) != 1 {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"se esperaba 1 estado de solicitud y llegaron %d",
				len(responseEstado.Data),
			)
	}

	estadoSolicitud := responseEstado.Data[0]

	// =========================
	// CREAR SOLICITUD
	// =========================

	req := models.SolicitudCreateRequest{
		TerceroId: solicitudComision.TerceroId,

		ComisionId: &models.IdReference{
			Id: comisionCierre.ComisionId,
		},

		Activo: true,

		TipoSolicitudId: models.IdReference{
			Id: tipoSolicitud.Id,
		},

		ObservacionCierre: comisionCierre.Observacion,
	}

	var respSolicitud models.ResponseCreateSolicitud

	err = request.SendJson(
		beego.AppConfig.String("UrlComisionesCrud")+"solicitud",
		"POST",
		&respSolicitud,
		&req,
	)

	if err != nil {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf("error creando solicitud: %v", err)
	}

	if !respSolicitud.Success {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"el servicio respondió con status %s",
				respSolicitud.Status,
			)
	}

	if respSolicitud.Status != "201" {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"respuesta inesperada creando solicitud: %s",
				respSolicitud.Status,
			)
	}

	if respSolicitud.Data.Id == 0 {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf("la solicitud creada no retornó id")
	}

	// =========================
	// CREAR DETALLE SOLICITUD
	// =========================

	var formularioOriginal map[string]interface{}

	err = json.Unmarshal(
		[]byte(detalleSolicitudBase.Formulario),
		&formularioOriginal,
	)
	if err != nil {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf("error parseando formulario: %v", err)
	}
	solicitante, ok := formularioOriginal["solicitante"].(map[string]interface{})
	if !ok {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf("no se encontró la sección solicitante")
	}

	nuevoFormulario := map[string]interface{}{
		"solicitante": map[string]interface{}{
			"q2_facultad":                 solicitante["q2_facultad"],
			"q3_nombres_apellidos":        solicitante["q3_nombres_apellidos"],
			"q4_documento_identificacion": solicitante["q4_documento_identificacion"],
		},
		"formulario_completado": false,
	}
	nuevoFormularioBytes, err := json.Marshal(nuevoFormulario)
	if err != nil {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf("error serializando formulario: %v", err)
	}

	detalleSolicitudProrroga := models.DetalleSolicitud{
		SolicitudId: &models.Solicitud{
			Id: respSolicitud.Data.Id,
		},
		Formulario: string(nuevoFormularioBytes),
		Activo:     true,
	}

	var respDetalleSolicitud models.ResponseCreateDetalleSolicitud

	err = request.SendJson(
		beego.AppConfig.String("UrlComisionesCrud")+"detalle_solicitud",
		"POST",
		&respDetalleSolicitud,
		&detalleSolicitudProrroga,
	)

	if err != nil {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf("error creando detalle solicitud: %v", err)
	}

	if !respDetalleSolicitud.Success {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"el servicio respondió con status %s",
				respDetalleSolicitud.Status,
			)
	}

	if respDetalleSolicitud.Status != "201" {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"respuesta inesperada creando detalle solicitud: %s",
				respDetalleSolicitud.Status,
			)
	}

	if respDetalleSolicitud.Data.Id == 0 {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf("el detalle de solicitud creada no retornó id")
	}

	// =========================
	// CREAR HISTORICO
	// =========================

	historico := models.HistoricoEstadoSolicitud{
		SolicitudId: &models.Solicitud{
			Id: respSolicitud.Data.Id,
		},

		EstadoSolicitudId: &models.EstadoSolicitud{
			Id: estadoSolicitud.Id,
		},

		RolUsuario: comisionCierre.CodigoAbreviacionRol,
		TerceroId:  solicitudComision.TerceroId,
		Activo:     true,
	}

	var respHistorico models.ResponseCreateHistoricoEstadoSolicitud

	err = request.SendJson(
		beego.AppConfig.String("UrlComisionesCrud")+"historico_estado_solicitud",
		"POST",
		&respHistorico,
		&historico,
	)

	if err != nil {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf("error creando historico: %v", err)
	}

	if !respHistorico.Success {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"el servicio de histórico respondió con status %s",
				respHistorico.Status,
			)
	}

	if respHistorico.Status != "201" {
		return models.CrearSolicitudCierreSalida{},
			fmt.Errorf(
				"respuesta inesperada creando histórico: %s",
				respHistorico.Status,
			)
	}

	// =========================
	// RESPUESTA
	// =========================

	var salidaCreacionCierre models.CrearSolicitudCierreSalida

	salidaCreacionCierre.ComisionId =
		comisionCierre.ComisionId

	salidaCreacionCierre.SolicitudCierreId =
		respSolicitud.Data.Id

	return salidaCreacionCierre, nil
}
