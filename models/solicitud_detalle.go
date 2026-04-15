package models

type SolicitudDetalles struct {
	Solicitud       *Solicitud
	EstadoSolicitud *EstadoSolicitud
	Formulario      interface{}
	Observacion     string
	Documentos      []DocumentoDetalle
	Observaciones   []ObservacionDetalle
}

type DocumentoDetalle struct {
	Id          int
	Rol         string
	IdDocumento int
	Nombre      string
	Enlace      string
	Tipo        *TipoDocumentoSolicitud
	Estado      *EstadoDocumento
}

type ObservacionDetalle struct {
	Id          int
	Rol         string
	Descripcion string
}
