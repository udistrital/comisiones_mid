package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:SolicitudController"],
        beego.ControllerComments{
            Method: "CrearSolicitud",
            Router: "/crear_solicitud",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
