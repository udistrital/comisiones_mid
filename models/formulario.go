package models

type Formulario struct {
	Beca                 FormularioBeca                 `json:"beca"`
	Solicitud            FormularioSolicitud            `json:"solicitud"`
	Solicitante          FormularioSolicitante          `json:"solicitante"`
	FinanciacionColombia FormularioFinanciacionColombia `json:"financiacion_colombia"`
	FinanciacionExterior FormularioFinanciacionExterior `json:"financiacion_exterior"`
	FormularioCompletado bool                           `json:"formulario_completado"`
}

type FormularioBeca struct {
	Q40CubrimientoBeca       string `json:"q40_cubrimiento_beca"`
	Q41InstitucionOtorga     string `json:"q41_institucion_otorga"`
	Q42TipoFinanciacionMonto string `json:"q42_tipo_financiacion_monto"`
	Q43DuracionBeca          string `json:"q43_duracion_beca"`
}

type FormularioSolicitante struct {
	Q1Fecha                    string `json:"q1_fecha"`
	Q2Facultad                 string `json:"q2_facultad"`
	Q3NombresApellidos         string `json:"q3_nombres_apellidos"`
	Q4DocumentoIdentificacion  string `json:"q4_documento_identificacion"`
	Q5Edad                     string `json:"q5_edad"`
	Q6Correo                   string `json:"q6_correo"`
	Q7Proyecto                 string `json:"q7_proyecto"`
	Q8Telefono                 string `json:"q8_telefono"`
	Q9Celular                  string `json:"q9_celular"`
	Q10ResolucionRh            string `json:"q10_resolucion_rh"`
	Q10FechaIngresoUniversidad string `json:"q10_fecha_ingreso_universidad"`
	Q11CategoriaIngreso        string `json:"q11_categoria_ingreso"`
	Q12CategoriaActual         string `json:"q12_categoria_actual"`
}

type FormularioSolicitud struct {
	Q13TipoEstudio                interface{}   `json:"q13_tipo_estudio"`
	Q14NombrePrograma             string        `json:"q14_nombre_programa"`
	Q15TituloAspira               string        `json:"q15_titulo_aspira"`
	Q16Universidad                string        `json:"q16_universidad"`
	Q17Pais                       string        `json:"q17_pais"`
	Q18Ciudad                     string        `json:"q18_ciudad"`
	Q19FechaAceptacion            string        `json:"q19_fecha_aceptacion"`
	Q20NumSemestres               string        `json:"q20_num_semestres"`
	Q22TipoApoyoRequerido         []interface{} `json:"q22_tipo_apoyo_requerido"`
	Q23FechaInicioEstudios        string        `json:"q23_fecha_inicio_estudios"`
	Q24FechaCulminacionEstudios   string        `json:"q24_fecha_culminacion_estudios"`
	Q25TiempoRequeridoCulminacion string        `json:"q25_tiempo_requerido_culminacion"`
	Q26CostoTotalRequerido        string        `json:"q26_costo_total_requerido"`
}

type FormularioFinanciacionColombia struct {
	Q27PagoMatriculaValor          string `json:"q27_pago_matricula_valor"`
	Q28PagoMatriculaTotal          string `json:"q28_pago_matricula_total"`
	Q29Tiquetes                    string `json:"q29_tiquetes"`
	Q30DescargaAcademicaHoras      string `json:"q30_descarga_academica_horas"`
	Q31DescargaAcademicaValorTotal string `json:"q31_descarga_academica_valor_total"`
	Q32CostoReemplazoDocente       string `json:"q32_costo_reemplazo_docente"`
}

type FormularioFinanciacionExterior struct {
	Q33ValorSalarioTiempoComision string `json:"q33_valor_salario_tiempo_comision"`
	Q34PagoMatriculaValor         string `json:"q34_pago_matricula_valor"`
	Q35PagoTotalMatricula         string `json:"q35_pago_total_matricula"`
	Q36Tiquetes                   string `json:"q36_tiquetes"`
	Q37SeguroMedico               string `json:"q37_seguro_medico"`
	Q38GastosInstalacion          string `json:"q38_gastos_instalacion"`
	Q39CostoReemplazoDocente      string `json:"q39_costo_reemplazo_docente"`
}
