package services

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/utils_oas/request"
)

// esCodCumplimiento informa si un codigo_abreviacion corresponde a un estado de cumplimiento.
func esCodCumplimiento(codigo string) bool {
	return strings.HasPrefix(codigo, "CUMP_") || strings.HasPrefix(codigo, "INCUMP_")
}

// ObtenerEstadosCumplimiento retorna los estados de estado_comision cuyo codigo
// empieza por CUMP_ o INCUMP_. Son los que se ofrecen en el dropdown del decano.
func ObtenerEstadosCumplimiento() ([]models.EstadoCumplimientoItem, error) {
	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	if baseCrud == "" {
		return nil, fmt.Errorf("no está configurado UrlComisionesCrud")
	}

	u, err := url.Parse(helpers.JoinURL(baseCrud, "/estado_comision"))
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("query", "Activo:true")
	q.Set("limit", "0")
	u.RawQuery = q.Encode()

	var envelope map[string]interface{}
	if err := request.GetJson(u.String(), &envelope); err != nil {
		return nil, fmt.Errorf("error consultando estado_comision: %v", err)
	}

	raw, _ := envelope["Data"].([]interface{})
	result := make([]models.EstadoCumplimientoItem, 0)
	for _, item := range raw {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		codigo := fmt.Sprintf("%v", row["CodigoAbreviacion"])
		if !esCodCumplimiento(codigo) {
			continue
		}
		idFloat, _ := row["Id"].(float64)
		nombre := fmt.Sprintf("%v", row["Nombre"])
		result = append(result, models.EstadoCumplimientoItem{
			Id:     int(idFloat),
			Nombre: nombre,
			Codigo: codigo,
		})
	}
	return result, nil
}

// ObtenerHistorialCumplimiento retorna todos los historico_estado_comision de la
// comision cuyo estado es de tipo cumplimiento, ordenados DESC (mas reciente primero).
func ObtenerHistorialCumplimiento(comisionId int) ([]models.RegistroCumplimientoItem, error) {
	if comisionId <= 0 {
		return nil, fmt.Errorf("comision_id es obligatorio")
	}

	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	if baseCrud == "" {
		return nil, fmt.Errorf("no está configurado UrlComisionesCrud")
	}

	u, err := url.Parse(helpers.JoinURL(baseCrud, "/historico_estado_comision"))
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("query", fmt.Sprintf("ComisionId.Id:%d", comisionId))
	q.Set("sortby", "FechaCreacion")
	q.Set("order", "desc")
	q.Set("limit", "0")
	u.RawQuery = q.Encode()

	logs.Info("[Cumplimiento] GET historico_estado_comision %s", u.String())

	var envelope map[string]interface{}
	if err := request.GetJson(u.String(), &envelope); err != nil {
		return nil, fmt.Errorf("error consultando historico_estado_comision: %v", err)
	}

	raw, _ := envelope["Data"].([]interface{})
	result := make([]models.RegistroCumplimientoItem, 0)

	for _, item := range raw {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		estadoObj, _ := row["EstadoComisionId"].(map[string]interface{})
		estadoCodigo := fmt.Sprintf("%v", estadoObj["CodigoAbreviacion"])
		if !esCodCumplimiento(estadoCodigo) {
			continue
		}

		idFloat, _ := row["Id"].(float64)
		activo, _ := row["Activo"].(bool)
		descripcion := fmt.Sprintf("%v", row["Descripcion"])
		if descripcion == "<nil>" {
			descripcion = ""
		}
		fechaCreacion := fmt.Sprintf("%v", row["FechaCreacion"])

		estadoIdFloat, _ := estadoObj["Id"].(float64)
		estadoNombre := fmt.Sprintf("%v", estadoObj["Nombre"])

		result = append(result, models.RegistroCumplimientoItem{
			Id:            int(idFloat),
			Descripcion:   descripcion,
			FechaCreacion: fechaCreacion,
			EstadoId:      int(estadoIdFloat),
			EstadoCodigo:  estadoCodigo,
			EstadoNombre:  estadoNombre,
			Activo:        activo,
		})
	}

	return result, nil
}

