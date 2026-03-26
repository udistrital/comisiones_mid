package models

type SolicitudDetalles struct {
	Solicitud       *Solicitud
	EstadoSolicitud *EstadoSolicitud
	Formulario      interface{}
	Documentos      []DocumentoDetalle
}

type DocumentoDetalle struct {
	Nombre string
	Enlace string
	Tipo   *TipoDocumentoSolicitud
	Estado *EstadoDocumento
}
