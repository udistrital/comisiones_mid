package models

type EditarSolicitud struct {
	TipoSolicitudId      int                            `json:"tipo_solicitud_id,omitempty"`
	Formulario           map[string]interface{}         `json:"formulario,omitempty"`
	Observacion          string                         `json:"observacion,omitempty"`
	DocumentosNuevos     []DocumentoCambioEstadoRequest `json:"documentos_nuevos,omitempty"`
	DocumentosDesactivar []int                          `json:"documentos_desactivar,omitempty"`
}

type EditarSolicitudResponse struct {
	SolicitudId                int    `json:"SolicitudId"`
	DetalleSolicitudId         int    `json:"DetalleSolicitudId,omitempty"`
	HistoricoEstadoSolicitudId int    `json:"HistoricoEstadoSolicitudId,omitempty"`
	DocumentoIds               []int  `json:"DocumentoIds,omitempty"`
	DocumentoSolicitudIds      []int  `json:"DocumentoSolicitudIds,omitempty"`
	DocumentosDesactivados     []int  `json:"DocumentosDesactivados,omitempty"`
	Mensaje                    string `json:"Mensaje"`
}
