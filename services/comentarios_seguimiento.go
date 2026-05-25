package services

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/utils_oas/request"
)

// descripcionComentario es el JSON almacenado en seguimiento.Descripcion.
type descripcionComentario struct {
	Rol                  string `json:"rol"`
	Nombre               string `json:"nombre"`
	NumeroIdentificacion string `json:"numero_identificacion"`
	Texto                string `json:"texto"`
}

// getHistoricoActivoComision retorna el Id del historico_estado_comision activo para una comision.
func getHistoricoActivoComision(baseCrud string, comisionId int) (int, error) {
	u, err := url.Parse(helpers.JoinURL(baseCrud, "/historico_estado_comision"))
	if err != nil {
		return 0, err
	}

	q := u.Query()
	q.Set("query", fmt.Sprintf("ComisionId.Id:%d,Activo:true", comisionId))
	q.Set("sortby", "FechaCreacion")
	q.Set("order", "desc")
	q.Set("limit", "1")
	u.RawQuery = q.Encode()

	var envelope map[string]interface{}
	if err := request.GetJson(u.String(), &envelope); err != nil {
		return 0, fmt.Errorf("error consultando historico de comision %d: %v", comisionId, err)
	}

	raw, ok := envelope["Data"].([]interface{})
	if !ok || len(raw) == 0 {
		return 0, fmt.Errorf("no se encontro historico activo para la comision %d", comisionId)
	}

	row, ok := raw[0].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("respuesta invalida: Data[0] no es objeto")
	}

	idStr := fmt.Sprintf("%v", row["Id"])
	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil || id <= 0 {
		return 0, fmt.Errorf("Id invalido en historico de comision: %s", idStr)
	}

	return id, nil
}

// ObtenerComentariosSeguimiento retorna los comentarios de un panel especifico de una comision.
func ObtenerComentariosSeguimiento(comisionId int, codigoTipo string) ([]models.ComentarioSeguimiento, error) {
	codigoTipo = strings.TrimSpace(codigoTipo)
	if comisionId <= 0 {
		return nil, fmt.Errorf("comision_id es obligatorio")
	}
	if codigoTipo == "" {
		return nil, fmt.Errorf("codigo_tipo_seguimiento es obligatorio")
	}

	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	if baseCrud == "" {
		return nil, fmt.Errorf("no esta configurado UrlComisionesCrud")
	}

	tipoId, err := GetIdByCodigoAbreviacion(baseCrud, "tipo_seguimiento", codigoTipo)
	if err != nil {
		return nil, fmt.Errorf("tipo_seguimiento '%s' no encontrado: %v", codigoTipo, err)
	}

	historicoId, err := getHistoricoActivoComision(baseCrud, comisionId)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(helpers.JoinURL(baseCrud, "/seguimiento"))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("query", fmt.Sprintf("HistoricoEstadoComisionId.Id:%d,TipoSeguimientoId.Id:%d,Activo:true", historicoId, tipoId))
	q.Set("sortby", "FechaCreacion")
	q.Set("order", "asc")
	q.Set("limit", "0")
	u.RawQuery = q.Encode()

	logs.Info("[ComentariosSeguimiento] GET %s", u.String())

	var envelope map[string]interface{}
	if err := request.GetJson(u.String(), &envelope); err != nil {
		return nil, fmt.Errorf("error consultando seguimientos: %v", err)
	}

	rawData, _ := envelope["Data"].([]interface{})
	resultado := make([]models.ComentarioSeguimiento, 0, len(rawData))

	for _, item := range rawData {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		idFloat, _ := row["Id"].(float64)
		fechaCreacion := fmt.Sprintf("%v", row["FechaCreacion"])

		var desc descripcionComentario
		if descStr, ok := row["Descripcion"].(string); ok && descStr != "" {
			if err := json.Unmarshal([]byte(descStr), &desc); err != nil {
				logs.Warning("[ComentariosSeguimiento] no se pudo parsear Descripcion del seguimiento id=%v: %v", row["Id"], err)
			}
		}

		resultado = append(resultado, models.ComentarioSeguimiento{
			Id:            int(idFloat),
			Rol:           desc.Rol,
			Texto:         desc.Texto,
			FechaCreacion: fechaCreacion,
		})
	}

	return resultado, nil
}

// CrearComentarioSeguimiento inserta un comentario en el panel indicado de una comision.
// Retorna el Id del registro creado en seguimiento.
func CrearComentarioSeguimiento(req models.CrearComentarioRequest) (int, error) {
	if req.ComisionId <= 0 {
		return 0, fmt.Errorf("comision_id es obligatorio")
	}
	if strings.TrimSpace(req.CodigoTipoSeguimiento) == "" {
		return 0, fmt.Errorf("codigo_tipo_seguimiento es obligatorio")
	}
	if strings.TrimSpace(req.Texto) == "" {
		return 0, fmt.Errorf("texto es obligatorio")
	}
	if strings.TrimSpace(req.Rol) == "" {
		return 0, fmt.Errorf("rol es obligatorio")
	}

	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	if baseCrud == "" {
		return 0, fmt.Errorf("no esta configurado UrlComisionesCrud")
	}

	tipoId, err := GetIdByCodigoAbreviacion(baseCrud, "tipo_seguimiento", req.CodigoTipoSeguimiento)
	if err != nil {
		return 0, fmt.Errorf("tipo_seguimiento '%s' no encontrado: %v", req.CodigoTipoSeguimiento, err)
	}

	historicoId, err := getHistoricoActivoComision(baseCrud, req.ComisionId)
	if err != nil {
		return 0, err
	}

	descBytes, err := json.Marshal(descripcionComentario{
		Rol:                  strings.TrimSpace(req.Rol),
		Nombre:               strings.TrimSpace(req.Nombre),
		NumeroIdentificacion: strings.TrimSpace(req.NumeroIdentificacion),
		Texto:                strings.TrimSpace(req.Texto),
	})
	if err != nil {
		return 0, fmt.Errorf("error serializando descripcion: %v", err)
	}

	payload := map[string]interface{}{
		"HistoricoEstadoComisionId": map[string]interface{}{"Id": historicoId},
		"TipoSeguimientoId":         map[string]interface{}{"Id": tipoId},
		"Descripcion":               string(descBytes),
		"Activo":                    true,
	}

	postURL := helpers.JoinURL(baseCrud, "/seguimiento")
	if err := helpers.ValidateAbsoluteURL(postURL); err != nil {
		return 0, err
	}

	logs.Info("[ComentariosSeguimiento] POST %s payload=%+v", postURL, payload)

	var postResp map[string]interface{}
	if err := request.SendJson(postURL, "POST", &postResp, payload); err != nil {
		return 0, fmt.Errorf("error creando comentario: %v", err)
	}

	id := helpers.ExtractIdAtoi(postResp)
	if id <= 0 {
		return 0, fmt.Errorf("comentario creado pero no se pudo extraer su Id")
	}

	return id, nil
}
