package models

type SolicitudDetalles struct {
	Solicitud  *Solicitud
	Formulario interface{}
	Documentos *DocumentoDetalle
}

type DocumentoDetalle struct{
	Nombre	string
	Enlace	string
	Tipo	*TipoDocumentoSolicitud
	Estado	*EstadoDocumento
}
