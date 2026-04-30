package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/comisiones_mid/services"
)

// ComisionSeguimientoController expone los endpoints de bandeja de seguimiento (fase 2).
// Todos los endpoints estan bajo el namespace /v1/seguimiento.
type ComisionSeguimientoController struct {
	beego.Controller
}

// GetComisionesSecretariaGeneral ...
// @Title Get Comisiones Secretaria General
// @Description Retorna todas las comisiones activas con su estado actual. Usado por secretaria general/academica.
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @router /comisiones_secretaria_general [get]
func (c *ComisionSeguimientoController) GetComisionesSecretariaGeneral() {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("[ComisionSeguimiento] panic en GetComisionesSecretariaGeneral: %v", r)
			c.Ctx.Output.SetStatus(500)
			c.Data["json"] = map[string]interface{}{
				"Success": false,
				"Status":  "500",
				"Message": "Error interno del servidor",
			}
			c.ServeJSON()
		}
	}()

	data, err := services.ObtenerBandejaSecretariaGeneral()
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "500",
			"Message": "Error obteniendo comisiones",
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

// GetComisionesDocente ...
// @Title Get Comisiones Docente
// @Description Retorna las comisiones activas del docente identificado por su numero de cedula.
// @Param	cedula	path	string	true	"Numero de cedula del docente"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @router /comisiones_docente/:cedula [get]
func (c *ComisionSeguimientoController) GetComisionesDocente() {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("[ComisionSeguimiento] panic en GetComisionesDocente: %v", r)
			c.Ctx.Output.SetStatus(500)
			c.Data["json"] = map[string]interface{}{
				"Success": false,
				"Status":  "500",
				"Message": "Error interno del servidor",
			}
			c.ServeJSON()
		}
	}()

	cedula := c.GetString(":cedula")

	if cedula == "" {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "400",
			"Message": "cedula es obligatoria",
		}
		c.ServeJSON()
		return
	}

	data, err := services.ObtenerBandejaDocente(cedula)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "500",
			"Message": "Error obteniendo comisiones del docente",
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

// GetComisionesDecano ...
// @Title Get Comisiones Decano
// @Description Retorna las comisiones de las facultades asignadas al decano, segun su cedula y datos del JBPM.
// @Param	cedula	path	string	true	"Numero de cedula del decano"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @router /comisiones_decano/:cedula [get]
func (c *ComisionSeguimientoController) GetComisionesDecano() {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("[ComisionSeguimiento] panic en GetComisionesDecano: %v", r)
			c.Ctx.Output.SetStatus(500)
			c.Data["json"] = map[string]interface{}{
				"Success": false,
				"Status":  "500",
				"Message": "Error interno del servidor",
			}
			c.ServeJSON()
		}
	}()

	cedula := c.GetString(":cedula")

	if cedula == "" {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "400",
			"Message": "cedula es obligatoria",
		}
		c.ServeJSON()
		return
	}

	data, err := services.ObtenerBandejaDecano(cedula)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "500",
			"Message": "Error obteniendo comisiones del decano",
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
