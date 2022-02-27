package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
	// se utiliza para obtener el tipo de una variable
)

type dataWeb struct {
	index    int
	city     string
	location string
	weather  string
}

type dataBad struct {
	index int
	city  string
}

//--MAIN---------------------------------------------------------------------------

func main() {

	// Inicia la medicion del tiempo de ejecucion
	now := time.Now()

	// 1. Se leen y procesan los parametros de entrada.
	// Se ingresan los args es un slice de strings
	args := os.Args

	// Si no hay argumentos finaliza el programa.
	if len(args[1:]) == 0 {
		fmt.Println("Se debe ingresar al menos un  nombre de Ciudad...")
		return
	}

	// 2. Se verifica que los nombres de las ciudades
	// sean del formato ciudad,PA donde PA son las siglas del pais.

	//crear una variable global que almacenará la expresión regular
	expresionRegular := regexp.MustCompile("^[A-Z][A-Za-z]+[,]+([A-Z][A-Z])?$")

	// Estas variables contienen los arreglos con parametros buenos y malos
	var dataCity []dataWeb // Ciudades con formato OK
	var badCity []dataBad  // Ciudades con formato no OK

	for idx, arg := range args[1:] {
		// range va desde 0 hasta el tamanio del arreglo
		if expresionRegular.Match([]byte(arg)) { // Si coincide con la expresion regular es TRUE
			// 3. Se guardan los parametros ok en un arreglos de ciudades
			dataCity = append(dataCity, dataWeb{idx + 1, arg, "vacio", "vacio"})
		} else {
			// Se guardan las ciudades con formato incorrecto
			// en un arreglo de malas ciudades
			badCity = append(badCity, dataBad{idx + 1, arg})
		}

	}
	// Se muestra un mensaje del numero de argumentos con formato incorrecto
	if len(badCity) > 0 {
		fmt.Println("\n\nAlgunos argumentos (", len(badCity), ") tienen formato incorrecto...")
	}

	respuesta := menu(badCity) // Se envia badCity para contar los malos

	// La unica respuesta que se procesa aqui es la 1
	// las otras se hacen en  menu()

	if respuesta == 1 {
		// En este lazo se envian una a una  las ciudades por el
		// canal.
		//for idx, arg := range args[1:] {
		for i := 0; i <= len(dataCity)-1; i++ {
			//fmt.Println("Procesando elemento [", i, "]")
			proCitys(&dataCity[i]) // Se llama a la funcion para procesar Ciudades
		}
	}

	// Se imprimen los resultados obtenidos
	impWeather(dataCity)

	fmt.Println("\n\n FIN DEL PROGRAMA..", string(time.Since(now)))
	fmt.Println("\n\nTiempo transcurrido:", time.Since(now))
}

//--IMPWEATHER---------------------------------------------------------------------------
func impWeather(cityes []dataWeb) {

	// Se ordena el Slice por el nombre de la ciudad.
	sort.Slice(cityes, func(i, j int) bool {
		return cityes[i].city < cityes[j].city
	})

	for i := 0; i < len(cityes); i++ {
		fmt.Println("No:", i+1)
		fmt.Println("Ciudad: ", cityes[i].city)
		fmt.Println("Info Geografica: ", cityes[i].location)
		fmt.Println("Info Clima: ", cityes[i].weather)
		fmt.Println("")

	}

}

//--MENU---------------------------------------------------------------------------

// Esta funcion muestra el menu principal.
// Retorna el numero de la opcion seleccionada

func menu(badCity []dataBad) int {

	fmt.Println("\n*** Menu ***")
	fmt.Println("------------")
	fmt.Println("0: No procesar ningun parametro.")
	fmt.Println("1: Procesar solamente parametros validos.")
	if len(badCity) > 0 {
		fmt.Println("2: Mostrar no validos")
	}
	fmt.Println("\nSeleccione opcion o <Enter> para terminar...")
	reader := bufio.NewReader(os.Stdin)          // se lee el numero desde el teclado
	entrada, _ := reader.ReadString('\n')        // Leer hasta el separador de salto de línea
	opcion := strings.TrimRight(entrada, "\r\n") // Remover el salto de línea de la entrada del usuario

	switch opcion {
	case "0":
		fmt.Println("No procesar ninguno de los parametros..")
		return 0
	case "1":
		fmt.Println("Procesando parametros validos...")
		return 1
	case "2":
		fmt.Println("\nListado de parametros no validos...")
		fmt.Println(badCity)
		fmt.Println("\nCorrija estos parametros en la linea de comandos.")
		fmt.Println("La primera letra del nombre de la ciudad en MAYUSCULA.")
		fmt.Println("Las dos letras del pais en MAYUSCULAS.")
		return 2

	default:
		fmt.Println("Saliendo del sistema")
		return 0
	}

}

