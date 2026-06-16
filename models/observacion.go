package models

type ObservacionCreate struct {
	HistoricoEstadoSolicitudId *HistoricoEstadoSolicitud `json:"HistoricoEstadoSolicitudId"`
	Descripcion                string                    `json:"Descripcion"`
	Activo                     bool                      `json:"Activo"`
}
