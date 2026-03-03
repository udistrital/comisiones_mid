package controllers

import (
	"github.com/astaxie/beego/logs"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/comisiones_mid/services"
)

// ComisionController operations for Comision
type ComisionController struct {
	beego.Controller
}

// URLMapping ...
func (c *ComisionController) URLMapping() {
	c.Mapping("CrearComision", c.CrearComision)
}

// Post ...
// @Title Create
// @Description create Comision
// @Param	body		body 	models.Comision	true		"body for Comision content"
// @Success 201 {object} models.Comision
// @Failure 403 body is empty
// @router /CrearComision [post]
func (c *ComisionController) CrearComision() {
	defer func() {
		if err := recover(); err != nil {
			logs.Error(err)
			localError := err.(map[string]interface{})
			c.Data["mesaage"] = (beego.AppConfig.String("appname") + "/" + "CrearComision" + "/" + (localError["funcion"]).(string))
			c.Data["data"] = (localError["err"])
			if status, ok := localError["status"]; ok {
				c.Abort(status.(string))
			} else {
				c.Abort("404")
			}
		}
	}()
	var v models.CrearComisionEntrada
	//var v map[string]interface{}
	json.Unmarshal(c.Ctx.Input.RequestBody, &v)
	if response, err := services.CrearSolicitud(v); err == nil {
		c.Ctx.Output.SetStatus(201)
		c.Data["json"] = map[string]interface{}{"Success": true, "Status": "201", "Message": "Successful", "Data": response}
	} else {
		panic(err)
	}
	c.ServeJSON()
}
