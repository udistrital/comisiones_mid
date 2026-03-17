package models

type CrearSolicitudEntrada struct {
	Identificacion     		int                                   	`json:"identificacion"`
	TipoSolicitudId    		int                                  	`json:"tipo_solicitud_id"`
	Formulario		   		map[string]interface{}                	`json:"formulario"`
	DocumentoSolicitud 		[]CrearDocumentoGestorDocumental     	`json:"documento_solicitud"`
	Observacion				string									`json:"observacion"`
	CodigoAbreviacionRol	string									`json:"cod_abreviacion_rol"`
}