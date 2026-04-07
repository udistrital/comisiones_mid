package models

type DocumentoSolicitud struct {
	Id int `json:"Id"`
	DocumentoId int `json:"DocumentoId"`
	HistoricoEstadoSolicitudId *HistoricoEstadoSolicitud `json:"HistoricoEstadoSolicitudId"`
	TipoDocumentoId *TipoDocumentoSolicitud `json:"TipoDocumentoId"`
	EstadoDocumentoId *EstadoDocumento `json:"EstadoDocumentoId"`
	Activo bool `json:"Activo"`
}