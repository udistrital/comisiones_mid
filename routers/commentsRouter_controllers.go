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
			Method:           "CancelarSolicitud",
			Router:           "/cancelar/:id",
			AllowHTTPMethods: []string{"put"},
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
			Method:           "DetallesSolicitud",
			Router:           "/detalles_solicitud/:id",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"],
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
			AllowHTTPMethods: []string{"post"},
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

	beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudPendienteCoordinadorController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudPendienteCoordinadorController"],
		beego.ControllerComments{
			Method:           "GetSolicitudesPendientesCoordinador",
			Router:           "/pendientes_coordinador/:numero_identificacion",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudPendienteDecanoController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudPendienteDecanoController"],
		beego.ControllerComments{
			Method:           "GetSolicitudesPendientesDecano",
			Router:           "/pendientes_decano/:numero_identificacion",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudPendienteSecretariaController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudPendienteSecretariaController"],
		beego.ControllerComments{
			Method:           "GetSolicitudesPendientesSecretaria",
			Router:           "/pendientes_secretaria/:numero_identificacion",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
