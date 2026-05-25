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

// ObtenerDetalleComision retorna el detalle completo de una comision para el panel de gestion (fase 2).
// Combina datos de: solicitud, detalle_solicitud (formulario FR-010) e historico_estado_comision.
func ObtenerDetalleComision(comisionId int) (models.DetalleComisionSeguimiento, error) {
	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	if baseCrud == "" {
		return models.DetalleComisionSeguimiento{}, fmt.Errorf("no esta configurado UrlComisionesCrud")
	}

	// 1. Buscar la solicitud asociada a esta comision
	solicitudObj, err := buscarSolicitudPorComisionId(baseCrud, comisionId)
	if err != nil || solicitudObj == nil {
		return models.DetalleComisionSeguimiento{}, fmt.Errorf("solicitud no encontrada para comision %d: %v", comisionId, err)
	}

	solicitudId, _ := parseInt(fmt.Sprintf("%v", solicitudObj["Id"]))
	fechaSolicitud := bandejaStr(solicitudObj["FechaCreacion"])

	// Fechas de comision vienen dentro del objeto ComisionId de la solicitud
	fechaInicio := ""
	fechaFin := ""
	if comisionObj, ok := solicitudObj["ComisionId"].(map[string]interface{}); ok {
		fechaInicio = bandejaStr(comisionObj["FechaInicio"])
		fechaFin = bandejaStr(comisionObj["FechaFinal"])
	}

	// 2. Obtener detalle_solicitud y parsear el formulario FR-010
	var docente, idDocente, correoDocente, facultad, programa, tipoEstudio, universidad, pais, ciudad, duracion string
	if solicitudId > 0 {
		detalleResp, errDetalle := obtenerDetalleSolicitud(baseCrud, solicitudId)
		if errDetalle != nil {
			logs.Warning("[detalle_comision] no se pudo obtener detalle_solicitud para solicitud %d: %v", solicitudId, errDetalle)
		} else {
			formulario, formularioErr := helpers.ObtenerDatosFormulario(detalleResp)
			if formularioErr != nil {
				logs.Warning("[detalle_comision] no se pudo parsear formulario de solicitud %d: %v", solicitudId, formularioErr)
			} else {
				docente = strings.TrimSpace(formulario.Solicitante.Q3NombresApellidos)
				idDocente = extraerSoloNumero(strings.TrimSpace(formulario.Solicitante.Q4DocumentoIdentificacion))
				correoDocente = strings.TrimSpace(formulario.Solicitante.Q6Correo)
				facultad = strings.TrimSpace(formulario.Solicitante.Q2Facultad)
				programa = strings.TrimSpace(formulario.Solicitud.Q14NombrePrograma)
				tipoEstudio = extraerTipoEstudio(formulario.Solicitud.Q13TipoEstudio)
				universidad = strings.TrimSpace(formulario.Solicitud.Q16Universidad)
				pais = strings.TrimSpace(formulario.Solicitud.Q17Pais)
				ciudad = strings.TrimSpace(formulario.Solicitud.Q18Ciudad)
				duracion = strings.TrimSpace(formulario.Solicitud.Q20NumSemestres)
			}
		}
	}

	// 3. Obtener estado actual del historico_estado_comision activo
	estadoComision := obtenerEstadoComisionActivo(baseCrud, comisionId)

	radicado := fmt.Sprintf("SOL-%d", solicitudId)

	return models.DetalleComisionSeguimiento{
		ComisionId:         comisionId,
		SolicitudId:        solicitudId,
		Radicado:           radicado,
		Docente:            docente,
		IdDocente:          idDocente,
		CorreoDocente:      correoDocente,
		Facultad:           facultad,
		Programa:           programa,
		TipoEstudio:        tipoEstudio,
		UniversidadDestino: universidad,
		PaisDestino:        pais,
		CiudadDestino:      ciudad,
		Duracion:           duracion,
		FechaSolicitud:     fechaSolicitud,
		FechaInicio:        fechaInicio,
		FechaFin:           fechaFin,
		EstadoComision:     estadoComision,
	}, nil
}

// obtenerEstadoComisionActivo consulta el historico_estado_comision activo y retorna
// el CodigoAbreviacion del estado actual. Retorna cadena vacia si no se encuentra.
func obtenerEstadoComisionActivo(baseCrud string, comisionId int) string {
	u, err := url.Parse(helpers.JoinURL(baseCrud, "/historico_estado_comision"))
	if err != nil {
		return ""
	}

	q := u.Query()
	q.Set("query", fmt.Sprintf("ComisionId.Id:%d,Activo:true", comisionId))
	q.Set("limit", "1")
	u.RawQuery = q.Encode()

	var resp map[string]interface{}
	if err := request.GetJson(u.String(), &resp); err != nil {
		logs.Warning("[detalle_comision] error consultando historico_estado_comision para comision %d: %v", comisionId, err)
		return ""
	}

	raw, ok := resp["Data"].([]interface{})
	if !ok || len(raw) == 0 {
		return ""
	}

	row, ok := raw[0].(map[string]interface{})
	if !ok {
		return ""
	}

	estadoObj, ok := row["EstadoComisionId"].(map[string]interface{})
	if !ok {
		return ""
	}

	return bandejaStr(estadoObj["CodigoAbreviacion"])
}

// extraerSoloNumero extrae el numero del final de un string como "CC 51653275" → "51653275".
// Si no hay espacios, retorna el string tal cual.
func extraerSoloNumero(s string) string {
	partes := strings.Fields(s)
	if len(partes) == 0 {
		return s
	}
	return partes[len(partes)-1]
}

// extraerTipoEstudio convierte el campo q13_tipo_estudio (interface{}) a string.
// Puede ser un string, un map con "label" o simplemente se imprime como texto.
func extraerTipoEstudio(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case map[string]interface{}:
		if label, ok := t["label"].(string); ok && label != "" {
			return strings.TrimSpace(label)
		}
		if nombre, ok := t["nombre"].(string); ok && nombre != "" {
			return strings.TrimSpace(nombre)
		}
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "<nil>" || s == "null" {
		return ""
	}
	return s
}

// parseInt convierte un string a int, retorna 0 en caso de error.
func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