// --PROCITYS---------------------------------------------------------------------------------------

func proCitys(city *dataWeb) {
	// Procesa las ciudades de una en una.

	//fmt.Println("Procesando en proCitys: ", city.index, " - ", city.city)

	var bytes []byte // bytes contiene el texto de la consulta web
	// El parametro es el nombre de la ciudad en formato ciudad,PA
	bytes, err := queryCityLocation(city.city)
	// manejo de error
	if err != nil {
		fmt.Println("Error fatal #1 en proCitys")
		log.Fatal(err)
	}
	// Si la cadena devuelta es [], significa que no se pudo hacer la consulta
	if len(bytes) < 3 {
		fmt.Println(fmt.Sprintf("La ciudad: [%s] no existe o el nombre esta mal escrito.", city.city))
		city.location = string(bytes) // Se almacena lo recuperado
		return                        // se sale de la funcion ya que no hay contenido
	}
	// Se descompone el nombre en sus dos partes ciudad,PA ciudad y pais.
	//  Se obtiene las posiciones donde se encuentra (,)

	iniCadena := strings.Index(city.city, ",")

	// Se obtienen las cadenas de la ciudad y el pais
	strCity := city.city[:iniCadena]
	strCountry := city.city[iniCadena+1:]
	// Buscamos los valores para: name, country, state, lat y lon
	// para hacer esto utilizamos expresiones regulares
	// STATE
	expresion := regexp.MustCompile(`\"state\":\"[a-zA-Z]*\"`)
	datFound := string(expresion.Find(bytes))
	iniCadena = strings.Index(datFound, ":")
	strState := datFound[iniCadena+2 : len(datFound)-1]
	// LAT
	expresion = regexp.MustCompile(`\"lat\":\-?[0-9]+\.?[0-9]+`)
	datFound = string(expresion.Find(bytes))
	iniCadena = strings.Index(datFound, ":")
	strLat := datFound[iniCadena+1:]
	// LON
	expresion = regexp.MustCompile(`\"lon\":\-?[0-9]+\.?[0-9]+`)
	datFound = string(expresion.Find(bytes))
	iniCadena = strings.Index(datFound, ":")
	strLon := datFound[iniCadena+1:]

	txtResultado := fmt.Sprintf("[Ciudad: %s Estado: %s Pais: %s Latitud: %s Longitud: %s]", strCity, strState, strCountry, strLat, strLon)

	city.location = txtResultado // Se introduce el resultado para la locacion

	// Se procede a obtener el tiempo con las coordenadas de
	// latitud y logitud obtenidas
	bytes, err = queryCityWeather(strLat, strLon)

	// manejo de error
	if err != nil {
		fmt.Println("Error fatal #1 en proCitys")
		log.Fatal(err)
	}
	// Si la cadena devuelta es [], significa que no se pudo hacer la consulta
	if len(bytes) < 3 {
		fmt.Println(fmt.Sprintf("La ciudad: [%s] no existe o las coordenadas (Lat:%s, Lon:%s) estan equivocadas.", city.city, strLat, strLon))
		city.location = string(bytes) // Se almacena lo recuperado
		return                        // se sale de la funcion ya que no hay contenido
	}

	//fmt.Println(string(bytes))
	city.weather = paramWeather(bytes)

}

