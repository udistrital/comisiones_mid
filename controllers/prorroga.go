package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/comisiones_mid/services"
)

// ProrrogaController operations for Prorroga
type ProrrogaController struct {
	beego.Controller
}

// URLMapping ...
func (c *ProrrogaController) URLMapping() {
	c.Mapping("CrearSolicitudProrroga", c.CrearSolicitudProrroga)
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
		"Message": "Consulta exitosa",
		"Data":    data,
	}
	c.ServeJSON()
}
