package models

type CrearDocumentoGestorDocumental struct{
	IdTipoDocumento int         `json:"IdTipoDocumento"`
	Nombre          string      `json:"nombre"`
	Descripcion     string      `json:"descripcion"`
	Metadatos       interface{} `json:"metadatos"`
	File            string      `json:"file"`
}