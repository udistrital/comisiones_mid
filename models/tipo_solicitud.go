package models

type TipoSolicitud struct {
	Id                int    `json:"Id"`
	Nombre            string `json:"Nombre"`
	CodigoAbreviacion string `json:"CodigoAbreviacion"`
}

type ResponseListaTipoSolicitud struct {
	Data    []TipoSolicitud `json:"Data"`
	Message string          `json:"Message"`
	Status  string          `json:"Status"`
	Success bool            `json:"Success"`
}
