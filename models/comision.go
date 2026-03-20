package models

type Comision struct {
	Id                int    `orm:"column(id);pk;auto"`
	Descripcion       string `orm:"column(descripcion);null"`
	Activo            bool   `orm:"column(activo);null"`
	FechaCreacion     string `orm:"column(fecha_creacion);type(timestamp without time zone);null"`
	FechaModificacion string `orm:"column(fecha_modificacion);type(timestamp without time zone);null"`
}
