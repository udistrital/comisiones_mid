package models

type HistoricoEstadoSolicitud struct {
	Id                int              `json:"Id"`
	SolicitudId       *Solicitud       `json:"SolicitudId"`
	EstadoSolicitudId *EstadoSolicitud `json:"EstadoSolicitudId"`
	RolUsuario        string           `json:"RolUsuario"`
	TerceroId         int              `json:"TerceroId"`
	Activo            bool             `json:"Activo"`
}
