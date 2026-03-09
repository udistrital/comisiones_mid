package models

type SolicitudInicial struct {     
	TerceroId         int            
	TipoSolicitudId   *TipoSolicitud     
	Activo            bool              
}

type TipoSolicitud struct{
	Id	int
}