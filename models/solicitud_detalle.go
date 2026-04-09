package models

type SolicitudDetalles struct {
	Solicitud       *Solicitud
	EstadoSolicitud *EstadoSolicitud
	Formulario      interface{}
	Observacion     string
	Documentos      []DocumentoDetalle
}

type DocumentoDetalle struct {
	Id     int
	Nombre string
	Enlace string
	Tipo   *TipoDocumentoSolicitud
	Estado *EstadoDocumento
}
