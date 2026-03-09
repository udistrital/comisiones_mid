package models

type Documento struct {
	Nombre        string          `json:"Nombre"`
	Descripcion   string          `json:"Descripcion"`
	TipoDocumento *TipoDocumento  `json:"TipoDocumento"`
	Metadatos     string          `json:"Metadatos"`
	Activo        bool            `json:"Activo"`
}

type TipoDocumento struct {
	Id int `json:"Id"`
}