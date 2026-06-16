package models

type HistoricoEstadoComision struct {
	Id               int             `json:"Id"`
	ComisionId       *Comision       `json:"ComisionId"`
	EstadoComisionId *EstadoComision `json:"EstadoComisionId"`
	RolUsuario       string          `json:"RolUsuario"`
	TerceroId        int             `json:"TerceroId"`
	Activo           bool            `json:"Activo"`
	Descripcion      string          `json:"Descripcion"`
}

type HistoricoEstadoComisionPUT struct {
	Id               int             `json:"Id"`
	ComisionId       *Comision       `json:"ComisionId"`
	EstadoComisionId *EstadoComision `json:"EstadoComisionId"`
	RolUsuario       string          `json:"RolUsuario"`
	TerceroId        int             `json:"TerceroId"`
	Activo           bool            `json:"Activo"`
	FechaCreacion    string          `json:"FechaCreacion"`
}

type ResponseCreateHistoricoEstadoComision struct {
	Data    HistoricoEstadoComision `json:"Data"`
	Message string                  `json:"Message"`
	Status  string                  `json:"Status"`
	Success bool                    `json:"Success"`
}

type ResponseListaHistoricoEstadoComision struct {
	Data    []HistoricoEstadoComision `json:"Data"`
	Message string                    `json:"Message"`
	Status  string                    `json:"Status"`
	Success bool                      `json:"Success"`
}

type ResponseListaHistoricoEstadoComisionPUT struct {
	Data    []HistoricoEstadoComisionPUT `json:"Data"`
	Message string                       `json:"Message"`
	Status  string                       `json:"Status"`
	Success bool                         `json:"Success"`
}

type HistoricoComisionProrroga struct {
	ComisionId    int    `json:"Comision_id"`
	HistoricoId   int    `json:"historico_id"`
	FechaCreacion string `json:"fecha_creacion"`
	Estado        string `json:"estado"`
}

type ResponseHistoricoComisionProrroga struct {
	Success bool                        `json:"Success"`
	Status  string                      `json:"Status"`
	Message string                      `json:"Message"`
	Data    []HistoricoComisionProrroga `json:"Data"`
}

type HistoricoComisionCierre struct {
	ComisionId    int    `json:"Comision_id"`
	HistoricoId   int    `json:"historico_id"`
	FechaCreacion string `json:"fecha_creacion"`
	Estado        string `json:"estado"`
}

type ResponseHistoricoComisionCierre struct {
	Success bool                      `json:"Success"`
	Status  string                    `json:"Status"`
	Message string                    `json:"Message"`
	Data    []HistoricoComisionCierre `json:"Data"`
}
