package services

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/utils_oas/request"
)

type CoordinadoresXML struct {
	XMLName       xml.Name         `xml:"coordinadores"`
	Coordinadores []CoordinadorXML `xml:"coordinador"`
}

type CoordinadorXML struct {
	NombreCoordinador string `xml:"coordinador"`
	CodigoCarrera     string `xml:"codigo_carrera"`
	Identificacion    string `xml:"identificacion"`
	NombreCarrera     string `xml:"nombre_carrera"`
}

type secretariaXML struct {
	XMLName xml.Name     `xml:"secretaria"`
	Persona []personaXML `xml:"persona"`
}

type personaXML struct {
	Apellidos         string `xml:"apellidos"`
	Estado            string `xml:"estado"`
	Identificacion    string `xml:"identificacion"`
	Dependencia       string `xml:"dependencia"`
	CodigoDependencia string `xml:"codigo_dependencia"`
	Nombres           string `xml:"nombres"`
}

func ObtenerSolicitudesPendientesCoordinador(numeroIdentificacion string) ([]models.SolicitudPendienteRevisor, error) {
	if strings.TrimSpace(numeroIdentificacion) == "" {
		return nil, fmt.Errorf("numeroIdentificacion es obligatorio")
	}

	baseCrud, err := getBaseURL("UrlComisionesCrud", "COMISIONES_MID_COMISIONES_CRUD")
	if err != nil {
		return nil, err
	}

	urlCoordinador, err := getBaseURL("UrlJBPM", "COMISONES_MID_ACADEMICA_JBPM")
	if err != nil {
		return nil, err
	}
	urlCoordinador = strings.TrimRight(urlCoordinador, "/") + "/coordinador_usuario/"

	proyectoCoordinador, err := obtenerProyectoCurricularCoordinador(urlCoordinador, numeroIdentificacion)
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener el proyecto curricular del coordinador: %v", err)
	}

	estadoPendiente := "REV_PROY"
	solicitudes, err := consultarSolicitudesPorEstado(baseCrud, estadoPendiente)
	if err != nil {
		return nil, fmt.Errorf("no se pudieron consultar las solicitudes pendientes: %v", err)
	}

	resultado := make([]models.SolicitudPendienteRevisor, 0)

	for _, solicitud := range solicitudes {
		solicitudId, err := strconv.Atoi(strings.TrimSpace(fmt.Sprintf("%v", solicitud["Id"])))
		if err != nil || solicitudId <= 0 {
			continue
		}

		detalleSolicitudResp, err := obtenerDetalleSolicitud(baseCrud, solicitudId)
		if err != nil {
			continue
		}

		datosFormulario, outputError := helpers.ObtenerDatosFormulario(detalleSolicitudResp)
		if outputError != nil {
			continue
		}

		proyectoSolicitud := strings.TrimSpace(datosFormulario.Solicitante.Q7Proyecto)
		if proyectoSolicitud == "" {
			continue
		}

		if normalizarTexto(proyectoSolicitud) != normalizarTexto(proyectoCoordinador) {
			continue
		}

		nombreDocente := strings.TrimSpace(datosFormulario.Solicitante.Q3NombresApellidos)
		documentoDocente := strings.TrimSpace(datosFormulario.Solicitante.Q4DocumentoIdentificacion)

		estadoSolicitud := ""
		if estadoObj, ok := solicitud["EstadoSolicitudId"].(map[string]interface{}); ok {
			estadoSolicitud = strings.TrimSpace(fmt.Sprintf("%v", estadoObj["Nombre"]))
			if estadoSolicitud == "" {
				estadoSolicitud = strings.TrimSpace(fmt.Sprintf("%v", estadoObj["CodigoAbreviacion"]))
			}
		}

		resultado = append(resultado, models.SolicitudPendienteRevisor{
			Id:               solicitudId,
			FechaCreacion:    strings.TrimSpace(fmt.Sprintf("%v", solicitud["FechaCreacion"])),
			NombreDocente:    nombreDocente,
			DocumentoDocente: documentoDocente,
			EstadoSolicitud:  estadoSolicitud,
		})
	}

	return resultado, nil
}

func obtenerProyectoCurricularCoordinador(baseURL, numeroIdentificacion string) (string, error) {
	urlFinal := strings.TrimRight(baseURL, "/") + "/" + strings.TrimSpace(numeroIdentificacion)

	var resp CoordinadoresXML
	if err := request.GetXml(urlFinal, &resp); err != nil {
		return "", err
	}

	if len(resp.Coordinadores) == 0 {
		return "", fmt.Errorf("no se encontró información de coordinador para la identificación %s", numeroIdentificacion)
	}

	proyecto := strings.TrimSpace(resp.Coordinadores[0].NombreCarrera)
	if proyecto == "" {
		return "", fmt.Errorf("la respuesta XML no trajo nombre_carrera")
	}

	return proyecto, nil
}

