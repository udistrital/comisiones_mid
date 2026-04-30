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

// ObtenerBandejaSecretariaGeneral retorna todas las comisiones activas con su estado actual.
// Es el endpoint para la secretaria general/academica que ve todo el universo de comisiones.
func ObtenerBandejaSecretariaGeneral() ([]models.ComisionBandeja, error) {
	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	if baseCrud == "" {
		return nil, fmt.Errorf("no esta configurado UrlComisionesCrud")
	}
	return obtenerTodasLasComisiones(baseCrud)
}

// ObtenerBandejaDocente retorna las comisiones del docente identificado por cedula.
// Filtra comparando la cedula contra el campo id_docente de cada comision (formato "CC 51653275"),
// lo cual es mas robusto que comparar TerceroId enteros entre sistemas distintos.
func ObtenerBandejaDocente(cedula string) ([]models.ComisionBandeja, error) {
	cedula = strings.TrimSpace(cedula)
	if cedula == "" {
		return nil, fmt.Errorf("cedula es obligatoria")
	}

	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	if baseCrud == "" {
		return nil, fmt.Errorf("no esta configurado UrlComisionesCrud")
	}

	todas, err := obtenerTodasLasComisiones(baseCrud)
	if err != nil {
		return nil, err
	}

	resultado := make([]models.ComisionBandeja, 0)
	for _, c := range todas {
		// id_docente viene como "CC 51653275" — se verifica que el ultimo token coincida
		partes := strings.Fields(c.IdDocente)
		if len(partes) > 0 && partes[len(partes)-1] == cedula {
			resultado = append(resultado, c)
		}
	}
	return resultado, nil
}

// ObtenerBandejaDecano retorna las comisiones correspondientes a las facultades del decano.
func ObtenerBandejaDecano(cedula string) ([]models.ComisionBandeja, error) {
	if strings.TrimSpace(cedula) == "" {
		return nil, fmt.Errorf("cedula es obligatoria")
	}

	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	if baseCrud == "" {
		return nil, fmt.Errorf("no esta configurado UrlComisionesCrud")
	}

	urlJBPM := strings.TrimSpace(beego.AppConfig.String("UrlJBPM"))
	if urlJBPM == "" {
		return nil, fmt.Errorf("no esta configurado UrlJBPM")
	}

	urlDecano := strings.TrimRight(urlJBPM, "/") + "/decano/"
	facultades, err := obtenerFacultadesDecano(urlDecano, cedula)
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener facultades del decano: %v", err)
	}

	todas, err := obtenerTodasLasComisiones(baseCrud)
	if err != nil {
		return nil, err
	}

	resultado := make([]models.ComisionBandeja, 0)
	for _, c := range todas {
		if c.Facultad != "" && contieneProyecto(facultades, c.Facultad) {
			resultado = append(resultado, c)
		}
	}
	return resultado, nil
}

