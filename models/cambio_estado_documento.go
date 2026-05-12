package models

type ActualizarEstadoDocumentoSolicitudRequest struct {
	DocumentoSolicitudId  int    `json:"DocumentoSolicitudId"`
	EstadoDocumentoCodigo string `json:"EstadoDocumentoCodigo,omitempty"`
}

type ActualizarEstadosDocumentoSolicitudRequest struct {
	Documentos []ActualizarEstadoDocumentoSolicitudRequest `json:"Documentos"`
}

type ActualizarEstadoDocumentoSolicitudResponse struct {
	DocumentoSolicitudId      int                    `json:"DocumentoSolicitudId"`
	EstadoDocumentoAnteriorId int                    `json:"EstadoDocumentoAnteriorId,omitempty"`
	EstadoDocumentoNuevoId    int                    `json:"EstadoDocumentoNuevoId"`
	Mensaje                   string                 `json:"Mensaje"`
	CrudResponse              map[string]interface{} `json:"CrudResponse,omitempty"`
}

type ActualizarEstadosDocumentoSolicitudResponse struct {
	Resultados []ActualizarEstadoDocumentoSolicitudResponse `json:"Resultados"`
	Total      int                                          `json:"Total"`
	Mensaje    string                                       `json:"Mensaje"`
}
