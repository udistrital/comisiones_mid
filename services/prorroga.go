package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/utils_oas/request"
)

func ValidarPuedeCrearSolicitudProrroga(comisionId int) (bool, string, error) {

	// =========================
	// CONSULTAR SOLICITUDES DE PRÓRROGA
	// =========================

	var responseBusquedaSolicitudProrroga models.ResponseListaSolicitud

	err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"solicitud?query=TipoSolicitudId__CodigoAbreviacion:SOL_PRORROGA,ComisionId__Id:"+
			fmt.Sprintf("%d", comisionId),
		&responseBusquedaSolicitudProrroga,
	)

	if err != nil {
		return false, "", err
	}

	if !responseBusquedaSolicitudProrroga.Success {
		return false, "",
			fmt.Errorf(
				"error consultando solicitudes de prórroga: status %s",
				responseBusquedaSolicitudProrroga.Status,
			)
	}

	if responseBusquedaSolicitudProrroga.Status != "200" {
		return false, "",
			fmt.Errorf(
				"respuesta inesperada consultando solicitudes de prórroga: %s",
				responseBusquedaSolicitudProrroga.Status,
			)
	}

	// =========================
	// SI NO HAY SOLICITUDES
	// =========================

	solicitudes := responseBusquedaSolicitudProrroga.Data

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
					"error consultando histórico de solicitudes de prórroga: status %s",
					responseHistorico.Status,
				)
		}

		if responseHistorico.Status != "200" {
			return false, "",
				fmt.Errorf(
					"respuesta inesperada consultando histórico de solicitudes de prórroga: %s",
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

func ConsultarHistoricoSolicitudesProrroga(
	comisionId int,
) ([]models.HistoricoSolicitudProrroga, error) {

	// =========================
	// CONSULTAR SOLICITUDES DE PRÓRROGA
	// =========================

	var responseBusquedaSolicitudProrroga models.ResponseListaSolicitud

	err := request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"solicitud?query=TipoSolicitudId__CodigoAbreviacion:SOL_PRORROGA,ComisionId__Id:"+
			fmt.Sprintf("%d", comisionId),
		&responseBusquedaSolicitudProrroga,
	)

	if err != nil {
		return nil, err
	}

	if !responseBusquedaSolicitudProrroga.Success {
		return nil,
			fmt.Errorf(
				"error consultando solicitudes de prórroga: status %s",
				responseBusquedaSolicitudProrroga.Status,
			)
	}

	if responseBusquedaSolicitudProrroga.Status != "200" {
		return nil,
			fmt.Errorf(
				"respuesta inesperada consultando solicitudes de prórroga: %s",
				responseBusquedaSolicitudProrroga.Status,
			)
	}

	// =========================
	// ARMAR HISTÓRICO
	// =========================

	historicoSolicitudes := []models.HistoricoSolicitudProrroga{}

	for _, solicitud := range responseBusquedaSolicitudProrroga.Data {

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
					"error consultando histórico de solicitudes de prórroga: status %s",
					responseHistorico.Status,
				)
		}

		if responseHistorico.Status != "200" {
			return nil,
				fmt.Errorf(
					"respuesta inesperada consultando histórico de solicitudes de prórroga: %s",
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
			models.HistoricoSolicitudProrroga{
				SolicitudId:   solicitud.Id,
				HistoricoId:   ultimoHistorico.Id,
				FechaCreacion: ultimoHistorico.SolicitudId.FechaCreacion,
				Estado:        ultimoHistorico.EstadoSolicitudId.Nombre,
			},
		)
	}

	return historicoSolicitudes, nil
}

