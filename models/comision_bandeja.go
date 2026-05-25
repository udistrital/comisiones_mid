package models

// ComisionBandeja representa una fila en la bandeja principal de seguimiento de comisiones (fase 2).
// Agrega datos de: historico_estado_comision, comision, solicitud y detalle_solicitud (formulario).
// Los campos de fecha de comision (FechaInicio, FechaFin) pueden ser cadena vacia si aun no
// han sido registrados en la tabla comision.
type ComisionBandeja struct {
	ComisionId     int    `json:"comision_id"`
	SolicitudId    int    `json:"solicitud_id"`
	TerceroId      int    `json:"tercero_id"`
	Docente        string `json:"docente"`
	IdDocente      string `json:"id_docente"`
	Programa       string `json:"programa"`
	Facultad       string `json:"facultad"`
	FechaSolicitud string `json:"fecha_solicitud"`
	FechaInicio    string `json:"fecha_inicio"`
	FechaFin       string `json:"fecha_fin"`
	EstadoComision string `json:"estado_comision"`
}
