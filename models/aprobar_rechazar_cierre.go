package models

type ResponseRechazarSolicitudCierre struct {
	SolicitudCierreId int `json:"solicitud_cierre_id"`
}

type CierreSolicitud struct {
	SolicitudId int    `json:"SolicitudId"`
	Observacion string `json:"Observacion"`
	TerceroId   int    `json:"TerceroId"`
	RolUsuario  string `json:"RolUsuario"`
}
