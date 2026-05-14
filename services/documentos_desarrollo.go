package services

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/utils_oas/request"
)

// momentosOrdenados define los grupos del panel en el orden de presentacion.
var momentosOrdenados = []struct {
	Prefijo string
	Nombre  string
}{
	{"INI_", "Inicio de la comisión"},
	{"DES_", "Desarrollo de la comisión"},
	{"PROR_", "Prórroga o renovación"},
	{"CUMP_", "Cumplimiento de compromisos"},
	{"CULM_", "Culminación académica"},
	{"POST_", "Post-comisión"},
}

// ObtenerDocumentosDesarrollo retorna los tipos de documento agrupados por momento,
// cada uno con el estado del documento subido para esa comision (si existe).
func ObtenerDocumentosDesarrollo(comisionId int) ([]models.GrupoDocumentosDesarrollo, error) {
	if comisionId <= 0 {
		return nil, fmt.Errorf("comision_id es obligatorio")
	}

	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	if baseCrud == "" {
		return nil, fmt.Errorf("no está configurado UrlComisionesCrud")
	}

	historicoId, err := getHistoricoActivoComision(baseCrud, comisionId)
	if err != nil {
		return nil, err
	}

	tipos, err := obtenerTiposDocumentoComision(baseCrud)
	if err != nil {
		return nil, err
	}

	docsSubidos, err := obtenerDocumentosComisionPorHistorico(baseCrud, historicoId)
	if err != nil {
		return nil, err
	}

	// mapa tipoId -> registro documento_comision subido
	docPorTipo := map[int]map[string]interface{}{}
	for _, doc := range docsSubidos {
		tipoId := extractIntFromDocField(doc, "TipoDocumentoId")
		if tipoId > 0 {
			docPorTipo[tipoId] = doc
		}
	}

	grupos := make([]models.GrupoDocumentosDesarrollo, 0, len(momentosOrdenados))
	for _, m := range momentosOrdenados {
		grupo := models.GrupoDocumentosDesarrollo{
			Momento:    m.Nombre,
			Prefijo:    m.Prefijo,
			Documentos: []models.DocumentoDesarrolloItem{},
		}

		for _, tipo := range tipos {
			codigo := fmt.Sprintf("%v", tipo["CodigoAbreviacion"])
			if !strings.HasPrefix(codigo, m.Prefijo) {
				continue
			}

			tipoId := int(tipo["Id"].(float64))
			nombre := fmt.Sprintf("%v", tipo["Nombre"])

			item := models.DocumentoDesarrolloItem{
				TipoId: tipoId,
				Codigo: codigo,
				Nombre: nombre,
			}

			if doc, ok := docPorTipo[tipoId]; ok {
				item.DocumentoComisionId = extractIntFromDocField(doc, "Id")
				item.DocumentoId = extractIntFromDocField(doc, "DocumentoId")
				item.Estado, item.EstadoNombre = extraerEstadoDocumentoComision(doc)
				if item.DocumentoId > 0 {
					item.Enlace = resolverEnlaceDocumentoId(item.DocumentoId)
				}
			}

			grupo.Documentos = append(grupo.Documentos, item)
		}

		grupos = append(grupos, grupo)
	}

	return grupos, nil
}

// SubirDocumentoDesarrollo sube un documento al gestor documental y registra el documento_comision.
// Retorna el Id del registro documento_comision creado.
func SubirDocumentoDesarrollo(req models.SubirDocumentoDesarrolloRequest) (int, error) {
	if req.ComisionId <= 0 {
		return 0, fmt.Errorf("comision_id es obligatorio")
	}
	if strings.TrimSpace(req.TipoDocumentoCodigo) == "" {
		return 0, fmt.Errorf("tipo_documento_codigo es obligatorio")
	}
	if strings.TrimSpace(req.Nombre) == "" {
		return 0, fmt.Errorf("nombre es obligatorio")
	}
	if strings.TrimSpace(req.File) == "" {
		return 0, fmt.Errorf("file es obligatorio")
	}

	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	if baseCrud == "" {
		return 0, fmt.Errorf("no está configurado UrlComisionesCrud")
	}

	historicoId, err := getHistoricoActivoComision(baseCrud, req.ComisionId)
	if err != nil {
		return 0, err
	}

	tipoDocumentoId, err := GetIdByCodigoAbreviacion(baseCrud, "tipo_documento_comision", req.TipoDocumentoCodigo)
	if err != nil {
		return 0, fmt.Errorf("tipo_documento_comision '%s' no encontrado: %v", req.TipoDocumentoCodigo, err)
	}

	// estado_documento_comision asume mismos registros que estado_documento
	estadoId, err := GetIdByCodigoAbreviacion(baseCrud, "estado_documento", "CARG")
	if err != nil {
		return 0, fmt.Errorf("no se pudo resolver estado inicial CARG: %v", err)
	}

	resultadoDocs, outputError := helpers.CrearDocumento([]models.CrearDocumentoGestorDocumental{
		{
			IdTipoDocumento: req.IdTipoDocumento,
			Nombre:          req.Nombre,
			Descripcion:     req.Descripcion,
			Metadatos:       map[string]interface{}{},
			File:            req.File,
		},
	})
	if outputError != nil {
		return 0, fmt.Errorf("error subiendo documento al gestor documental: %v", outputError)
	}
	if len(resultadoDocs) == 0 {
		return 0, fmt.Errorf("gestor documental no retornó el documento creado")
	}

	documentoId, err := strconv.Atoi(fmt.Sprintf("%v", resultadoDocs[0]["id"]))
	if err != nil || documentoId <= 0 {
		return 0, fmt.Errorf("id de documento inválido retornado por gestor documental")
	}

	postURL := helpers.JoinURL(baseCrud, "/documento_comision")
	payload := map[string]interface{}{
		"DocumentoId":               documentoId,
		"HistoricoEstadoComisionId": map[string]interface{}{"Id": historicoId},
		"TipoDocumentoId":           map[string]interface{}{"Id": tipoDocumentoId},
		"EstadoDocumentoComisionId": map[string]interface{}{"Id": estadoId},
		"Activo":                    true,
	}

	logs.Info("[DocumentosDesarrollo] POST documento_comision payload=%+v", payload)

	var postResp map[string]interface{}
	if err := request.SendJson(postURL, "POST", &postResp, payload); err != nil {
		return 0, fmt.Errorf("error creando documento_comision: %v", err)
	}

	id := helpers.ExtractIdAtoi(postResp)
	if id <= 0 {
		return 0, fmt.Errorf("documento_comision creado pero no se pudo extraer su Id")
	}

	return id, nil
}

