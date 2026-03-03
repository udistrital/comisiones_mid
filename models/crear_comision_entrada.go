package models

type CrearComisionEntrada struct {
	Tercero_id string 	`json:"tercero_id"`
	Detalle_solicitud 	map[string]interface{} `json:"detalle_solicitud"`
	Documento_solicitud string `json:"documento_solicitud"`
}

