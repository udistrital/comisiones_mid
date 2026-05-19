package models

type HistoricoEstadoSolicitud struct {
	Id                int              `json:"Id"`
	SolicitudId       *Solicitud       `json:"SolicitudId"`
	EstadoSolicitudId *EstadoSolicitud `json:"EstadoSolicitudId"`
	RolUsuario        string           `json:"RolUsuario"`
	TerceroId         int              `json:"TerceroId"`
	Activo            bool             `json:"Activo"`
}

type ResponseCreateHistoricoEstadoSolicitud struct {
	Data    Solicitud `json:"Data"`
	Message string    `json:"Message"`
	Status  string    `json:"Status"`
	Success bool      `json:"Success"`
}
