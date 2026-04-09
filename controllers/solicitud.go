package controllers

import (
	"encoding/json"

	"fmt"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/comisiones_mid/helpers"
	"github.com/udistrital/comisiones_mid/models"
	"github.com/udistrital/comisiones_mid/services"
)

// SolicitudController operations for Solicitud
type SolicitudController struct {
	beego.Controller
}

// URLMapping ...
func (c *SolicitudController) URLMapping() {
	c.Mapping("crear_solicitud", c.CrearSolicitud)
	c.Mapping("prueba_documento", c.PruebaDocumento)
	c.Mapping("solicitudes_by_identificacion", c.SolicitudByIdentificacion)
	c.Mapping("detalles_solicitud", c.DetallesSolicitud)
	c.Mapping("Post", c.Post)
	c.Mapping("GetOne", c.GetOne)
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("Put", c.Put)
	c.Mapping("Delete", c.Delete)
}

// Crear Solicitud ...
// @Title Create
// @Description create Solicitud
// @Param	body		body 	models.Solicitud	true		"body for Solicitud content"
// @Success 201 {object} models.Solicitud
// @Failure 403 body is empty
// @router /crear_solicitud [post]
func (c *SolicitudController) CrearSolicitud() {

	// 🔥 DEFER SEGURO
	defer func() {
		if r := recover(); r != nil {

			logs.Error("PANIC RECOVER:", r)

			response := map[string]interface{}{
				"Success": false,
				"Status":  500,
				"Message": "Error interno del servidor",
				"Data":    nil,
			}
			if errMap, ok := r.(map[string]interface{}); ok {

				if s, ok := errMap["Status"].(int); ok {
					response["Status"] = s
				} else if s, ok := errMap["status"].(string); ok {
					if parsed, err := strconv.Atoi(s); err == nil {
						response["Status"] = parsed
					}
				}

				if msg, ok := errMap["Message"].(string); ok && msg != "" {
					response["Message"] = msg
				} else if msg, ok := errMap["error"].(string); ok {
					response["Message"] = msg
				}

				if data, ok := errMap["Data"]; ok {
					response["Data"] = data
				}
			}

			statusCode := 500
			if s, ok := response["Status"].(int); ok {
				statusCode = s
			}

			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = response
			c.ServeJSON()
		}
	}()

	var v models.CrearSolicitudEntrada
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

	response, apiError := services.CrearSolicitud(v)

	if apiError != nil {

		statusCode := 500

		if s, ok := apiError["Status"].(int); ok {
			statusCode = s
		} else if s, ok := apiError["Status"].(string); ok {
			if parsed, err := strconv.Atoi(s); err == nil {
				statusCode = parsed
			}
		}

		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = apiError
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = map[string]interface{}{
		"Success": true,
		"Status":  201,
		"Message": "Solicitud creada correctamente",
		"Data":    response,
	}
	c.ServeJSON()
}

// TestDocumento ...
// @Title Test Documento
// @Description prueba creación documento
// @Success 200 {object} map[string]interface{}
// @router /prueba_documento [post]
func (c *SolicitudController) PruebaDocumento() {

	defer func() {
		if err := recover(); err != nil {
			logs.Error(err)
			localError := err.(map[string]interface{})
			c.Data["mesaage"] = (beego.AppConfig.String("appname") + "/" + "PruebaDocumento" + "/" + (localError["funcion"]).(string))
			c.Data["data"] = (localError["err"])
			if status, ok := localError["status"]; ok {
				c.Abort(status.(string))
			} else {
				c.Abort("404")
			}
		}
	}()
	var v []models.CrearDocumentoGestorDocumental
	//var v map[string]interface{}
	json.Unmarshal(c.Ctx.Input.RequestBody, &v)
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err == nil {
		if response, err := helpers.CrearDocumento(v); err == nil {
			c.Ctx.Output.SetStatus(201)
			c.Data["json"] = map[string]interface{}{"Success": true, "Status": "201", "Message": "Successful", "Data": response}
		} else {
			panic(err)
		}
	}
	c.ServeJSON()
}

// Post ...
// @Title Create
// @Description create Solicitud
// @Param	body		body 	models.Solicitud	true		"body for Solicitud content"
// @Success 201 {object} models.Solicitud
// @Failure 403 body is empty
// @router / [post]
func (c *SolicitudController) Post() {
	var req models.CambioEstadoSolicitudRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.CustomAbort(400, "JSON inválido: "+err.Error())
		return
	}

	if req.SolicitudId <= 0 {
		c.CustomAbort(400, "SolicitudId es obligatorio")
		return
	}

	resp, err := services.CambiarEstadoSolicitud(req.SolicitudId, req)
	if err != nil {
		c.CustomAbort(400, err.Error())
		return
	}

	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = map[string]interface{}{"Success": true, "Status": 201, "Message": "Estado actualizado correctamente", "Data": resp}
	c.ServeJSON()
}

