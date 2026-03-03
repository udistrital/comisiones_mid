package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:ComisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/comisiones_mid/controllers:ComisionController"],
        beego.ControllerComments{
            Method: "CrearComision",
            Router: "/CrearComision",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
