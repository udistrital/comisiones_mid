package models

type SolicitudInicial struct {
	Id					int
	TerceroId       	int
	TipoSolicitudId 	*TipoSolicitud
	Activo          	bool
}

type TipoSolicitud struct {
	Id int
}

type ResponseSolicitud struct {
	Data    SolicitudInicial 		`json:"Data"`
	Message string                  `json:"Message"`
	Status  string                  `json:"Status"`
	Success bool                    `json:"Success"`
}