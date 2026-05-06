package services

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/utils_oas/request"
)

func CrearComision(baseCrud string, solicitudId int, terceroId int, rolUsuario string) (int, error) {
	if solicitudId <= 0 {
		return 0, fmt.Errorf("solicitudId es obligatorio")
	}

	// 1. Consultar la solicitud por id
	getSolicitudURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/solicitud/%d", solicitudId))
	if err := helpers.ValidateAbsoluteURL(getSolicitudURL); err != nil {
		return 0, err
	}

	var solicitudResp map[string]interface{}
	if err := request.GetJson(getSolicitudURL, &solicitudResp); err != nil {
		return 0, fmt.Errorf("error consultando solicitud %d: %v", solicitudId, err)
	}

	solicitudObj := helpers.UnwrapDataToMap(solicitudResp)
	if solicitudObj == nil {
		return 0, fmt.Errorf("respuesta inválida al consultar solicitud %d", solicitudId)
	}

	// 2. Si ya existe comisión asociada, no crear otra y continuar normal
	comisionExistenteId := ExtraerComisionIdDesdeSolicitud(solicitudObj)
	if comisionExistenteId > 0 {
		return comisionExistenteId, nil
	}

	// 3. Crear la comisión
	postComisionURL := helpers.JoinURL(baseCrud, "/comision")
	if err := helpers.ValidateAbsoluteURL(postComisionURL); err != nil {
		return 0, err
	}

	payloadComision := map[string]interface{}{
		"Descripcion": fmt.Sprintf("Comisión generada automáticamente desde la solicitud %d", solicitudId),
		"Activo":      true,
	}

	var postComisionResp map[string]interface{}
	if err := request.SendJson(postComisionURL, "POST", &postComisionResp, payloadComision); err != nil {
		return 0, fmt.Errorf("error creando comisión: %v", err)
	}

	comisionId := helpers.ExtractIdAtoi(postComisionResp)
	if comisionId <= 0 {
		return 0, fmt.Errorf("se creó la comisión pero no se pudo extraer su Id")
	}

	// 4. Obtener la fecha de creación real de la solicitud usando query
	fechaCreacionSolicitud, err := GetFechaCreacionSolicitud(baseCrud, solicitudId)
	if err != nil {
		return 0, fmt.Errorf("no se pudo obtener la fecha de creación de la solicitud %d: %v", solicitudId, err)
	}

	// 5. Extraer tipo de solicitud para reenviarlo en el PUT
	tipoSolicitudId := ExtraerIdRelacion(solicitudObj["TipoSolicitudId"])
	if tipoSolicitudId <= 0 {
		return 0, fmt.Errorf("no se pudo extraer TipoSolicitudId de la solicitud %d", solicitudId)
	}

	// 6. Asociar comisión a la solicitud
	payloadSolicitudUpdate := map[string]interface{}{
		"Id":                solicitudId,
		"TerceroId":         terceroId,
		"TipoSolicitudId":   map[string]interface{}{"Id": tipoSolicitudId},
		"ComisionId":        map[string]interface{}{"Id": comisionId},
		"ObservacionCierre": fmt.Sprintf("%v", solicitudObj["ObservacionCierre"]),
		"Activo":            true,
		"FechaCreacion":     fechaCreacionSolicitud,
	}

	logs.Info("payloadSolicitudUpdate limpio=%+v", payloadSolicitudUpdate)

	var putSolicitudResp map[string]interface{}
	if err := request.SendJson(getSolicitudURL, "PUT", &putSolicitudResp, payloadSolicitudUpdate); err != nil {
		return 0, fmt.Errorf("error asociando comisión %d a la solicitud %d: %v", comisionId, solicitudId, err)
	}

	// 7. Resolver estado inicial de comisión
	estadoComisionInicialId, err := GetIdByCodigoAbreviacion(baseCrud, "estado_comision", "COM_INI")
	if err != nil {
		return comisionId, fmt.Errorf("la comisión fue creada y asociada, pero no se pudo resolver el estado inicial de comisión: %v", err)
	}

	// 8. Crear histórico inicial de comisión
	postHistoricoURL := helpers.JoinURL(baseCrud, "/historico_estado_comision")
	if err := helpers.ValidateAbsoluteURL(postHistoricoURL); err != nil {
		return comisionId, err
	}

	payloadHistorico := map[string]interface{}{
		"ComisionId":       map[string]interface{}{"Id": comisionId},
		"EstadoComisionId": map[string]interface{}{"Id": estadoComisionInicialId},
		"TerceroId":        terceroId,
		"RolUsuario":       strings.TrimSpace(rolUsuario),
		"Descripcion":      "Histórico inicial generado automáticamente al cerrar la solicitud",
		"Activo":           true,
	}

	var postHistoricoResp map[string]interface{}
	if err := request.SendJson(postHistoricoURL, "POST", &postHistoricoResp, payloadHistorico); err != nil {
		return comisionId, fmt.Errorf("la comisión fue creada y asociada, pero no se pudo crear el histórico inicial de comisión: %v", err)
	}

	historicoComisionId := helpers.ExtractIdAtoi(postHistoricoResp)
	if historicoComisionId <= 0 {
		return comisionId, fmt.Errorf("la comisión fue creada y asociada, pero no se pudo confirmar el Id del histórico de comisión")
	}

	// 9. Confirmar creación del histórico de comisión
	if err := ConfirmarHistoricoEstadoComision(baseCrud, historicoComisionId, comisionId); err != nil {
		return comisionId, fmt.Errorf("la comisión fue creada y asociada, pero no se pudo confirmar la creación del histórico de comisión: %v", err)
	}

	return comisionId, nil
}

