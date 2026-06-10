package models

type Comision struct {
	Id                int    `orm:"column(id);pk;auto"`
	Descripcion       string `orm:"column(descripcion);null"`
	Activo            bool   `orm:"column(activo);null"`
	FechaInicio       string `orm:"column(fecha_inicio);type(timestamp without time zone);null"`
	FechaFinal        string `orm:"column(fecha_final);type(timestamp without time zone);null"`
	FechaCreacion     string `orm:"column(fecha_creacion);type(timestamp without time zone);null"`
	FechaModificacion string `orm:"column(fecha_modificacion);type(timestamp without time zone);null"`
}