// CrearRegistroCumplimiento inserta un nuevo historico_estado_comision con el estado
// de cumplimiento seleccionado por el decano, y desactiva el historico anterior activo.
func CrearRegistroCumplimiento(req models.CrearRegistroCumplimientoRequest) (int, error) {
	if req.ComisionId <= 0 {
		return 0, fmt.Errorf("comision_id es obligatorio")
	}
	if req.EstadoId <= 0 {
		return 0, fmt.Errorf("estado_id es obligatorio")
	}
	if strings.TrimSpace(req.Descripcion) == "" {
		return 0, fmt.Errorf("descripcion es obligatoria")
	}
	if strings.TrimSpace(req.Rol) == "" {
		return 0, fmt.Errorf("rol es obligatorio")
	}

	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	if baseCrud == "" {
		return 0, fmt.Errorf("no está configurado UrlComisionesCrud")
	}

	// Validar que estado_id corresponde a un estado de cumplimiento.
	estadoCodigo, err := validarEstadoCumplimiento(baseCrud, req.EstadoId)
	if err != nil {
		return 0, err
	}
	logs.Info("[Cumplimiento] estado validado: id=%d codigo=%s", req.EstadoId, estadoCodigo)

	// Desactivar el historico activo actual de la comision.
	if err := desactivarHistoricoActivo(baseCrud, req.ComisionId); err != nil {
		return 0, err
	}

	// Insertar nuevo historico con el estado de cumplimiento seleccionado.
	postURL := helpers.JoinURL(baseCrud, "/historico_estado_comision")
	payload := map[string]interface{}{
		"ComisionId":       map[string]interface{}{"Id": req.ComisionId},
		"EstadoComisionId": map[string]interface{}{"Id": req.EstadoId},
		"RolUsuario":       strings.TrimSpace(req.Rol),
		"Descripcion":      strings.TrimSpace(req.Descripcion),
		"Activo":           true,
	}

	logs.Info("[Cumplimiento] POST historico_estado_comision payload=%+v", payload)

	var postResp map[string]interface{}
	if err := request.SendJson(postURL, "POST", &postResp, payload); err != nil {
		return 0, fmt.Errorf("error creando historico_estado_comision: %v", err)
	}

	id := helpers.ExtractIdAtoi(postResp)
	if id <= 0 {
		return 0, fmt.Errorf("registro creado pero no se pudo extraer su Id")
	}

	return id, nil
}

// validarEstadoCumplimiento confirma que el estado_id dado corresponde a un estado
// con prefijo CUMP_ o INCUMP_ y retorna su codigo_abreviacion.
func validarEstadoCumplimiento(baseCrud string, estadoId int) (string, error) {
	u := helpers.JoinURL(baseCrud, fmt.Sprintf("/estado_comision/%d", estadoId))
	var resp map[string]interface{}
	if err := request.GetJson(u, &resp); err != nil {
		return "", fmt.Errorf("error consultando estado_comision %d: %v", estadoId, err)
	}
	data := helpers.UnwrapDataToMap(resp)
	if data == nil {
		return "", fmt.Errorf("estado_comision %d no encontrado", estadoId)
	}
	codigo := fmt.Sprintf("%v", data["CodigoAbreviacion"])
	if !esCodCumplimiento(codigo) {
		return "", fmt.Errorf("estado_id %d (%s) no es un estado de cumplimiento", estadoId, codigo)
	}
	return codigo, nil
}

// desactivarHistoricoActivo busca el historico_estado_comision activo de la comision
// y lo marca como activo=false.
func desactivarHistoricoActivo(baseCrud string, comisionId int) error {
	historicoId, err := getHistoricoActivoComision(baseCrud, comisionId)
	if err != nil {
		// Si no hay historico activo, no es un error bloqueante — continuamos.
		logs.Warning("[Cumplimiento] no se encontro historico activo para comision %d: %v", comisionId, err)
		return nil
	}

	getURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/historico_estado_comision/%d", historicoId))
	var getResp map[string]interface{}
	if err := request.GetJson(getURL, &getResp); err != nil {
		return fmt.Errorf("error obteniendo historico %d: %v", historicoId, err)
	}

	data := helpers.UnwrapDataToMap(getResp)
	if data == nil {
		return fmt.Errorf("historico %d no encontrado para desactivar", historicoId)
	}

	data["Activo"] = false

	putURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/historico_estado_comision/%d", historicoId))
	var putResp map[string]interface{}
	if err := request.SendJson(putURL, "PUT", &putResp, data); err != nil {
		return fmt.Errorf("error desactivando historico %d: %v", historicoId, err)
	}

	logs.Info("[Cumplimiento] historico %d desactivado", historicoId)
	return nil
}