func GetFechaCreacionSolicitud(baseCrud string, solicitudId int) (string, error) {
	u, err := url.Parse(helpers.JoinURL(baseCrud, "/solicitud"))
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("query", fmt.Sprintf("Id:%d", solicitudId))
	q.Set("limit", "1")
	u.RawQuery = q.Encode()

	var envelope map[string]interface{}
	if err := request.GetJson(u.String(), &envelope); err != nil {
		return "", fmt.Errorf("error consultando solicitud por query: %v", err)
	}

	logs.Info("respuesta solicitud por query=%+v", envelope)

	raw := envelope["Data"]
	arr, ok := raw.([]interface{})
	if !ok || len(arr) == 0 {
		return "", fmt.Errorf("no se encontró la solicitud %d", solicitudId)
	}

	row, ok := arr[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("respuesta inválida: Data[0] no es objeto")
	}

	logs.Info("row solicitud por query=%+v", row)

	candidatos := []string{
		"FechaCreacion",
		"fecha_creacion",
		"Fecha_creacion",
		"fechaCreacion",
	}

	for _, key := range candidatos {
		if v, ok := row[key]; ok && v != nil {
			fecha := strings.TrimSpace(fmt.Sprintf("%v", v))
			if fecha != "" && fecha != "<nil>" && fecha != "Z" {
				return fecha, nil
			}
		}
	}

	return "", fmt.Errorf("la solicitud %d no trae una FechaCreacion válida", solicitudId)
}

func ExtraerComisionIdDesdeSolicitud(solicitudObj map[string]interface{}) int {
	if solicitudObj == nil {
		return 0
	}

	v, ok := solicitudObj["ComisionId"]
	if !ok || v == nil {
		return 0
	}

	switch vv := v.(type) {
	case map[string]interface{}:
		if id, err := strconv.Atoi(fmt.Sprintf("%v", vv["Id"])); err == nil && id > 0 {
			return id
		}
	default:
		if id, err := strconv.Atoi(fmt.Sprintf("%v", vv)); err == nil && id > 0 {
			return id
		}
	}

	return 0
}

func ExtraerIdRelacion(v interface{}) int {
	if v == nil {
		return 0
	}

	switch vv := v.(type) {
	case map[string]interface{}:
		id, _ := strconv.Atoi(fmt.Sprintf("%v", vv["Id"]))
		return id
	default:
		id, _ := strconv.Atoi(fmt.Sprintf("%v", vv))
		return id
	}
}

func ConfirmarHistoricoEstadoComision(baseCrud string, historicoComisionId int, comisionId int) error {
	if historicoComisionId <= 0 {
		return fmt.Errorf("historicoComisionId es obligatorio")
	}

	getURL := helpers.JoinURL(baseCrud, fmt.Sprintf("/historico_estado_comision/%d", historicoComisionId))
	if err := helpers.ValidateAbsoluteURL(getURL); err != nil {
		return err
	}

	var getResp map[string]interface{}
	if err := request.GetJson(getURL, &getResp); err != nil {
		return fmt.Errorf("error consultando histórico de comisión %d: %v", historicoComisionId, err)
	}

	obj := helpers.UnwrapDataToMap(getResp)
	if obj == nil {
		return fmt.Errorf("respuesta inválida al consultar histórico de comisión %d", historicoComisionId)
	}

	idConsultado, _ := strconv.Atoi(fmt.Sprintf("%v", obj["Id"]))
	if idConsultado != historicoComisionId {
		return fmt.Errorf("el histórico consultado no coincide con el creado")
	}

	if v, ok := obj["ComisionId"]; ok && v != nil {
		switch vv := v.(type) {
		case map[string]interface{}:
			idComisionResp, _ := strconv.Atoi(fmt.Sprintf("%v", vv["Id"]))
			if idComisionResp != comisionId {
				return fmt.Errorf("el histórico creado no quedó asociado a la comisión esperada")
			}
		default:
			idComisionResp, _ := strconv.Atoi(fmt.Sprintf("%v", vv))
			if idComisionResp != comisionId {
				return fmt.Errorf("el histórico creado no quedó asociado a la comisión esperada")
			}
		}
	}

	return nil
}
