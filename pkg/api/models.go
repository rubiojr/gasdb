package api

// StationWithDistance associates a GasStation with a computed distance.
type StationWithDistance struct {
	Station  *GasStation
	Distance float64
}

// GasStationList represents the response structure from the fuel price API.
type GasStationList struct {
	Fecha             string       `json:"Fecha"`
	ListaEESSPrecio   []GasStation `json:"ListaEESSPrecio"`
	Nota              string       `json:"Nota"`
	ResultadoConsulta string       `json:"ResultadoConsulta"`
}

// GasStation represents a single fuel station and its price information.
type GasStation struct {
	CP                      string `json:"C.P."`
	Direccion               string `json:"Dirección"`
	Horario                 string `json:"Horario"`
	Latitud                 string `json:"Latitud"`
	Localidad               string `json:"Localidad"`
	Longitud                string `json:"Longitud (WGS84)"`
	Margen                  string `json:"Margen"`
	Municipio               string `json:"Municipio"`
	PrecioBiodiesel         string `json:"Precio Biodiesel"`
	PrecioBioetanol         string `json:"Precio Bioetanol"`
	PrecioGasNaturalComp    string `json:"Precio Gas Natural Comprimido"`
	PrecioGasNaturalLicuado string `json:"Precio Gas Natural Licuado"`
	PrecioGasesLicuados     string `json:"Precio Gases licuados del petróleo"`
	PrecioGasoleoA          string `json:"Precio Gasoleo A"`
	PrecioGasoleoB          string `json:"Precio Gasoleo B"`
	PrecioGasoleoPremium    string `json:"Precio Gasoleo Premium"`
	PrecioGasolina95E10     string `json:"Precio Gasolina 95 E10"`
	PrecioGasolina95E5      string `json:"Precio Gasolina 95 E5"`
	PrecioGasolina95E5Prem  string `json:"Precio Gasolina 95 E5 Premium"`
	PrecioGasolina98E10     string `json:"Precio Gasolina 98 E10"`
	PrecioGasolina98E5      string `json:"Precio Gasolina 98 E5"`
	PrecioHidrogeno         string `json:"Precio Hidrogeno"`
	Provincia               string `json:"Provincia"`
	Remision                string `json:"Remisión"`
	Rotulo                  string `json:"Rótulo"`
	TipoVenta               string `json:"Tipo Venta"`
	PorcentajeBioEtanol     string `json:"% BioEtanol"`
	PorcentajeEsterMetilico string `json:"% Éster metílico"`
	IDEESS                  string `json:"IDEESS"`
	IDMunicipio             string `json:"IDMunicipio"`
	IDProvincia             string `json:"IDProvincia"`
	IDCCAA                  string `json:"IDCCAA"`
}
