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

func ObtenerSolicitudesPendientesCoordinador(numeroIdentificacion string) ([]models.SolicitudPendienteRevisor, error) {
	if strings.TrimSpace(numeroIdentificacion) == "" {
		return nil, fmt.Errorf("numeroIdentificacion es obligatorio")
	}

	// CRUD comisiones
	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	logs.Info("UrlComisionesCrud=%q", baseCrud)
	if baseCrud == "" {
		return []models.SolicitudPendienteRevisor{}, fmt.Errorf("no esta configurado UrlComisionesCrud")
	}

	urlCoordinador := strings.TrimSpace(beego.AppConfig.String("UrlJBPM"))
	logs.Info("UrlJBPM=%q", urlCoordinador)
	if urlCoordinador == "" {
		return []models.SolicitudPendienteRevisor{}, fmt.Errorf("no esta configurado UrlJBPM")
	}
	urlCoordinador = strings.TrimRight(urlCoordinador, "/") + "/coordinador_usuario/"

	proyectosCoordinador, err := obtenerProyectosCurricularesCoordinador(urlCoordinador, numeroIdentificacion)
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener los proyectos curriculares del coordinador: %v", err)
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

		if !contieneProyecto(proyectosCoordinador, proyectoSolicitud) {
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

func obtenerProyectosCurricularesCoordinador(baseURL, numeroIdentificacion string) ([]string, error) {
	urlFinal := strings.TrimRight(baseURL, "/") + "/" + strings.TrimSpace(numeroIdentificacion)

	var resp models.CoordinadoresXML
	if err := request.GetXml(urlFinal, &resp); err != nil {
		return nil, err
	}

	if len(resp.Coordinadores) == 0 {
		return nil, fmt.Errorf("no se encontro informacion de coordinador para la identificacion %s", numeroIdentificacion)
	}

	proyectos := make([]string, 0, len(resp.Coordinadores))
	proyectosNormalizados := make(map[string]struct{})

	for _, coordinador := range resp.Coordinadores {
		proyecto := strings.TrimSpace(coordinador.NombreCarrera)
		if proyecto == "" {
			continue
		}

		proyectoNormalizado := normalizarTexto(proyecto)
		if _, exists := proyectosNormalizados[proyectoNormalizado]; exists {
			continue
		}

		proyectosNormalizados[proyectoNormalizado] = struct{}{}
		proyectos = append(proyectos, proyecto)
	}

	if len(proyectos) == 0 {
		return nil, fmt.Errorf("la respuesta XML no trajo nombre_carrera")
	}

	return proyectos, nil
}

func ObtenerSolicitudesPendientesSecretaria(numeroIdentificacion string) ([]models.SolicitudPendienteRevisor, error) {
	if strings.TrimSpace(numeroIdentificacion) == "" {
		return nil, fmt.Errorf("numeroIdentificacion es obligatorio")
	}
	// CRUD comisiones
	baseCrud := strings.TrimSpace(beego.AppConfig.String("UrlComisionesCrud"))
	logs.Info("UrlComisionesCrud=%q", baseCrud)
	if baseCrud == "" {
		return []models.SolicitudPendienteRevisor{}, fmt.Errorf("no esta configurado UrlComisionesCrud")
	}

	urlSecretaria := strings.TrimSpace(beego.AppConfig.String("UrlJBPM"))
	logs.Info("UrlJBPM=%q", urlSecretaria)
	if urlSecretaria == "" {
		return []models.SolicitudPendienteRevisor{}, fmt.Errorf("no esta configurado UrlJBPM")
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

	var resp models.SecretariaXML

	if err := request.GetXml(urlFinal, &resp); err != nil {
		return "", err
	}

	if len(resp.Persona) == 0 {
		return "", fmt.Errorf("no se encontro informacion del secretario(a) para la identificacion %s", numeroIdentificacion)
	}

	dependencia := strings.TrimSpace(resp.Persona[0].Dependencia)
	if dependencia == "" {
		return "", fmt.Errorf("La respuesta XML no trajo dependencia")
	}

	return dependencia, nil
}

func consultarSolicitudesPorEstado(baseCrud, codigoEstado string) ([]map[string]interface{}, error) {
	u, err := url.Parse(helpers.JoinURL(baseCrud, "/historico_estado_solicitud"))
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
		return nil, fmt.Errorf("respuesta invalida consultando historicos")
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
	u, err := url.Parse(helpers.JoinURL(baseCrud, "/detalle_solicitud"))
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
	s = strings.ReplaceAll(s, "?", "a")
	s = strings.ReplaceAll(s, "?", "e")
	s = strings.ReplaceAll(s, "?", "i")
	s = strings.ReplaceAll(s, "?", "o")
	s = strings.ReplaceAll(s, "?", "u")
	s = strings.ReplaceAll(s, "?", "u")
	return s
}

func contieneProyecto(proyectos []string, proyectoSolicitud string) bool {
	proyectoSolicitudNormalizado := normalizarTexto(proyectoSolicitud)
	for _, proyecto := range proyectos {
		if normalizarTexto(proyecto) == proyectoSolicitudNormalizado {
			return true
		}
	}

	return false
}
