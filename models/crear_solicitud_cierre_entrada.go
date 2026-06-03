package models

type CrearSolicitudCierreEntrada struct {
	ComisionId           int    `json:"comision_id"`
	Observacion          string `json:"observacion"`
	CodigoAbreviacionRol string `json:"cod_abreviacion_rol"`
}
