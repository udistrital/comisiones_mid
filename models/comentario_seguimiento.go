package models

// ComentarioSeguimiento representa un comentario de un panel de comision, listo para el front.
type ComentarioSeguimiento struct {
	Id            int    `json:"id"`
	Rol           string `json:"rol"`
	Texto         string `json:"texto"`
	FechaCreacion string `json:"fecha_creacion"`
}

// CrearComentarioRequest es el body del POST /comentario.
type CrearComentarioRequest struct {
	ComisionId            int    `json:"comision_id"`
	CodigoTipoSeguimiento string `json:"codigo_tipo_seguimiento"`
	Rol                   string `json:"rol"`
	Nombre                string `json:"nombre"`
	NumeroIdentificacion  string `json:"numero_identificacion"`
	Texto                 string `json:"texto"`
}
