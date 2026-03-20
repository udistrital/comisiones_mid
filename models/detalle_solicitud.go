package models

type DetalleSolicitud struct {
	SolicitudId *Solicitud
	Formulario  string
	Activo      bool
}