func ObtenerSolicitudesPendientesSecretaria(numeroIdentificacion string) ([]models.SolicitudPendienteRevisor, error) {
	if strings.TrimSpace(numeroIdentificacion) == "" {
		return nil, fmt.Errorf("numeroIdentificacion es obligatorio")
	}

	baseCrud, err := getBaseURL("UrlComisionesCrud", "COMISIONES_MID_COMISIONES_CRUD")
	if err != nil {
		return nil, err
	}

	urlSecretaria, err := getBaseURL("UrlJBPM", "COMISONES_MID_ACADEMICA_JBPM")
	if err != nil {
		return nil, err
	}
	urlSecretaria = strings.TrimRight(urlSecretaria, "/") + "/secretaria_academica/"

	dependenciaSecretaria, err := obtenerDependenciaSecretaria(urlSecretaria, numeroIdentificacion)
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener dependencia del secretario(a): %v", err)
	}

	estadoPendiente := "REV_SEC_ACAD"
	solicitudes, err := consultarSolicitudesPorEstado(baseCrud, estadoPendiente)
	if err != nil {
		return nil, fmt.Errorf("no se pudieron consultar las solicitudes pendientes: %v", err)
	}

	resultado := make([]models.SolicitudPendienteRevisor, 0)

	for _, solicitud := range solicitudes {
		solicitudId, err := strconv.Atoi(strings.TrimSpace(fmt.Sprintf("%v", solicitud["Id"])))
		if err != nil || solicitudId <= 0 {
			continue
		}

		detalleSolicitudResp, err := obtenerDetalleSolicitud(baseCrud, solicitudId)
		if err != nil {
			continue
		}

		datosFormulario, outputError := helpers.ObtenerDatosFormulario(detalleSolicitudResp)
		if outputError != nil {
			continue
		}

		proyectoSolicitud := strings.TrimSpace(datosFormulario.Solicitante.Q2Facultad)
		if proyectoSolicitud == "" {
			continue
		}

		if normalizarTexto(proyectoSolicitud) != normalizarTexto(dependenciaSecretaria) {
			continue
		}

		nombreDocente := strings.TrimSpace(datosFormulario.Solicitante.Q3NombresApellidos)
		documentoDocente := strings.TrimSpace(datosFormulario.Solicitante.Q4DocumentoIdentificacion)

		estadoSolicitud := ""
		if estadoObj, ok := solicitud["EstadoSolicitudId"].(map[string]interface{}); ok {
			estadoSolicitud = strings.TrimSpace(fmt.Sprintf("%v", estadoObj["Nombre"]))
			if estadoSolicitud == "" {
				estadoSolicitud = strings.TrimSpace(fmt.Sprintf("%v", estadoObj["CodigoAbreviacion"]))
			}
		}

		resultado = append(resultado, models.SolicitudPendienteRevisor{
			Id:               solicitudId,
			FechaCreacion:    strings.TrimSpace(fmt.Sprintf("%v", solicitud["FechaCreacion"])),
			NombreDocente:    nombreDocente,
			DocumentoDocente: documentoDocente,
			EstadoSolicitud:  estadoSolicitud,
		})
	}

	return resultado, nil

}

func obtenerDependenciaSecretaria(baseURL, numeroIdentificacion string) (string, error) {
	urlFinal := strings.TrimRight(baseURL, "/") + "/" + strings.TrimSpace(numeroIdentificacion)

	var resp secretariaXML

	if err := request.GetXml(urlFinal, &resp); err != nil {
		return "", err
	}

	if len(resp.Persona) == 0 {
		return "", fmt.Errorf("no se encontro información del secretario(a) para la identificación %s", numeroIdentificacion)
	}

	dependencia := strings.TrimSpace(resp.Persona[0].Dependencia)
	if dependencia == "" {
		return "", fmt.Errorf("La respuesta XML no trajo dependencia")
	}

	return dependencia, nil
}

func consultarSolicitudesPorEstado(baseCrud, codigoEstado string) ([]map[string]interface{}, error) {
	u, err := url.Parse(joinURL(baseCrud, "/historico_estado_solicitud"))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("query", fmt.Sprintf("Activo:true,EstadoSolicitudId.CodigoAbreviacion:%s", strings.TrimSpace(codigoEstado)))
	q.Set("limit", "0")
	u.RawQuery = q.Encode()

	var resp map[string]interface{}
	if err := request.GetJson(u.String(), &resp); err != nil {
		return nil, err
	}

	raw, ok := resp["Data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("respuesta inválida consultando históricos")
	}

	resultado := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		if row, ok := item.(map[string]interface{}); ok {
			solicitudObj, ok := row["SolicitudId"].(map[string]interface{})
			if ok {
				solicitudObj["EstadoSolicitudId"] = row["EstadoSolicitudId"]
				resultado = append(resultado, solicitudObj)
			}
		}
	}

	return resultado, nil
}

func obtenerDetalleSolicitud(baseCrud string, solicitudId int) (map[string]interface{}, error) {
	u, err := url.Parse(joinURL(baseCrud, "/detalle_solicitud"))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("query", fmt.Sprintf("SolicitudId:%d,Activo:true", solicitudId))
	q.Set("limit", "1")
	u.RawQuery = q.Encode()

	var resp map[string]interface{}
	if err := request.GetJson(u.String(), &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func normalizarTexto(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "á", "a")
	s = strings.ReplaceAll(s, "é", "e")
	s = strings.ReplaceAll(s, "í", "i")
	s = strings.ReplaceAll(s, "ó", "o")
	s = strings.ReplaceAll(s, "ú", "u")
	s = strings.ReplaceAll(s, "ü", "u")
	return s
}