// Buscar Solicitud por Identificacion...
// @Title Create
// @Description search Solicitud
// @Param	body		body 	models.Solicitud	true		"body for Solicitud content"
// @Success 201 {object} models.Solicitud
// @Failure 403 body is empty
// @router /solicitudes_by_identificacion/:id [get]
func (c *SolicitudController) SolicitudByIdentificacion() {
	defer func() {
		if err := recover(); err != nil {
			logs.Error(err)
			localError := err.(map[string]interface{})
			fmt.Println("ERROR ", localError)
			c.Data["mesaage"] = (beego.AppConfig.String("appname") + "/" + "SolicitudByIdentificacion" + "/" + (localError["error"]).(string))
			c.Data["data"] = (localError["err"])
			if status, ok := localError["status"]; ok {
				c.Abort(strconv.Itoa(status.(int)))
			} else {
				c.Abort("404")
			}
		}
	}()
	idStr := c.Ctx.Input.Param(":id")
	fmt.Println("ENTRA A BUSCAR")
	fmt.Println(idStr)
	id, err := strconv.Atoi(idStr)
	if err == nil {
		if response, err := services.BuscarSolicitudIdentificacion(id); err == nil {
			c.Ctx.Output.SetStatus(201)
			c.Data["json"] = map[string]interface{}{"Success": true, "Status": "201", "Message": "Successful", "Data": response}
		} else {
			panic(err)
		}
	} else {
		panic(err)
	}

	c.ServeJSON()
}

// Buscar Detalles de una Solicitud por Id...
// @Title Create
// @Description search Solicitud
// @Param	body		body 	models.Solicitud	true		"body for Solicitud content"
// @Success 201 {object} models.Solicitud
// @Failure 403 body is empty
// @router /detalles_solicitud/:id [get]
func (c *SolicitudController) DetallesSolicitud() {
	defer func() {
		if err := recover(); err != nil {
			logs.Error(err)
			localError := err.(map[string]interface{})
			fmt.Println("ERROR ", localError)
			c.Data["mesaage"] = (beego.AppConfig.String("appname") + "/" + "DetalleSolicitud" + "/" + (localError["error"]).(string))
			c.Data["data"] = (localError["err"])
			if status, ok := localError["status"]; ok {
				c.Abort(strconv.Itoa(status.(int)))
			} else {
				c.Abort("404")
			}
		}
	}()
	idStr := c.Ctx.Input.Param(":id")
	fmt.Println("ENTRA A BUSCAR")
	fmt.Println(idStr)
	id, err := strconv.Atoi(idStr)
	if err == nil {
		if response, err := services.BuscarDetallesSolicitud(id); err == nil {
			c.Ctx.Output.SetStatus(201)
			c.Data["json"] = map[string]interface{}{"Success": true, "Status": "201", "Message": "Successful", "Data": response}
		} else {
			panic(err)
		}
	} else {
		panic(err)
	}

	c.ServeJSON()
}

// GetOne ...
// @Title GetOne
// @Description get Solicitud by id
// @Param	id		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Solicitud
// @Failure 403 :id is empty
// @router /:id [get]
func (c *SolicitudController) GetOne() {

}

// GetAll ...
// @Title GetAll
// @Description get Solicitud
// @Param	query	query	string	false	"Filter. e.g. col1:v1,col2:v2 ..."
// @Param	fields	query	string	false	"Fields returned. e.g. col1,col2 ..."
// @Param	sortby	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	order	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	limit	query	string	false	"Limit the size of result set. Must be an integer"
// @Param	offset	query	string	false	"Start position of result set. Must be an integer"
// @Success 200 {object} models.Solicitud
// @Failure 403
// @router / [get]
func (c *SolicitudController) GetAll() {

}

