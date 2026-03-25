package models

type SolicitudResumen struct {
	Id       			int    `json:"id"`
	Activo   			bool   `json:"activo"`
	Programa 			string `json:"programa"`
	Nombre   			string `json:"nombre"`
	FechaCreacion		string `json:"fecha_creacion"`
	EstadoSolicitud		*EstadoSolicitud `json:"esado_solicitud"`
}
