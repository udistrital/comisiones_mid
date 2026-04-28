package models

type ActualizarEstadoDocumentoSolicitudRequest struct {
	DocumentoSolicitudId  int    `json:"DocumentoSolicitudId"`
	EstadoDocumentoCodigo string `json:"EstadoDocumentoCodigo,omitempty"`
}

type ActualizarEstadoDocumentoSolicitudResponse struct {
	DocumentoSolicitudId      int                    `json:"DocumentoSolicitudId"`
	EstadoDocumentoAnteriorId int                    `json:"EstadoDocumentoAnteriorId,omitempty"`
	EstadoDocumentoNuevoId    int                    `json:"EstadoDocumentoNuevoId"`
	Mensaje                   string                 `json:"Mensaje"`
	CrudResponse              map[string]interface{} `json:"CrudResponse,omitempty"`
}
