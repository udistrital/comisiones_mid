package models

type DocumentoSolicitud struct {
	Id                    int                       `orm:"column(id);pk;auto"`
	DocumentoId           int                       `orm:"column(documento_solicitud);null"`
	SolicitudEstadoEvento *HistoricoEstadoSolicitud `orm:"column(solicitud_estado_evento_id);null"`
	TipoDocumento         *TipoDocumentoSolicitud   `orm:"column(tipo_documento_id);null"`
	EstadoDocumento       *EstadoDocumento          `orm:"column(estado_documento_id);null"`
	Activo                bool                      `orm:"column(activo);null"`
	FechaCreacion         string                    `orm:"column(fecha_creacion);type(timestamp without time zone);null"`
	FechaModificacion     string                    `orm:"column(fecha_modificacion);type(timestamp without time zone);null"`
}
