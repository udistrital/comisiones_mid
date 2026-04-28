package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/utils_oas/request"
)

func CrearDocumentosSolicitud(
	baseCrud string,
	historicoNuevoId int,
	documentosReq []models.DocumentoCambioEstadoRequest,
) ([]int, []int, error) {
	if historicoNuevoId <= 0 {
		return nil, nil, fmt.Errorf("historicoNuevoId es obligatorio")
	}

	if len(documentosReq) == 0 {
		return []int{}, []int{}, nil
	}

	documentosGestor := make([]models.CrearDocumentoGestorDocumental, 0, len(documentosReq))

	for i, doc := range documentosReq {
		if strings.TrimSpace(doc.Nombre) == "" {
			return nil, nil, fmt.Errorf("Documentos[%d].Nombre es obligatorio", i)
		}
		if strings.TrimSpace(doc.File) == "" {
			return nil, nil, fmt.Errorf("Documentos[%d].File es obligatorio", i)
		}
		if doc.IdTipoDocumento <= 0 {
			return nil, nil, fmt.Errorf("Documentos[%d].IdTipoDocumento es obligatorio", i)
		}

		documentosGestor = append(documentosGestor, models.CrearDocumentoGestorDocumental{
			IdTipoDocumento: doc.IdTipoDocumento,
			Nombre:          strings.TrimSpace(doc.Nombre),
			Descripcion:     strings.TrimSpace(doc.Descripcion),
			Metadatos:       doc.Metadatos,
			File:            strings.TrimSpace(doc.File),
		})
	}

	resultadoDocs, outputError := helpers.CrearDocumento(documentosGestor)
	if outputError != nil {
		return nil, nil, fmt.Errorf("error creando documentos en gestor documental: %v", outputError)
	}

	if len(resultadoDocs) == 0 {
		return nil, nil, fmt.Errorf("no se recibió respuesta con documentos creados")
	}

	if len(resultadoDocs) != len(documentosReq) {
		return nil, nil, fmt.Errorf("la cantidad de documentos creados no coincide con la cantidad enviada")
	}

	documentoIds := make([]int, 0, len(resultadoDocs))
	documentoSolicitudIds := make([]int, 0, len(resultadoDocs))

	for i, resultado := range resultadoDocs {
		documentoId, err := strconv.Atoi(fmt.Sprintf("%v", resultado["id"]))
		if err != nil || documentoId <= 0 {
			return nil, nil, fmt.Errorf("no se pudo extraer el id del documento creado en la posición %d", i)
		}

		var tipoDocumentoSolicitudId int
		if strings.TrimSpace(documentosReq[i].TipoDocumento) != "" {
			tipoDocumentoSolicitudId, err = getIdByCodigoAbreviacion(
				baseCrud,
				"tipo_documento_solicitud",
				documentosReq[i].TipoDocumento,
			)
			if err != nil {
				return nil, nil, fmt.Errorf("no se pudo resolver TipoDocumentoSolicitud del documento %d: %v", i, err)
			}
		}

		var estadoDocumentoId int
		if strings.TrimSpace(documentosReq[i].EstadoDocumento) != "" {
			estadoDocumentoId, err = getIdByCodigoAbreviacion(
				baseCrud,
				"estado_documento",
				documentosReq[i].EstadoDocumento,
			)
			if err != nil {
				return nil, nil, fmt.Errorf("no se pudo resolver EstadoDocumento del documento %d: %v", i, err)
			}
		}

		documentoSolicitudId, err := crearDocumentoSolicitud(
			baseCrud,
			historicoNuevoId,
			documentoId,
			tipoDocumentoSolicitudId,
			estadoDocumentoId,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error creando documento_solicitud para el documento %d: %v", i, err)
		}

		documentoIds = append(documentoIds, documentoId)
		documentoSolicitudIds = append(documentoSolicitudIds, documentoSolicitudId)
	}

	return documentoIds, documentoSolicitudIds, nil
}

func ActualizarEstadoDocumento(req models.ActualizarEstadoDocumentoSolicitudRequest) (models.ActualizarEstadoDocumentoSolicitudResponse, error) {

	if req.DocumentoSolicitudId <= 0 {
		return models.ActualizarEstadoDocumentoSolicitudResponse{}, fmt.Errorf("DocumentoSolicitudId es obligatorio")
	}

	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	logs.Info("UrlComisionesCrud=%q", baseCrud)
	if baseCrud == "" {
		return models.ActualizarEstadoDocumentoSolicitudResponse{}, fmt.Errorf("no esta configurado UrlComisionesCrud")
	}

	getURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/documento_solicitud/%d", req.DocumentoSolicitudId))
	if err := helpers.ValidateAbsoluteURL(getURL); err != nil {
		return models.ActualizarEstadoDocumentoSolicitudResponse{}, err
	}

	var getResp map[string]interface{}
	if err := request.GetJson(getURL, &getResp); err != nil {
		return models.ActualizarEstadoDocumentoSolicitudResponse{}, fmt.Errorf("error consultando documento_solicitud: %v", err)
	}

	obj := helpers.UnwrapDataToMap(getResp)
	if obj == nil {
		return models.ActualizarEstadoDocumentoSolicitudResponse{}, fmt.Errorf("respuesta invalida al consultar documento_solicitud %d", req.DocumentoSolicitudId)
	}

	estadoNuevoId, err := getIdByCodigoAbreviacion(baseCrud, "estado_documento", req.EstadoDocumentoCodigo)
	if err != nil {
		return models.ActualizarEstadoDocumentoSolicitudResponse{}, fmt.Errorf("no se pudo resolver EstadoDocumentoCodigo=%s: %v", req.EstadoDocumentoCodigo, err)
	}

	estadoAnteriorId := 0
	if estadoObj, ok := obj["EstadoDocumentoId"].(map[string]interface{}); ok {
		estadoAnteriorId, _ = strconv.Atoi(fmt.Sprintf("%v", estadoObj["Id"]))
	} else if obj["EstadoDocumentoId"] != nil {
		estadoAnteriorId, _ = strconv.Atoi(fmt.Sprintf("%v", obj["EstadoDocumentoId"]))
	}

	obj["EstadoDocumentoId"] = map[string]interface{}{"Id": estadoNuevoId}

	var putResp map[string]interface{}
	if err := request.SendJson(getURL, "PUT", &putResp, obj); err != nil {
		return models.ActualizarEstadoDocumentoSolicitudResponse{}, fmt.Errorf("error actualizando estado del documento_solicitud %d: %v", req.DocumentoSolicitudId, err)
	}

	return models.ActualizarEstadoDocumentoSolicitudResponse{
		DocumentoSolicitudId:      req.DocumentoSolicitudId,
		EstadoDocumentoAnteriorId: estadoAnteriorId,
		EstadoDocumentoNuevoId:    estadoNuevoId,
		Mensaje:                   "Estado del documento actualizado correctamente",
		CrudResponse:              putResp,
	}, nil
}
