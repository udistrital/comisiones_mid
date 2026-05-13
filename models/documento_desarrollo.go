package models

// DocumentoDesarrolloItem representa un tipo de documento con su estado actual en la comision.
// Si el docente aun no ha subido ese documento, DocumentoComisionId y DocumentoId son 0.
type DocumentoDesarrolloItem struct {
	TipoId              int    `json:"tipo_id"`
	Codigo              string `json:"codigo"`
	Nombre              string `json:"nombre"`
	DocumentoComisionId int    `json:"documento_comision_id"`
	DocumentoId         int    `json:"documento_id"`
	Enlace              string `json:"enlace"`
	Estado              string `json:"estado"`
	EstadoNombre        string `json:"estado_nombre"`
}

// GrupoDocumentosDesarrollo agrupa documentos por momento del proceso.
type GrupoDocumentosDesarrollo struct {
	Momento    string                   `json:"momento"`
	Prefijo    string                   `json:"prefijo"`
	Documentos []DocumentoDesarrolloItem `json:"documentos"`
}

// SubirDocumentoDesarrolloRequest es el body del POST /documento_desarrollo.
type SubirDocumentoDesarrolloRequest struct {
	ComisionId          int    `json:"comision_id"`
	TipoDocumentoCodigo string `json:"tipo_documento_codigo"`
	IdTipoDocumento     int    `json:"id_tipo_documento"`
	Nombre              string `json:"nombre"`
	Descripcion         string `json:"descripcion"`
	File                string `json:"file"`
}
