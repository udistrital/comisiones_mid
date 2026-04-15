package models

type DocumentoCambioEstadoRequest struct {
	IdTipoDocumento int                    `json:"IdTipoDocumento"`
	TipoDocumento   string                 `json:"TipoDocumento,omitempty"`
	EstadoDocumento string                 `json:"EstadoDocumento,omitempty"`
	Nombre          string                 `json:"Nombre"`
	Descripcion     string                 `json:"Descripcion,omitempty"`
	Metadatos       map[string]interface{} `json:"Metadatos,omitempty"`
	File            string                 `json:"File"`
}

type CambioEstadoSolicitudRequest struct {
	SolicitudId          int                            `json:"SolicitudId"`
	NuevoEstado          string                         `json:"NuevoEstado"`
	RolUsuario           string                         `json:"RolUsuario"`
	NumeroIdentificacion string                         `json:"NumeroIdentificacion"`
	Observacion          string                         `json:"Observacion,omitempty"`
	Documentos           []DocumentoCambioEstadoRequest `json:"Documentos,omitempty"`
}

type CambioEstadoSolicitudResponse struct {
	SolicitudId               int                    `json:"SolicitudId"`
	EstadoAnteriorId          int                    `json:"EstadoAnteriorId,omitempty"`
	EstadoDestinoId           int                    `json:"EstadoDestinoId"`
	TerceroId                 int                    `json:"TerceroId"`
	HistoricoAnteriorId       int                    `json:"HistoricoAnteriorId,omitempty"`
	HistoricoNuevoId          int                    `json:"HistoricoNuevoId,omitempty"`
	ObservacionId             int                    `json:"ObservacionId,omitempty"`
	DocumentoId               int                    `json:"DocumentoId,omitempty"`
	DocumentoSolicitudId      int                    `json:"DocumentoSolicitudId,omitempty"`
	DocumentoIds              []int                  `json:"DocumentoIds,omitempty"`
	DocumentoSolicitudIds     []int                  `json:"DocumentoSolicitudIds,omitempty"`
	ComisionId                int                    `json:"ComisionId,omitempty"`
	HistoricoEstadoComisionId int                    `json:"HistoricoEstadoComisionId,omitempty"`
	Mensaje                   string                 `json:"Mensaje"`
	CrudResponse              map[string]interface{} `json:"CrudResponse,omitempty"`
}
