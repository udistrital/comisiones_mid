package models

type SolicitudPendienteRevisor struct {
	Id               int    `json:"id"`
	FechaCreacion    string `json:"fecha_creacion"`
	NombreDocente    string `json:"nombre_docente"`
	DocumentoDocente string `json:"documento_docente"`
	EstadoSolicitud  string `json:"estado_solicitud"`
}
