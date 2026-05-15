package models

// DetalleComisionSeguimiento agrupa los datos del panel de gestión de una comisión (fase 2).
// Combina datos de: solicitud, detalle_solicitud (formulario FR-010) e historico_estado_comision.
type DetalleComisionSeguimiento struct {
	ComisionId         int    `json:"comision_id"`
	SolicitudId        int    `json:"solicitud_id"`
	Radicado           string `json:"radicado"`
	Docente            string `json:"docente"`
	IdDocente          string `json:"id_docente"`
	CorreoDocente      string `json:"correo_docente"`
	Facultad           string `json:"facultad"`
	Programa           string `json:"programa"`
	TipoEstudio        string `json:"tipo_estudio"`
	UniversidadDestino string `json:"universidad_destino"`
	PaisDestino        string `json:"pais_destino"`
	CiudadDestino      string `json:"ciudad_destino"`
	Duracion           string `json:"duracion"`
	FechaSolicitud     string `json:"fecha_solicitud"`
	FechaInicio        string `json:"fecha_inicio"`
	FechaFin           string `json:"fecha_fin"`
	EstadoComision     string `json:"estado_comision"`
}
