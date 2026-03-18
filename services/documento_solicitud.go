package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
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
		if strings.TrimSpace(doc.NombreArchivo) == "" {
			return nil, nil, fmt.Errorf("Documentos[%d].NombreArchivo es obligatorio", i)
		}
		if strings.TrimSpace(doc.File) == "" {
			return nil, nil, fmt.Errorf("Documentos[%d].Base64Documento es obligatorio", i)
		}
		if doc.IdTipoDocumento <= 0 {
			return nil, nil, fmt.Errorf("Documentos[%d].IdTipoDocumento es obligatorio", i)
		}

		documentosGestor = append(documentosGestor, models.CrearDocumentoGestorDocumental{
			IdTipoDocumento: doc.IdTipoDocumento,
			Nombre:          strings.TrimSpace(doc.NombreArchivo),
			Descripcion:     strings.TrimSpace(doc.DescripcionDocumento),
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
