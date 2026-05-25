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
	Data    HistoricoEstadoSolicitud `json:"Data"`
	Message string                   `json:"Message"`
	Status  string                   `json:"Status"`
	Success bool                     `json:"Success"`
}

type ResponseListaHistoricoEstadoSolicitud struct {
	Data    []HistoricoEstadoSolicitud `json:"Data"`
	Message string                     `json:"Message"`
	Status  string                     `json:"Status"`
	Success bool                       `json:"Success"`
}

type HistoricoSolicitudProrroga struct {
	SolicitudId   int    `json:"solicitud_id"`
	HistoricoId   int    `json:"historico_id"`
	FechaCreacion string `json:"fecha_creacion"`
	Estado        string `json:"estado"`
}

type ResponseHistoricoSolicitudProrroga struct {
	Success bool                         `json:"Success"`
	Status  string                       `json:"Status"`
	Message string                       `json:"Message"`
	Data    []HistoricoSolicitudProrroga `json:"Data"`
}
