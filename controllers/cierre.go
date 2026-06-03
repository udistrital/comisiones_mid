package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/comisiones_mid/services"
)

// ProrrogaController operations for Cierre
type CierreController struct {
	beego.Controller
}

// URLMapping ...
func (c *CierreController) URLMapping() {
	c.Mapping("CrearSolicitudCierre", c.CrearSolicitudCierre)
}

// Post ...
// @Title Create
// @Description create Cierre
// @Param	body		body 	models.Cierre	true		"body for Cierre content"
// @Success 201 {object} models.Cierre
// @Failure 403 body is empty
// @router /crear_solicitud_cierre [post]
func (c *CierreController) CrearSolicitudCierre() {
	fmt.Println("ENTRA A CREAR CIERRE")
	var v models.CrearSolicitudCierreEntrada
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
	data, err := services.CrearSolicitudCierre(v)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "500",
			"Message": "Error creando la solicitud de cierre",
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
