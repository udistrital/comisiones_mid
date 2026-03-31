package models

type CancelarSolicitudResponse struct {
	SolicitudId                     int   `json:"SolicitudId"`
	SolicitudDesactivada            bool  `json:"SolicitudDesactivada"`
	DetalleSolicitudDesactivados    []int `json:"DetalleSolicitudDesactivados,omitempty"`
	HistoricosDesactivados          []int `json:"HistoricosDesactivados,omitempty"`
	ObservacionesDesactivadas       []int `json:"ObservacionesDesactivadas,omitempty"`
	DocumentosSolicitudDesactivados []int `json:"DocumentosSolicitudDesactivados,omitempty"`
}
