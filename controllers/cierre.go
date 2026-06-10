package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/comisiones_mid/services"
)

const consultaExitosa = "Consulta exitosa"
const errJsonInvalido = "JSON inválido"

// ProrrogaController operations for Cierre
type CierreController struct {
	beego.Controller
}

// URLMapping ...
func (c *CierreController) URLMapping() {
	c.Mapping("CrearSolicitudCierre", c.CrearSolicitudCierre)
	c.Mapping("ValidarSolicitudCierre", c.ValidarSolicitudCierre)
	c.Mapping("ConsultarHistoricoSolicitudesCierre", c.ConsultarHistoricoSolicitudesCierre)
	c.Mapping("RechazarSolicitudCierre", c.RechazarSolicitudCierre)
	c.Mapping("AprobarSolicitudCierre", c.AprobarSolicitudCierre)
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
			"Message": errJsonInvalido,
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
		"Message": consultaExitosa,
		"Data":    data,
	}
	c.ServeJSON()
}

// ======================================================
// VALIDAR SI PUEDE CREAR CIERRE
// ======================================================

// Get ...
// @Title Validar Solicitud Cierre
// @Description valida si una comisión puede crear una nueva solicitud de cierre
// @Param	id	path	int	true		"Id comisión"
// @Success 200 {object} map[string]interface{}
// @Failure 400 id inválido
// @Failure 500 error interno
// @router /validar_solicitud_cierre/:id [get]
func (c *CierreController) ValidarSolicitudCierre() {

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

	puedeCrear, mensaje, err := services.ValidarPuedeCrearSolicitudCierre(
		int(comisionId),
	)

	if err != nil {

		c.Ctx.Output.SetStatus(500)

		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "500",
			"Message": "Error validando solicitud de cierre",
			"Error":   err.Error(),
		}

		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(200)

	c.Data["json"] = map[string]interface{}{
		"Success": true,
		"Status":  "200",
		"Message": consultaExitosa,
		"Data": map[string]interface{}{
			"puede_crear_cierre": puedeCrear,
			"mensaje":            mensaje,
		},
	}

	c.ServeJSON()
}

// @Title ConsultarHistoricoSolicitudesCierre
// @Description Consulta el histórico de solicitudes de cierre
// @Success 200 {object} models.ResponseHistoricoSolicitudCierre
// @Failure 400 id inválido
// @Failure 500 error interno
// @router /historico_solicitudes_cierre/:id [get]
func (c *CierreController) ConsultarHistoricoSolicitudesCierre() {

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

	data, err := services.ConsultarHistoricoSolicitudesCierre(comisionId)

	if err != nil {

		c.Ctx.Output.SetStatus(500)

		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "500",
			"Message": "Error consultando histórico de solicitudes de cierre",
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
		"Message": consultaExitosa,
		"Data":    data,
	}

	c.ServeJSON()
}

// @Title RechazarSolicitudCierre
// @Description Rechaza la solicitud de cierre
// @Success 200 {object} models.ResponseRechazarSolicitudCierre
// @Failure 400 id inválido
// @Failure 500 error interno
// @router /rechazar_cierre [post]
func (c *CierreController) RechazarSolicitudCierre() {

	fmt.Println("ENTRA A CREAR CIERRE")
	var v models.CierreAprobacionSolicitud
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err != nil {

		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  400,
			"Message": errJsonInvalido,
			"Data":    nil,
		}
		c.ServeJSON()
		return
	}

	data, err := services.RechazarSolicitudCierre(v)

	if err != nil {

		c.Ctx.Output.SetStatus(500)

		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "500",
			"Message": "Error rechazando la solicitud de cierre",
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
		"Message": consultaExitosa,
		"Data":    data,
	}

	c.ServeJSON()
}

// @Title AprobarSolicitudCierre
// @Description Aprueba la solicitud de cierre
// @Success 200 {object} models.ResponseAprobarSolicitudCierre
// @Failure 400 id inválido
// @Failure 500 error interno
// @router /aprobar_cierre [post]
func (c *CierreController) AprobarSolicitudCierre() {

	fmt.Println("ENTRA A CREAR CIERRE")
	var v models.CierreAprobacionSolicitud
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err != nil {

		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  400,
			"Message": errJsonInvalido,
			"Data":    nil,
		}
		c.ServeJSON()
		return
	}

	data, err := services.AprobarSolicitudCierre(v)

	if err != nil {

		c.Ctx.Output.SetStatus(500)

		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "500",
			"Message": "Error rechazando la solicitud de cierre",
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
		"Message": consultaExitosa,
		"Data":    data,
	}

	c.ServeJSON()
}
