package controllers

import (
	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/services"
)

type SolicitudPendienteSecretariaController struct {
	beego.Controller
}

// GetSolicitudesPendientesSecretaria ...
// @Title Get Solicitudes Pendientes Secretaria
// @Description Retorna las solicitudes pendientes por revisar de la secretaria academica según su número de identificación
// @Param	numero_identificacion	path 	string	true	"Número de identificación del secretario(a)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @router /pendientes_secretaria/:numero_identificacion [get]
func (c *SolicitudPendienteSecretariaController) GetSolicitudesPendientesSecretaria() {
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

	data, err := services.ObtenerSolicitudesPendientesSecretaria(numeroIdentificacion)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "500",
			"Message": "Error obteniendo solicitudes pendientes de la secretaria",
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
