package models

type CambioEstadoSolicitudRequest struct {
	SolicitudId           int    `json:"SolicitudId"`
	EstadoDestinoCodigo   string `json:"EstadoDestinoCodigo"`
	RolUsuarioCodigo      string `json:"RolUsuarioCodigo"`
	NumeroIdentificacion  string `json:"NumeroIdentificacion"`
	DocumentoApiId        int    `json:"DocumentoApiId,omitempty"`
	TipoDocumentoCodigo   string `json:"TipoDocumentoCodigo,omitempty"`
	EstadoDocumentoCodigo string `json:"EstadoDocumentoCodigo,omitempty"`
}

type CambioEstadoSolicitudResponse struct {
	SolicitudId          int                    `json:"SolicitudId"`
	EstadoAnteriorId     int                    `json:"EstadoAnteriorId"`
	HistoricoAnteriorId  int                    `json:"HistoricoAnteriorId"`
	EstadoDestinoId      int                    `json:"EstadoDestinoId"`
	HistoricoNuevoId     int                    `json:"HistoricoNuevoId"`
	TerceroId            int                    `json:"TerceroId"`
	DocumentoUuid        string                 `json:"DocumentoUuid,omitempty"`
	DocumentoSolicitudId int                    `json:"DocumentoSolicitudId,omitempty"`
	CrudResponse         map[string]interface{} `json:"CrudResponse"`
	Mensaje              string                 `json:"Mensaje"`
}
