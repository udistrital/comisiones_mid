package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           "/",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"],
		beego.ControllerComments{
			Method:           "GetOne",
			Router:           "/:id",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"],
		beego.ControllerComments{
			Method:           "Put",
			Router:           "/:id",
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           "/:id",
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"],
		beego.ControllerComments{
			Method:           "CrearSolicitud",
			Router:           "/crear_solicitud",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"],
		beego.ControllerComments{
			Method:           "PruebaDocumento",
			Router:           "/prueba_documento",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:UserController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:UserController"],
		beego.ControllerComments{
			Method:           "PostEstados",
			Router:           "/estados",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"],
		beego.ControllerComments{
			Method:           "PruebaDocumento",
			Router:           "/prueba_documento",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"],
		beego.ControllerComments{
			Method:           "SolicitudByIdentificacion",
			Router:           "/solicitudes_by_identificacion/:id",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
