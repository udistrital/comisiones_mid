package controllers

import (
	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/services"
)

type SolicitudPendienteDecanoController struct {
	beego.Controller
}

// GetSolicitudesPendientesDecano ...
// @Title Get Solicitudes Pendientes Decano
// @Description Retorna las solicitudes pendientes por revisar de un decano según su número de identificación
// @Param	numero_identificacion	path 	string	true	"Número de identificación del decano"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @router /pendientes_decano/:numero_identificacion [get]
func (c *SolicitudPendienteDecanoController) GetSolicitudesPendientesDecano() {
	numeroIdentificacion := c.GetString(":numero_identificacion")

	if numeroIdentificacion == "" {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "400",
			"Message": "numero_identificacion es obligatorio",
		}
		c.ServeJSON()
		return
	}

	data, err := services.ObtenerSolicitudesPendientesDecano(numeroIdentificacion)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "500",
			"Message": "Error obteniendo solicitudes pendientes del decano",
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
