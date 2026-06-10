package models

type ResponseRechazarAprobarSolicitudCierre struct {
	SolicitudCierreId int `json:"solicitud_cierre_id"`
}

type CierreAprobacionSolicitud struct {
	SolicitudId int    `json:"SolicitudId"`
	ComisionId  int    `json:"ComisionId"`
	Observacion string `json:"Observacion"`
	TerceroId   int    `json:"TerceroId"`
	RolUsuario  string `json:"RolUsuario"`
}