// --PARAMWEATHER---------------------------------------------------------------------------------------
func paramWeather(bytes []byte) string {
	// Del resultado obtenido vamos a obtener los parametros
	// que se pueden extraer con la misma expresion regular
	// TEMP
	expresion := regexp.MustCompile(`\"temp\":\-?[0-9]+\.?[0-9]+`)
	datFound := string(expresion.Find(bytes))
	iniCadena := strings.Index(datFound, ":")
	strTemp := datFound[iniCadena+1:]
	// TEMP MIN
	expresion = regexp.MustCompile(`\"temp_min\":\-?[0-9]+\.?[0-9]+`)
	datFound = string(expresion.Find(bytes))
	iniCadena = strings.Index(datFound, ":")
	strTempMin := datFound[iniCadena+1:]
	// TEMP MAX
	expresion = regexp.MustCompile(`\"temp_max\":\-?[0-9]+\.?[0-9]+`)
	datFound = string(expresion.Find(bytes))
	iniCadena = strings.Index(datFound, ":")
	strTempMax := datFound[iniCadena+1:]
	// PRESION
	expresion = regexp.MustCompile(`\"pressure\":\-?[0-9]+\.?[0-9]+`)
	datFound = string(expresion.Find(bytes))
	iniCadena = strings.Index(datFound, ":")
	strPres := datFound[iniCadena+1:]
	// HUMEDAD
	expresion = regexp.MustCompile(`\"humidity\":\-?[0-9]+\.?[0-9]+`)
	datFound = string(expresion.Find(bytes))
	iniCadena = strings.Index(datFound, ":")
	strHum := datFound[iniCadena+1:]
	// VISIBILIDAD
	expresion = regexp.MustCompile(`\"visibility\":\-?[0-9]+\.?[0-9]+`)
	datFound = string(expresion.Find(bytes))
	iniCadena = strings.Index(datFound, ":")
	strVisi := datFound[iniCadena+1:]
	// VIENTO
	expresion = regexp.MustCompile(`\"speed\":\-?[0-9]+\.?[0-9]+`)
	datFound = string(expresion.Find(bytes))
	iniCadena = strings.Index(datFound, ":")
	strWindSpeed := datFound[iniCadena+1:]

	txtResultado := fmt.Sprintf("[Temperatura: %s Temp Minima: %s Temp Maxima: %s Presion: %s Humedad: %s Visibilidad: %s Vel Viento: %s]", strTemp, strTempMin, strTempMax, strPres, strHum, strVisi, strWindSpeed)

	return txtResultado
}

// --QUERYCITYLOCATION---------------------------------------------------------------------------------------

func queryCityLocation(city string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	// codigo nuevo
	urlFind := fmt.Sprintf("https://api.openweathermap.org/geo/1.0/direct?q=%s&limit=5&appid=3582889a6bd6c9bd3d2867e51f420116", city)
	bytes, err := getWebBytes(client, urlFind)
	// manejo de error
	if err != nil {
		fmt.Println("Error fatal en : queryCityWeather()")
		log.Fatal(err)
	}
	return bytes, nil
}

// --QUERYCITYWEATHER---------------------------------------------------------------------------------------

func queryCityWeather(latCity string, lonCity string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	// codigo nuevo
	urlFind := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?units=metric&lat=%s&lon=%s&appid=3582889a6bd6c9bd3d2867e51f420116", latCity, lonCity)

	bytes, err := getWebBytes(client, urlFind)
	// manejo de error
	if err != nil {
		fmt.Println("Error fatal en : queryCityWeather()")
		log.Fatal(err)
	}
	return bytes, nil
}

//--GETWEBBYTES---------------------------------------------------------------------------------------
//
// Esta funcion obtiene el contenido del sitio web, de acuerdo
// al URL que se le de como parametro.

func getWebBytes(client *http.Client, url string) ([]byte, error) {
	// Se verifica si la solicitud se ha construido bien.
	// Si se tiene un error en la conexion o la consulta se termina el programa
	// y se muestra el mensaje de error generado.

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		fmt.Println("Error fatal #1 en getWebBytes")
		log.Fatal(err)
	}
	// Ejecutar la consulta. El método Do ejecuta la soliciutd.
	res, err := client.Do(req)
	// Si se tiene un error en la conexion o la consulta se termina el programa
	// y se muestra el mensaje de error generado.
	if err != nil {
		fmt.Println("Error fatal #2 en getWebBytes")
		log.Fatal(err)
	}
	// La siguiente funcion lee todo el "body" de la pagina web.
	bytes, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}
	return bytes, nil

} // end getWebBytes()