func CrearSolicitudProrroga(
	solicitudProrroga models.CrearSolicitudProrrogaEntrada,
) (prorroga models.CrearSolicitudProrrogaSalida, err error) {

	puedeCrear, estadoActual, err := ValidarPuedeCrearSolicitudProrroga(
		solicitudProrroga.ComisionId,
	)

	if err != nil {
		return models.CrearSolicitudProrrogaSalida{}, err
	}

	if !puedeCrear {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"El maestro ya tiene una solicitud de prórroga en estado %s",
				estadoActual,
			)
	}

	type TipoDocumentoTemp struct {
		Id                int
		CodigoAbreviacion string
	}

	type DocumentoTemporal struct {
		TipoDocumentoSolicitudId int
		Documento                models.CrearDocumentoGestorDocumental
	}

	// =========================
	// CONSULTAR SOLICITUD BASE
	// =========================

	var responseSolicitud models.ResponseListaSolicitud

	err = request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"solicitud?query=ComisionId__Id:"+
			fmt.Sprintf("%d", solicitudProrroga.ComisionId)+
			",TipoSolicitudId__CodigoAbreviacion:SOL_INI",
		&responseSolicitud,
	)

	if err != nil {
		return models.CrearSolicitudProrrogaSalida{}, err
	}

	if !responseSolicitud.Success {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"error consultando solicitud base: status %s",
				responseSolicitud.Status,
			)
	}

	if responseSolicitud.Status != "200" {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"respuesta inesperada consultando solicitud base: %s",
				responseSolicitud.Status,
			)
	}

	if len(responseSolicitud.Data) != 1 {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"se esperaba 1 solicitud y llegaron %d",
				len(responseSolicitud.Data),
			)
	}

	solicitudComision := responseSolicitud.Data[0]

	// =========================
	// CONSULTAR SOLICITUD BASE
	// =========================

	var responseDetalleSolicitud models.ResponseDetalleSolicitud

	err = request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"detalle_solicitud?query=SolicitudId__Id:"+
			fmt.Sprintf("%d", solicitudComision.Id),
		&responseDetalleSolicitud,
	)

	if err != nil {
		return models.CrearSolicitudProrrogaSalida{}, err
	}

	if !responseDetalleSolicitud.Success {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"error consultando detalle solicitud base: status %s",
				responseDetalleSolicitud.Status,
			)
	}

	if responseDetalleSolicitud.Status != "200" {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"respuesta inesperada consultando detalle solicitud base: %s",
				responseDetalleSolicitud.Status,
			)
	}

	if len(responseDetalleSolicitud.Data) != 1 {
		return models.CrearSolicitudProrrogaSalida{},
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
			"tipo_solicitud?query=CodigoAbreviacion:SOL_PRORROGA",
		&responseTipo,
	)

	if err != nil {
		return models.CrearSolicitudProrrogaSalida{}, err
	}

	if !responseTipo.Success {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"error consultando tipo solicitud: status %s",
				responseTipo.Status,
			)
	}

	if responseTipo.Status != "200" {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"respuesta inesperada consultando tipo solicitud: %s",
				responseTipo.Status,
			)
	}

	if len(responseTipo.Data) != 1 {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"se esperaba 1 tipo de solicitud y llegaron %d",
				len(responseTipo.Data),
			)
	}

	tipoSolicitud := responseTipo.Data[0]

	// =========================
	// CONSULTAR ESTADO DOCUMENTO SOLICITUD
	// =========================

	var responseEstadoDocumentoSolicitud models.ResponseListaEstadoDocumento

	err = request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"estado_documento?query=CodigoAbreviacion:ENV_REV_SEC_GRAL",
		&responseEstadoDocumentoSolicitud,
	)

	if err != nil {
		return models.CrearSolicitudProrrogaSalida{}, err
	}

	if !responseEstadoDocumentoSolicitud.Success {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"error consultando estado documento: status %s",
				responseEstadoDocumentoSolicitud.Status,
			)
	}

	if responseEstadoDocumentoSolicitud.Status != "200" {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"respuesta inesperada consultando estado documento: %s",
				responseEstadoDocumentoSolicitud.Status,
			)
	}

	if len(responseEstadoDocumentoSolicitud.Data) != 1 {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"se esperaba 1 tipo de estado documento y llegaron %d",
				len(responseEstadoDocumentoSolicitud.Data),
			)
	}

	estadoSolicitudDocumento := responseEstadoDocumentoSolicitud.Data[0]

	// =========================
	// CONSULTAR ESTADO SOLICITUD
	// =========================

	var responseEstado models.ResponseListaEstadoSolicitud

	err = request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"estado_solicitud?query=CodigoAbreviacion:REV_SEC_GRAL",
		&responseEstado,
	)

	if err != nil {
		return models.CrearSolicitudProrrogaSalida{}, err
	}

	if !responseEstado.Success {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"error consultando estado solicitud: status %s",
				responseEstado.Status,
			)
	}

	if responseEstado.Status != "200" {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"respuesta inesperada consultando estado solicitud: %s",
				responseEstado.Status,
			)
	}

	if len(responseEstado.Data) != 1 {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"se esperaba 1 estado de solicitud y llegaron %d",
				len(responseEstado.Data),
			)
	}

	estadoSolicitud := responseEstado.Data[0]

	// =========================
	// CONSULTAR TIPOS DOCUMENTO
	// =========================

	var responseTipoDocumentoSolicitud models.ResponseListaTipoDocumentoSolicitud

	err = request.GetJson(
		beego.AppConfig.String("UrlComisionesCrud")+
			"tipo_documento_solicitud?limit=-1&query=CodigoAbreviacion__startswith:SOL_PRO",
		&responseTipoDocumentoSolicitud,
	)

	if err != nil {
		return models.CrearSolicitudProrrogaSalida{}, err
	}

	if !responseTipoDocumentoSolicitud.Success {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"error consultando tipos documento: status %s",
				responseTipoDocumentoSolicitud.Status,
			)
	}

	if responseTipoDocumentoSolicitud.Status != "200" {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"respuesta inesperada consultando tipos documento: %s",
				responseTipoDocumentoSolicitud.Status,
			)
	}

	if len(responseTipoDocumentoSolicitud.Data) == 0 {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf("no se encontraron tipos de documento")
	}

	var tiposDocumento []TipoDocumentoTemp

	for _, tipo := range responseTipoDocumentoSolicitud.Data {

		tiposDocumento = append(
			tiposDocumento,
			TipoDocumentoTemp{
				Id:                tipo.Id,
				CodigoAbreviacion: tipo.CodigoAbreviacion,
			},
		)
	}

	// =========================
	// VALIDAR DOCUMENTOS
	// =========================

	tiposDocumentoMap := make(map[string]bool)

	for _, tipo := range tiposDocumento {
		tiposDocumentoMap[tipo.CodigoAbreviacion] = true
	}

	var comprobacionDocumentos int

	for _, doc := range solicitudProrroga.DocumentosSolicitudProrroga {

		if tiposDocumentoMap[doc.CodigoAbreviacionDoc] {
			comprobacionDocumentos++
		}
	}

	if comprobacionDocumentos != len(tiposDocumento) {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf("los documentos enviados no coinciden con los requeridos")
	}

	// =========================
	// CREAR SOLICITUD
	// =========================

	req := models.SolicitudCreateRequest{
		TerceroId: solicitudComision.TerceroId,

		ComisionId: &models.IdReference{
			Id: solicitudProrroga.ComisionId,
		},

		Activo: true,

		TipoSolicitudId: models.IdReference{
			Id: tipoSolicitud.Id,
		},

		ObservacionCierre: solicitudProrroga.Observacion,
	}

	var respSolicitud models.ResponseCreateSolicitud

	err = request.SendJson(
		beego.AppConfig.String("UrlComisionesCrud")+"solicitud",
		"POST",
		&respSolicitud,
		&req,
	)

	if err != nil {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf("error creando solicitud: %v", err)
	}

	if !respSolicitud.Success {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"el servicio respondió con status %s",
				respSolicitud.Status,
			)
	}

	if respSolicitud.Status != "201" {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"respuesta inesperada creando solicitud: %s",
				respSolicitud.Status,
			)
	}

	if respSolicitud.Data.Id == 0 {
		return models.CrearSolicitudProrrogaSalida{},
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
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf("error parseando formulario: %v", err)
	}
	solicitante, ok := formularioOriginal["solicitante"].(map[string]interface{})
	if !ok {
		return models.CrearSolicitudProrrogaSalida{},
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
		return models.CrearSolicitudProrrogaSalida{},
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
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf("error creando detalle solicitud: %v", err)
	}

	if !respDetalleSolicitud.Success {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"el servicio respondió con status %s",
				respDetalleSolicitud.Status,
			)
	}

	if respDetalleSolicitud.Status != "201" {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"respuesta inesperada creando detalle solicitud: %s",
				respDetalleSolicitud.Status,
			)
	}

	if respDetalleSolicitud.Data.Id == 0 {
		return models.CrearSolicitudProrrogaSalida{},
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

		RolUsuario: solicitudProrroga.CodigoAbreviacionRol,
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
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf("error creando historico: %v", err)
	}

	if !respHistorico.Success {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"el servicio de histórico respondió con status %s",
				respHistorico.Status,
			)
	}

	if respHistorico.Status != "201" {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf(
				"respuesta inesperada creando histórico: %s",
				respHistorico.Status,
			)
	}

	// =========================
	// PREPARAR DOCUMENTOS
	// =========================

	var documentosTemporales []DocumentoTemporal

	for _, documento := range solicitudProrroga.DocumentosSolicitudProrroga {

		for _, tipo := range tiposDocumento {

			if tipo.CodigoAbreviacion == documento.CodigoAbreviacionDoc {

				documentosTemporales = append(
					documentosTemporales,
					DocumentoTemporal{
						TipoDocumentoSolicitudId: tipo.Id,
						Documento:                documento.DocumentoSolicitud,
					},
				)
			}
		}
	}

	// =========================
	// CREAR DOCUMENTOS
	// =========================

	var documentosCreacionProrroga []models.CrearDocumentoGestorDocumental

	for _, docTemp := range documentosTemporales {

		documentosCreacionProrroga = append(
			documentosCreacionProrroga,
			docTemp.Documento,
		)
	}

	documentosResponse, errDoc :=
		helpers.CrearDocumento(documentosCreacionProrroga)

	if errDoc != nil {
		return models.CrearSolicitudProrrogaSalida{},
			fmt.Errorf("error creando documentos")
	}

	// =========================
	// VINCULAR DOCUMENTOS
	// =========================

	for i, doc := range documentosResponse {

		idDoc := doc["id"].(int)

		documentoSolicitud := models.DocumentoSolicitud{

			DocumentoId: idDoc,

			HistoricoEstadoSolicitudId: &models.HistoricoEstadoSolicitud{
				Id: respHistorico.Data.Id,
			},

			TipoDocumentoId: &models.TipoDocumentoSolicitud{
				Id: documentosTemporales[i].TipoDocumentoSolicitudId,
			},

			EstadoDocumentoId: &models.EstadoDocumento{
				Id: estadoSolicitudDocumento.Id,
			},

			Activo: true,
		}

		var respDoc map[string]interface{}

		err = request.SendJson(
			beego.AppConfig.String("UrlComisionesCrud")+
				"documento_solicitud",
			"POST",
			&respDoc,
			&documentoSolicitud,
		)

		if err != nil {

			return models.CrearSolicitudProrrogaSalida{},
				fmt.Errorf(
					"error vinculando documento %d: %v",
					idDoc,
					err,
				)
		}
	}

	// =========================
	// RESPUESTA
	// =========================

	var salidaCreacionProrroga models.CrearSolicitudProrrogaSalida

	salidaCreacionProrroga.ComisionId =
		solicitudProrroga.ComisionId

	salidaCreacionProrroga.SolicitudProrrogaId =
		respSolicitud.Data.Id

	return salidaCreacionProrroga, nil
}
