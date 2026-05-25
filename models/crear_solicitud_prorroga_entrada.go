package models

type CrearSolicitudProrrogaEntrada struct {
	ComisionId                  int                 `json:"comision_id"`
	DocumentosSolicitudProrroga []DocumentoProrroga `json:"documentos_solicitud_prorroga"`
	Observacion                 string              `json:"observacion"`
	CodigoAbreviacionRol        string              `json:"cod_abreviacion_rol"`
}

type DocumentoProrroga struct {
	CodigoAbreviacionDoc string                         `json:"codigo_abreviacion"`
	DocumentoSolicitud   CrearDocumentoGestorDocumental `json:"documento_solicitud"`
}
