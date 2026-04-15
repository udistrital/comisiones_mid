package main

import (
	"github.com/astaxie/beego"
	_ "github.com/udistrital/comisiones_mid/routers"

	"github.com/astaxie/beego/plugins/cors"
	apistatus "github.com/udistrital/utils_oas/apiStatusLib"
	"github.com/udistrital/utils_oas/customerrorv2"
	"github.com/udistrital/utils_oas/xray"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"PUT", "PATCH", "GET", "POST", "OPTIONS", "DELETE"},
		AllowHeaders: []string{"Origin", "x-requested-with",
			"content-type",
			"accept",
			"origin",
			"authorization",
			"x-csrftoken",
			"pragma",
			"cache-control"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	xray.InitXRay()
	beego.ErrorController(&customerrorv2.CustomErrorController{})
	apistatus.Init()
	beego.Run()
}