// obtenerTodasLasComisiones consulta historico_estado_comision con Activo:true y construye
// la lista de ComisionBandeja. Para cada comision activa busca la solicitud asociada y el
// formulario del docente. Si algun dato no existe aun (ej. fecha_inicio en comision),
// devuelve cadena vacia en lugar de fallar.
func obtenerTodasLasComisiones(baseCrud string) ([]models.ComisionBandeja, error) {
	u, err := url.Parse(helpers.JoinURL(baseCrud, "/historico_estado_comision"))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("query", "Activo:true")
	q.Set("limit", "0")
	u.RawQuery = q.Encode()

	logs.Info("[bandeja_comisiones] consultando historicos activos de comision: %s", u.String())

	var resp map[string]interface{}
	if err := request.GetJson(u.String(), &resp); err != nil {
		return nil, fmt.Errorf("error consultando historicos de comision: %v", err)
	}

	raw, ok := resp["Data"].([]interface{})
	if !ok || len(raw) == 0 {
		return []models.ComisionBandeja{}, nil
	}

	resultado := make([]models.ComisionBandeja, 0, len(raw))
	vistos := make(map[int]bool)

	for _, item := range raw {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		comisionObj, ok := row["ComisionId"].(map[string]interface{})
		if !ok {
			continue
		}

		comisionId, err := strconv.Atoi(fmt.Sprintf("%v", comisionObj["Id"]))
		if err != nil || comisionId <= 0 {
			continue
		}

		// Un solo historico activo por comision (invariante de negocio).
		// Si por algun motivo hay duplicados, se toma el primero.
		if vistos[comisionId] {
			continue
		}
		vistos[comisionId] = true

		// Estado actual de la comision
		estadoStr := ""
		if estadoObj, ok := row["EstadoComisionId"].(map[string]interface{}); ok {
			estadoStr = mapEstadoComision(bandejaStr(estadoObj["CodigoAbreviacion"]))
		}

		// Fechas de comision — pueden ser nulas si la tabla aun no fue poblada
		fechaInicio := formatearFecha(comisionObj["FechaInicio"])
		fechaFin := formatearFecha(comisionObj["FechaFinal"])

		// Buscar la solicitud que origino esta comision
		solicitudObj, err := buscarSolicitudPorComisionId(baseCrud, comisionId)
		if err != nil || solicitudObj == nil {
			logs.Warning("[bandeja_comisiones] solicitud no encontrada para comision %d: %v", comisionId, err)
			resultado = append(resultado, models.ComisionBandeja{
				ComisionId:     comisionId,
				FechaInicio:    fechaInicio,
				FechaFin:       fechaFin,
				EstadoComision: estadoStr,
			})
			continue
		}

		solicitudId, _ := strconv.Atoi(fmt.Sprintf("%v", solicitudObj["Id"]))
		terceroId, _ := strconv.Atoi(fmt.Sprintf("%v", solicitudObj["TerceroId"]))
		fechaSolicitud := formatearFecha(solicitudObj["FechaCreacion"])

		// Extraer datos del docente desde el formulario
		var docente, idDocente, programa, facultad string
		if solicitudId > 0 {
			detalleResp, err := obtenerDetalleSolicitud(baseCrud, solicitudId)
			if err == nil {
				formulario, formularioErr := helpers.ObtenerDatosFormulario(detalleResp)
				if formularioErr == nil {
					docente = strings.TrimSpace(formulario.Solicitante.Q3NombresApellidos)
					idDocente = strings.TrimSpace(formulario.Solicitante.Q4DocumentoIdentificacion)
					programa = strings.TrimSpace(formulario.Solicitante.Q7Proyecto)
					facultad = strings.TrimSpace(formulario.Solicitante.Q2Facultad)
				}
			}
		}

		resultado = append(resultado, models.ComisionBandeja{
			ComisionId:     comisionId,
			SolicitudId:    solicitudId,
			TerceroId:      terceroId,
			Docente:        docente,
			IdDocente:      idDocente,
			Programa:       programa,
			Facultad:       facultad,
			FechaSolicitud: fechaSolicitud,
			FechaInicio:    fechaInicio,
			FechaFin:       fechaFin,
			EstadoComision: estadoStr,
		})
	}

	return resultado, nil
}

// buscarSolicitudPorComisionId realiza la busqueda inversa: dado un comision_id, encuentra
// la solicitud que tiene ese ComisionId asociado.
func buscarSolicitudPorComisionId(baseCrud string, comisionId int) (map[string]interface{}, error) {
	u, err := url.Parse(helpers.JoinURL(baseCrud, "/solicitud"))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("query", fmt.Sprintf("ComisionId.Id:%d,Activo:true", comisionId))
	q.Set("limit", "1")
	u.RawQuery = q.Encode()

	var resp map[string]interface{}
	if err := request.GetJson(u.String(), &resp); err != nil {
		return nil, err
	}

	raw, ok := resp["Data"].([]interface{})
	if !ok || len(raw) == 0 {
		return nil, nil
	}

	row, ok := raw[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Data[0] no es un objeto valido")
	}
	return row, nil
}

// bandejaStr extrae un string seguro de cualquier valor de interfaz.
// Retorna cadena vacia para nulos, "null", "<nil>" o timestamps zero de Go/Postgres.
func bandejaStr(v interface{}) string {
	if v == nil {
		return ""
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	switch s {
	case "", "<nil>", "null", "0001-01-01T00:00:00Z", "0001-01-01 00:00:00 +0000 UTC":
		return ""
	}
	return s
}

// formatearFecha normaliza un timestamp a YYYY-MM-DD.
// Soporta "2026-03-20 09:13:44..." (Go) y "2026-01-15T00:00:00Z" (ISO 8601).
func formatearFecha(v interface{}) string {
	s := bandejaStr(v)
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

// mapEstadoComision convierte el CodigoAbreviacion del CRUD al valor EstadoComision
// que espera el front de seguimiento (models/estados.model.ts).
// Actualizar esta tabla cuando se agreguen nuevos estados al CRUD.
func mapEstadoComision(codigo string) string {
	switch codigo {
	case "COM_INI":
		return "EN_EJECUCION"
	case "COM_FIN":
		return "FINALIZADA"
	case "COM_CAN":
		return "CANCELADA"
	case "COM_PRR_SOL":
		return "PRORROGA_SOLICITADA"
	case "COM_PRR_APR":
		return "PRORROGA_APROBADA"
	case "COM_REV":
		return "EN_REVISION"
	case "COM_INC":
		return "INCUMPLIDA"
	default:
		return "PENDIENTE"
	}
}
