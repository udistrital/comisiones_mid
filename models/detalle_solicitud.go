package models

type DetalleSolicitud struct {
	SolicitudId		*SolicitudInicial
	Formulario		string
	Activo			bool
}
