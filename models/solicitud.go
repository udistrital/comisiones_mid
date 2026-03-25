package models

type Solicitud struct {
	Id                int            `json:"Id"`
	TerceroId         int            `json:"TerceroId"`
	TipoSolicitudId   *TipoSolicitud `json:"TipoSolicitudId"`
	ComisionId        *Comision      `json:"ComisionId"`
	ObservacionCierre string         `json:"ObservacionCierre"`
	Activo            bool           `json:"Activo"`
	FechaCreacion     string         `json:"FechaCreacion"`
	FechaModificacion string         `json:"FechaModificacion"`
}

type SolicitudCreateRequest struct {
	Id                int          `json:"Id"`
	TerceroId         int          `json:"TerceroId"`
	TipoSolicitudId   IdReference  `json:"TipoSolicitudId"`
	ComisionId        *IdReference `json:"ComisionId"`
	ObservacionCierre string       `json:"ObservacionCierre"`
	Activo            bool         `json:"Activo"`
}

type ResponseSolicitud struct {
	Data    Solicitud `json:"Data"`
	Message string    `json:"Message"`
	Status  string    `json:"Status"`
	Success bool      `json:"Success"`
}