// Put ...
// @Title Put
// @Description update the Solicitud
// @Param	id		path 	string	true		"The id you want to update"
// @Param	body		body 	models.Solicitud	true		"body for Solicitud content"
// @Success 200 {object} models.Solicitud
// @Failure 403 :id is not int
// @router /:id [put]
func (c *SolicitudController) Put() {
	idStr := c.Ctx.Input.Param(":id")
	solicitudId, err := strconv.Atoi(idStr)
	if err != nil || solicitudId <= 0 {
		c.CustomAbort(400, "id de solicitud inválido")
		return
	}

	var req models.EditarSolicitud
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.CustomAbort(400, "JSON inválido: "+err.Error())
		return
	}

	resp, err := services.EditarSolicitud(solicitudId, req)
	if err != nil {
		c.CustomAbort(400, err.Error())
		return
	}

	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = map[string]interface{}{
		"Success": true,
		"Status":  200,
		"Message": "Solicitud actualizada correctamente",
		"Data":    resp,
	}
	c.ServeJSON()
}

// Delete ...
// @Title Delete
// @Description delete the Solicitud
// @Param	id		path 	string	true		"The id you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 id is empty
// @router /:id [delete]
func (c *SolicitudController) Delete() {

}

// PostEstados ...
// @Title Crear estado de solicitud
// @Description Desactiva el histórico vigente (Activo=false) y crea un nuevo histórico con estado destino
// @Param   body   body   map[string]interface{}  true  "Body con SolicitudId + codigos + NumeroIdentificacion"
// @Success 201 {object} map[string]interface{}
// @Failure 400 bad request
// @router /estados [post]
func (c *SolicitudController) PostEstados() {

	var body map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &body); err != nil {
		c.CustomAbort(400, "JSON inválido: "+err.Error())
		return
	}

	rawSid, ok := body["SolicitudId"]
	if !ok {
		c.CustomAbort(400, "SolicitudId es obligatorio")
		return
	}

	solicitudId, err := strconv.Atoi(fmt.Sprintf("%v", rawSid))
	if err != nil || solicitudId <= 0 {
		c.CustomAbort(400, "SolicitudId inválido")
		return
	}

	var req models.CambioEstadoSolicitudRequest
	b, _ := json.Marshal(body)
	if err := json.Unmarshal(b, &req); err != nil {
		c.CustomAbort(400, "Body inválido: "+err.Error())
		return
	}

	resp, err := services.CambiarEstadoSolicitud(solicitudId, req)
	if err != nil {
		c.CustomAbort(400, err.Error())
		return
	}

	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = map[string]interface{}{
		"Success": true,
		"Status":  201,
		"Message": "Estado actualizado correctamente",
		"Data":    resp,
	}
	c.ServeJSON()
}

// CancelarSolicitud ...
// @Title Cancelación de la solicitud
// @Description Desactiva la solicitud y todos sus registros asociados activos
// @Param   id   path   int  true  "Id de la solicitud"
// @Success 200 {object} map[string]interface{}
// @Failure 400 bad request
// @router /cancelar/:id [put]
func (c *SolicitudController) CancelarSolicitud() {
	idStr := c.Ctx.Input.Param(":id")
	solicitudId, err := strconv.Atoi(idStr)
	if err != nil || solicitudId <= 0 {
		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "400",
			"Message": "id de solicitud inválido",
		}
		c.Ctx.Output.SetStatus(400)
		c.ServeJSON()
		return
	}

	resultado, err := services.CancelarSolicitud(solicitudId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"Success": false,
			"Status":  "500",
			"Message": "error cancelando solicitud",
			"Error":   err.Error(),
		}
		c.Ctx.Output.SetStatus(500)
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"Success": true,
		"Status":  "200",
		"Message": "solicitud cancelada correctamente",
		"Data":    resultado,
	}
	c.Ctx.Output.SetStatus(200)
	c.ServeJSON()
}
