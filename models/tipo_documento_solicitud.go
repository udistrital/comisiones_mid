package models

type TipoDocumentoSolicitud struct {
	Id                int    `orm:"column(id);pk;auto"`
	Nombre            string `orm:"column(nombre);null"`
	Descripcion       string `orm:"column(descripcion);null"`
	CodigoAbreviacion string `orm:"column(codigo_abreviacion);null"`
	Activo            bool   `orm:"column(activo);null"`
	FechaCreacion     string `orm:"column(fecha_creacion);type(timestamp without time zone);null"`
	FechaModificacion string `orm:"column(fecha_modificacion);type(timestamp without time zone);null"`
}
