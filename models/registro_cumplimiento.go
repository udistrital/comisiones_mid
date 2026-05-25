package models

// EstadoCumplimientoItem es un estado de comision de tipo cumplimiento,
// retornado por GET /estados_cumplimiento para poblar el dropdown del decano.
type EstadoCumplimientoItem struct {
	Id     int    `json:"id"`
	Nombre string `json:"nombre"`
	Codigo string `json:"codigo"`
}

// RegistroCumplimientoItem representa una entrada del historial de cumplimiento
// de una comision, retornado por GET /historial_cumplimiento/:comision_id.
type RegistroCumplimientoItem struct {
	Id            int    `json:"id"`
	Descripcion   string `json:"descripcion"`
	FechaCreacion string `json:"fecha_creacion"`
	EstadoId      int    `json:"estado_id"`
	EstadoCodigo  string `json:"estado_codigo"`
	EstadoNombre  string `json:"estado_nombre"`
	Activo        bool   `json:"activo"`
}

// CrearRegistroCumplimientoRequest es el body del POST /registro_cumplimiento.
type CrearRegistroCumplimientoRequest struct {
	ComisionId           int    `json:"comision_id"`
	EstadoId             int    `json:"estado_id"`
	Descripcion          string `json:"descripcion"`
	Rol                  string `json:"rol"`
	Nombre               string `json:"nombre"`
	NumeroIdentificacion string `json:"numero_identificacion"`
}
