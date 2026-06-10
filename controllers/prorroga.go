package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/comisiones_mid/services"
)

const consultaExitosa2 = "Consulta exitosa"

// ProrrogaController operations for Prorroga
type ProrrogaController struct {
	beego.Controller
}

// URLMapping ...
func (c *ProrrogaController) URLMapping() {
	c.Mapping("CrearSolicitudProrroga", c.CrearSolicitudProrroga)
	c.Mapping("ValidarSolicitudProrroga", c.ValidarSolicitudProrroga)
	c.Mapping("ConsultarHistoricoSolicitudesProrroga", c.ConsultarHistoricoSolicitudesProrroga)
}

// Post ...
// @Title Create
// @Description create Prorroga
// @Param	body		body 	models.Prorroga	true		"body for Prorroga content"
// @Success 201 {object} models.Prorroga
// @Failure 403 body is empty
// @router /crear_solicitud_prorroga [post]
func (c *ProrrogaController) CrearSolicitudProrroga() {
	fmt.Println("ENTRA A CREAR PRORROGA")
	var v models.CrearSolicitudProrrogaEntrada
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err != nil {

		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  400,
			"Message": "JSON inválido",
			"Data":    nil,
		}
		c.ServeJSON()
		return
	}
	data, err := services.CrearSolicitudProrroga(v)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "500",
			"Message": "Error creando la solicitud de prorroga",
			"Error":   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = map[string]interface{}{
		"Success": true,
		"Status":  "200",
		"Message": consultaExitosa2,
		"Data":    data,
	}
	c.ServeJSON()
}

// ======================================================
// VALIDAR SI PUEDE CREAR PRÓRROGA
// ======================================================

// Get ...
// @Title Validar Solicitud Prórroga
// @Description valida si una comisión puede crear una nueva prórroga
// @Param	id	path	int	true		"Id comisión"
// @Success 200 {object} map[string]interface{}
// @Failure 400 id inválido
// @Failure 500 error interno
// @router /validar_solicitud_prorroga/:id [get]
func (c *ProrrogaController) ValidarSolicitudProrroga() {

	idParam := c.Ctx.Input.Param(":id")

	if idParam == "" {

		c.Ctx.Output.SetStatus(400)

		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "400",
			"Message": "El id de la comisión es obligatorio",
			"Data":    nil,
		}

		c.ServeJSON()
		return
	}

	// CONVERTIR ID A ENTERO
	comisionId, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {

		c.Ctx.Output.SetStatus(400)

		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "400",
			"Message": "El id de la comisión es inválido",
			"Data":    nil,
		}

		c.ServeJSON()
		return
	}

	puedeCrear, mensaje, err := services.ValidarPuedeCrearSolicitudProrroga(
		int(comisionId),
	)

	if err != nil {

		c.Ctx.Output.SetStatus(500)

		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "500",
			"Message": "Error validando solicitud de prórroga",
			"Error":   err.Error(),
		}

		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(200)

	c.Data["json"] = map[string]interface{}{
		"Success": true,
		"Status":  "200",
		"Message": consultaExitosa2,
		"Data": map[string]interface{}{
			"puede_crear_prorroga": puedeCrear,
			"mensaje":              mensaje,
		},
	}

	c.ServeJSON()
}

// @Title ConsultarHistoricoSolicitudesProrroga
// @Description Consulta el histórico de solicitudes de prórroga
// @Success 200 {object} models.ResponseHistoricoSolicitudProrroga
// @Failure 400 id inválido
// @Failure 500 error interno
// @router /historico_solicitudes_prorroga/:id [get]
func (c *ProrrogaController) ConsultarHistoricoSolicitudesProrroga() {

	idParam := c.Ctx.Input.Param(":id")

	if idParam == "" {

		c.Ctx.Output.SetStatus(400)

		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "400",
			"Message": "El id de la comisión es obligatorio",
			"Data":    nil,
		}

		c.ServeJSON()
		return
	}

	comisionId, err := strconv.Atoi(idParam)

	if err != nil {

		c.Ctx.Output.SetStatus(400)

		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "400",
			"Message": "El id de la comisión es inválido",
			"Data":    nil,
		}

		c.ServeJSON()
		return
	}

	data, err := services.ConsultarHistoricoSolicitudesProrroga(comisionId)

	if err != nil {

		c.Ctx.Output.SetStatus(500)

		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "500",
			"Message": "Error consultando histórico de solicitudes de prórroga",
			"Error":   err.Error(),
			"Data":    nil,
		}

		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(200)

	c.Data["json"] = map[string]interface{}{
		"Success": true,
		"Status":  "200",
		"Message": consultaExitosa2,
		"Data":    data,
	}

	c.ServeJSON()
}
