package models

type CambioEstadoSolicitudRequest struct {
	SolicitudId          int    `json:"SolicitudId"`
	NuevoEstado          string `json:"NuevoEstado"`
	RolUsuario           string `json:"RolUsuario"`
	NumeroIdentificacion string `json:"NumeroIdentificacion"`
	TipoDocumento        string `json:"TipoDocumento,omitempty"`
	EstadoDocumento      string `json:"EstadoDocumento,omitempty"`
	//	Base64Documento      string `json:"Base64Documento,omitempty"`
	NombreArchivo        string `json:"NombreArchivo,omitempty"`
	DescripcionDocumento string `json:"descripcion_documento,omitempty"`
	Metadatos            string `json:"metadatos,omitempty"`
}

type CambioEstadoSolicitudResponse struct {
	SolicitudId          int                    `json:"SolicitudId"`
	EstadoAnteriorId     int                    `json:"EstadoAnteriorId"`
	HistoricoAnteriorId  int                    `json:"HistoricoAnteriorId"`
	EstadoDestinoId      int                    `json:"EstadoDestinoId"`
	HistoricoNuevoId     int                    `json:"HistoricoNuevoId"`
	TerceroId            int                    `json:"TerceroId"`
	DocumentoId          int                    `json:"DocumentoId,omitempty"`
	DocumentoSolicitudId int                    `json:"DocumentoSolicitudId,omitempty"`
	CrudResponse         map[string]interface{} `json:"CrudResponse"`
	Mensaje              string                 `json:"Mensaje"`
}
