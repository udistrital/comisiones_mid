package models

type CrearSolicitudEntrada struct {
	Identificacion     int
	TipoSolicitudId    int
	DetalleSolicitud   map[string]interface{}
	DocumentoSolicitud string
}
