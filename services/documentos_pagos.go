package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/utils_oas/request"
)

const codigoPagoSoporte = "PAG_SOPORTE"

// ObtenerDocumentosPago retorna todos los documentos de tipo PAG_SOPORTE
// asociados al historico activo de la comision, con info del subidor.
func ObtenerDocumentosPago(comisionId int) ([]models.DocumentoPagoItem, error) {
	if comisionId <= 0 {
		return nil, fmt.Errorf("comision_id es obligatorio")
	}

	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	if baseCrud == "" {
		return nil, fmt.Errorf("no está configurado UrlComisionesCrud")
	}

	tipoId, err := GetIdByCodigoAbreviacion(baseCrud, "tipo_documento_comision", codigoPagoSoporte)
	if err != nil {
		return nil, fmt.Errorf("tipo_documento_comision '%s' no encontrado: %v", codigoPagoSoporte, err)
	}

	docsSubidos, err := obtenerDocumentosComisionPorComisionYTipo(baseCrud, comisionId, tipoId)
	if err != nil {
		return nil, err
	}

	result := make([]models.DocumentoPagoItem, 0, len(docsSubidos))
	for _, doc := range docsSubidos {
		item := models.DocumentoPagoItem{
			DocumentoComisionId: extractIntFromDocField(doc, "Id"),
			DocumentoId:         extractIntFromDocField(doc, "DocumentoId"),
		}
		item.Estado, item.EstadoNombre = extraerEstadoDocumentoComision(doc)

		if item.DocumentoId > 0 {
			item.Enlace, item.Nombre, item.SubidoPorRol, item.SubidoPorNombre = resolverDetalleDocumentoPago(item.DocumentoId)
		}

		result = append(result, item)
	}

	return result, nil
}

// obtenerDocumentosComisionPorComisionYTipo consulta documento_comision filtrando
// por comision (todos sus historicos) y tipo de documento.
// Esto evita que los documentos "desaparezcan" al cambiar de estado.
func obtenerDocumentosComisionPorComisionYTipo(baseCrud string, comisionId, tipoId int) ([]map[string]interface{}, error) {
	urlStr := helpers.JoinURL(baseCrud, fmt.Sprintf(
		"/documento_comision?query=HistoricoEstadoComisionId.ComisionId.Id:%d,TipoDocumentoId.Id:%d,Activo:true&limit=0",
		comisionId, tipoId,
	))

	logs.Info("[DocumentosPago] GET documento_comision %s", urlStr)

	var envelope map[string]interface{}
	if err := request.GetJson(urlStr, &envelope); err != nil {
		return nil, fmt.Errorf("error consultando documento_comision: %v", err)
	}

	raw, _ := envelope["Data"].([]interface{})
	result := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		if row, ok := item.(map[string]interface{}); ok {
			result = append(result, row)
		}
	}
	return result, nil
}

// resolverDetalleDocumentoPago consulta documento/{id} y extrae enlace, nombre
// y el JSON de descripcion con subido_por_rol y subido_por_nombre.
func resolverDetalleDocumentoPago(docId int) (enlace, nombre, subidoPorRol, subidoPorNombre string) {
	baseDoc := strings.TrimSpace(beego.AppConfig.String("UrlDocumentos"))
	if baseDoc == "" || docId <= 0 {
		return
	}

	var detalle map[string]interface{}
	if err := request.GetJson(helpers.JoinURL(baseDoc, fmt.Sprintf("documento/%d", docId)), &detalle); err != nil {
		logs.Warning("[DocumentosPago] no se pudo obtener detalle para documento_id=%d: %v", docId, err)
		return
	}

	enlace, _ = detalle["Enlace"].(string)
	nombre, _ = detalle["Nombre"].(string)

	descripcionStr, _ := detalle["Descripcion"].(string)
	if descripcionStr != "" {
		var meta struct {
			Rol    string `json:"rol"`
			Nombre string `json:"nombre"`
		}
		if err := json.Unmarshal([]byte(descripcionStr), &meta); err == nil {
			subidoPorRol = meta.Rol
			subidoPorNombre = meta.Nombre
		}
	}

	return
}
