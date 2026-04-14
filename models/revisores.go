package models

import "encoding/xml"

type CoordinadoresXML struct {
	XMLName       xml.Name         `xml:"coordinadores"`
	Coordinadores []CoordinadorXML `xml:"coordinador"`
}

type CoordinadorXML struct {
	NombreCoordinador string `xml:"coordinador"`
	CodigoCarrera     string `xml:"codigo_carrera"`
	Identificacion    string `xml:"identificacion"`
	NombreCarrera     string `xml:"nombre_carrera"`
}

type SecretariaXML struct {
	XMLName xml.Name     `xml:"secretaria"`
	Persona []PersonaXML `xml:"persona"`
}

type PersonaXML struct {
	Apellidos         string `xml:"apellidos"`
	Estado            string `xml:"estado"`
	Identificacion    string `xml:"identificacion"`
	Dependencia       string `xml:"dependencia"`
	CodigoDependencia string `xml:"codigo_dependencia"`
	Nombres           string `xml:"nombres"`
}

type DecanosXML struct {
	XMLName xml.Name    `xml:"facultad"`
	Decanos []DecanoXML `xml:"decano"`
}

type DecanoXML struct {
	FechaDesde     string `xml:"fecha_desde"`
	CodigoFacultad string `xml:"codigo_facultad"`
	Nombre         string `xml:"nombre"`
	FechaHasta     string `xml:"fecha_hasta"`
	Facultad       string `xml:"facultad"`
}
