package models

// DocumentoPagoItem representa un documento subido al panel de gestión de pagos.
type DocumentoPagoItem struct {
	DocumentoComisionId int    `json:"documento_comision_id"`
	DocumentoId         int    `json:"documento_id"`
	Nombre              string `json:"nombre"`
	SubidoPorRol        string `json:"subido_por_rol"`
	SubidoPorNombre     string `json:"subido_por_nombre"`
	Enlace              string `json:"enlace"`
	Estado              string `json:"estado"`
	EstadoNombre        string `json:"estado_nombre"`
}
