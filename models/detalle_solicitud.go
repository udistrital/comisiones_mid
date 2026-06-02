package models

type DetalleSolicitud struct {
	SolicitudId *Solicitud
	Formulario  string
	Activo      bool
}

type ResponseDetalleSolicitud struct {
	Data    []DetalleSolicitud `json:"Data"`
	Message string             `json:"Message"`
	Status  string             `json:"Status"`
	Success bool               `json:"Success"`
}

type ResponseCreateDetalleSolicitud struct {
	Data    Solicitud `json:"Data"`
	Message string    `json:"Message"`
	Status  string    `json:"Status"`
	Success bool      `json:"Success"`
}
