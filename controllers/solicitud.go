package controllers

import (
	"encoding/json"

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
}

// Post ...
// @Title Create
// @Description create Solicitud
// @Param	body		body 	models.Solicitud	true		"body for Solicitud content"
// @Success 201 {object} models.Solicitud
// @Failure 403 body is empty
// @router /crear_solicitud [post]
func (c *SolicitudController) CrearSolicitud() {
	defer func() {
		if err := recover(); err != nil {
			logs.Error(err)
			localError := err.(map[string]interface{})
			c.Data["mesaage"] = (beego.AppConfig.String("appname") + "/" + "CrearSolicitud" + "/" + (localError["funcion"]).(string))
			c.Data["data"] = (localError["err"])
			if status, ok := localError["status"]; ok {
				c.Abort(status.(string))
			} else {
				c.Abort("404")
			}
		}
	}()
	var v models.CrearSolicitudEntrada
	//var v map[string]interface{}
	json.Unmarshal(c.Ctx.Input.RequestBody, &v)
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err == nil{
		if response, err := services.CrearSolicitud(v); err == nil {
			c.Ctx.Output.SetStatus(201)
			c.Data["json"] = map[string]interface{}{"Success": true, "Status": "201", "Message": "Successful", "Data": response}
		} else {
			panic(err)
		}
	}
	c.ServeJSON()
}

// TestDocumento ...
// @Title Test Documento
// @Description prueba creación documento
// @Success 200 {object} map[string]interface{}
// @router /prueba_documento [get]
func (c *SolicitudController) PruebaDocumento() {

	id := helpers.CrearDocumento(models.Documento{
		Nombre:      	"documento prueba comisiones",
		Descripcion: 	"prueba para comisiones",
		Metadatos: 		"{\"NombreArchivo\":\"Resolucion_486.pdf\",\"FechaCreacion\":\"12_Dec_2018_13:47:57\",\"Tipo\":\"Archivo\",\"IdNuxeo\":\"b72eeb98-f3d1-4e07-afdd-f3ea0fa612f6\",\"Observaciones\":\"Ninguna\"}",
		TipoDocumento: &models.TipoDocumento{
			Id: 6,
		},
		Activo:      	true,
	})

	c.Data["json"] = map[string]interface{}{
		"Success": true,
		"Id":      id,
	}

	c.ServeJSON()
}