# comisiones_mid

✔️ Check: API MID para la gestión e integración del sistema de Comisiones de la Universidad Distrital Francisco José de Caldas.

## Especificaciones Técnicas

## Especificaciones Técnicas

### Tecnologías Implementadas y Versiones
* [Golang](https://github.com/udistrital/introduccion_oas/blob/master/instalacion_de_herramientas/golang.md)
* [BeeGo](https://github.com/udistrital/introduccion_oas/blob/master/instalacion_de_herramientas/beego.md)
* [Docker](https://docs.docker.com/engine/install/ubuntu/)
* [Docker Compose](https://docs.docker.com/compose/)

## Variables de Entorno

## Variables de Entorno

```bash
# Parámetros del API
COMISIONES_MID_HTTPPORT=[Puerto de exposición del API]
COMISIONES_MID_RUNMODE=[Modo de ejecución (dev, test, prod)]

# Parámetros de configuración
PARAMETER_STORE=[Ruta del Parameter Store]

# Servicios externos
COMISIONES_MID_TERCEROS=[URL del servicio terceros_crud]
COMISIONES_MID_COMISIONES_CRUD=[URL del servicio comisiones_crud]
COMISIONES_MID_DOCUMENTOS_CRUD=[URL del servicio documentos_crud]
COMISIONES_MID_GESTORDOCUMENTAL=[URL del servicio gestor_documental]
COMISIONES_MID_ACADEMICA_JBPM=[URL del servicio academica_jbpm]
```

**NOTA:** Las variables se pueden consultar en el archivo `conf/app.conf` y están identificadas de acuerdo con los lineamientos institucionales para la definición de variables de entorno. Todas deben configurarse antes de la ejecución del servicio.


## Ejecución del Proyecto

### 1. Obtener el repositorio

```bash
go get github.com/udistrital/comisiones_mid
```

### 2. Ubicarse en la carpeta del proyecto

```bash
cd $GOPATH/src/github.com/udistrital/comisiones_mid
```

### 3. Cambiar a la rama de desarrollo

```bash
git pull origin develop
git checkout develop
```

### 4. Configurar variables de entorno y ejecutar

```bash
COMISIONES_MID_HTTP_PORT=8080 bee run
```

---

## Ejecución mediante Docker

### Construcción de imagen

```bash
docker build -t comisiones_mid .
```

### Ejecución de contenedor

```bash
docker run -p 8080:8080 comisiones_mid
```

---

## Ejecución mediante Docker Compose

### 1. Clonar el repositorio

```bash
git clone -b develop https://github.com/udistrital/comisiones_mid.git
```

### 2. Ingresar al directorio

```bash
cd comisiones_mid
```

### 3. Crear archivo de variables personalizadas

```bash
touch custom.env
```

### 4. Crear red de contenedores

```bash
docker network create back_end
```

### 5. Levantar servicios

```bash
docker-compose up --build
```

### 6. Verificar ejecución

```bash
docker ps
```

---

## Estructura General

```text
comisiones_mid/
├── conf/
├── controllers/
├── models/
├── routers/
├── helpers/
├── services/
├── utils/
├── swagger/
└── main.go
```

---

## Documentación API

Una vez ejecutado el servicio, la documentación Swagger estará disponible en:

```text
http://localhost:8080/swagger/
```

---

## Pruebas

## Ejecución Pruebas

### Pruebas unitarias

Las pruebas automatizadas se encuentran en el directorio `tests/`.

Ejecutar todas las pruebas:

```bash
go test ./tests/...
```

### Cobertura de pruebas

```bash
go test ./tests/... -cover
```

### Reporte detallado

```bash
go test ./tests/... -v
```


## Estado CI


| Develop | Relese 0.0.1 | Master | Sonar |
| -- | -- | -- | -- |
| [![Build Status](https://hubci.portaloas.udistrital.edu.co/api/badges/udistrital/resoluciones_mid_v2/status.svg?ref=refs/heads/develop)](https://hubci.portaloas.udistrital.edu.co/udistrital/resoluciones_mid_v2) | [![Build Status](https://hubci.portaloas.udistrital.edu.co/api/badges/udistrital/resoluciones_mid_v2/status.svg?ref=refs/heads/release/0.0.1)](https://hubci.portaloas.udistrital.edu.co/udistrital/resoluciones_mid_v2) | [![Build Status](https://hubci.portaloas.udistrital.edu.co/api/badges/udistrital/resoluciones_mid_v2/status.svg)](https://hubci.portaloas.udistrital.edu.co/udistrital/resoluciones_mid_v2) | [![Quality Gate Status](https://sonarqube.portaloas.udistrital.edu.co/api/project_badges/measure?project=udistrital%3Aresoluciones_mid_v2&metric=alert_status)](https://sonarqube.portaloas.udistrital.edu.co/dashboard?id=udistrital%3Aresoluciones_mid_v2) |

---

## Licencia

This file is part of resoluciones_docentes_mid.

resoluciones_docentes_mid is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

resoluciones_docentes_mid is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with resoluciones_docentes_mid. If not, see https://www.gnu.org/licenses/.