// DesactivarDocumentoDesarrollo hace soft delete de un documento_comision (activo=false).
func DesactivarDocumentoDesarrollo(documentoComisionId int) error {
	if documentoComisionId <= 0 {
		return fmt.Errorf("documento_comision_id es obligatorio")
	}

	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	if baseCrud == "" {
		return fmt.Errorf("no está configurado UrlComisionesCrud")
	}

	getURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/documento_comision/%d", documentoComisionId))
	var getResp map[string]interface{}
	if err := request.GetJson(getURL, &getResp); err != nil {
		return fmt.Errorf("error obteniendo documento_comision %d: %v", documentoComisionId, err)
	}

	data := helpers.UnwrapDataToMap(getResp)
	if data == nil {
		return fmt.Errorf("documento_comision %d no encontrado", documentoComisionId)
	}

	data["Activo"] = false

	putURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/documento_comision/%d", documentoComisionId))
	var putResp map[string]interface{}
	if err := request.SendJson(putURL, "PUT", &putResp, data); err != nil {
		return fmt.Errorf("error desactivando documento_comision %d: %v", documentoComisionId, err)
	}

	return nil
}

func obtenerTiposDocumentoComision(baseCrud string) ([]map[string]interface{}, error) {
	u, err := url.Parse(helpers.JoinURL(baseCrud, "/tipo_documento_comision"))
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("query", "Activo:true")
	q.Set("limit", "0")
	u.RawQuery = q.Encode()

	var envelope map[string]interface{}
	if err := request.GetJson(u.String(), &envelope); err != nil {
		return nil, fmt.Errorf("error consultando tipo_documento_comision: %v", err)
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

func obtenerDocumentosComisionPorHistorico(baseCrud string, historicoId int) ([]map[string]interface{}, error) {
	u, err := url.Parse(helpers.JoinURL(baseCrud, "/documento_comision"))
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("query", fmt.Sprintf("HistoricoEstadoComisionId.Id:%d,Activo:true", historicoId))
	q.Set("limit", "0")
	u.RawQuery = q.Encode()

	logs.Info("[DocumentosDesarrollo] GET documento_comision %s", u.String())

	var envelope map[string]interface{}
	if err := request.GetJson(u.String(), &envelope); err != nil {
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

// resolverEnlaceDocumentoId consulta UrlDocumentos/documento/{id} y retorna el Enlace.
// Retorna "" si falla o no está configurado.
func resolverEnlaceDocumentoId(docId int) string {
	baseDoc := strings.TrimSpace(beego.AppConfig.String("UrlDocumentos"))
	if baseDoc == "" || docId <= 0 {
		return ""
	}
	var detalle map[string]interface{}
	if err := request.GetJson(helpers.JoinURL(baseDoc, fmt.Sprintf("documento/%d", docId)), &detalle); err != nil {
		logs.Warning("[DocumentosDesarrollo] no se pudo obtener enlace para documento_id=%d: %v", docId, err)
		return ""
	}
	enlace, _ := detalle["Enlace"].(string)
	return enlace
}

// extractIntFromDocField extrae un int de un campo que puede ser float64 directo
// o un objeto anidado {"Id": N} (patron tipico del ORM de Beego).
func extractIntFromDocField(row map[string]interface{}, field string) int {
	val, ok := row[field]
	if !ok {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return int(v)
	case map[string]interface{}:
		if id, ok := v["Id"].(float64); ok {
			return int(id)
		}
	}
	return 0
}

// extraerEstadoDocumentoComision obtiene codigo y nombre del estado del documento.
func extraerEstadoDocumentoComision(doc map[string]interface{}) (codigo, nombre string) {
	val, ok := doc["EstadoDocumentoComisionId"]
	if !ok {
		return "", ""
	}
	if obj, ok := val.(map[string]interface{}); ok {
		codigo = fmt.Sprintf("%v", obj["CodigoAbreviacion"])
		nombre = fmt.Sprintf("%v", obj["Nombre"])
	}
	return
}
