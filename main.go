package main

import (
	"bufio"
	"context"
	"encoding/json"
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
	index   int
	city    string
	state   string
	country string
	// location string // Se elimina
	latCity float32
	lonCity float32
	weather string
	// Se agregan estos dos campos para saber cuando se crea
	// y cuando se termina de llenar los datos.
	tcreate time.Time // cuando se crea el dato
	tdone   time.Time // cuando finaliza la creacion

}

type dataBad struct {
	index int
	city  string
}

// Esta estructura es para recuperar la
// UBICACION DE UNA CIUDAD

type Ubicacion struct {
	Name    string  `json:"name"`
	Lat     float32 `json:"lat"`
	Lon     float32 `json:"lon"`
	Country string  `json:"country"`
	State   string  `json:"state"`
}

//--MAIN---------------------------------------------------------------------------

func main() {

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
	// Se modifica la expresion regular para que acepte tildes el el nombre
	expresionRegular := regexp.MustCompile("^[A-Z][A-Za-záéíóú]+[,]+([A-Z][A-Z])?$")

	// Estas variables contienen los arreglos con parametros buenos y malos
	var dataCity []dataWeb // Ciudades con formato OK
	var badCity []dataBad  // Ciudades con formato no OK
	var cityEmpty []string // Aqui se almacenan los nombres de las ciudades
	// para las que no se recupera informacion

	for idx, arg := range args[1:] {
		// range va desde 0 hasta el tamanio del arreglo
		if expresionRegular.Match([]byte(arg)) { // Si coincide con la expresion regular es TRUE
			// Se descompone la ciudad en nombre y pais Quito , EC
			posComma := strings.Index(arg, ",")
			nameCity := arg[0:posComma]
			codCountry := arg[posComma+1:]
			// 3. Se guardan los parametros ok en un arreglos de ciudades
			dataCity = append(dataCity, dataWeb{idx + 1, nameCity, "vacio", codCountry, 0, 0, "vacio", time.Now(), time.Now()})
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

	// Inicia la medicion del tiempo de ejecucion luego
	// de que se selecciona la opcion.
	now := time.Now()

	if respuesta == 1 {
		// En este lazo se envian una a una  las ciudades por el
		// canal.
		//for idx, arg := range args[1:] {
		for i := 0; i <= len(dataCity)-1; i++ {
			// Se llama a la funcion para procesar Ciudades
			// Se procesa una ciudad en cada llamada.
			go proCitysLoc(&dataCity[i])
			time.Sleep(90 * time.Millisecond)
		}
		// originalmente en una sola llamada se hacian
		// las consultas para obtener ubicacion y clima.
		//
		// en la llamada anterior solamente se obtiene
		// la latitud y la longitud
		//
		// en la siguiente llamada se obtiene la longitud y la latitud
		// de la ciudad

		for i := 0; i <= len(dataCity)-1; i++ {
			//fmt.Println("Procesando elemento [", i, "]")
			go proCityWeather(&dataCity[i]) // Se llama a la funcion para procesar el clima
			time.Sleep(80 * time.Millisecond)
		}

		// De las que no se ha podido recuperar el clima.
		// Se cuenta las ciudades que no hay recuperado weather

		vacios := 0

		for i := 0; i <= len(dataCity)-1; i++ {
			if dataCity[i].weather == "vacio" {
				vacios = vacios + 1
				cityEmpty = append(cityEmpty, dataCity[i].city)
			}
		}

		if vacios > 0 {
			// Si hay registros sin contenido del clima se vuelven a procesar
			for i := 0; i <= len(dataCity)-1; i++ {
				if dataCity[i].weather == "vacio" {
					// Solamente se vuelven a consultar las vacias
					go proCityWeather(&dataCity[i]) // Se llama a la funcion para procesar el clima
					time.Sleep(60 * time.Millisecond)
				}
			}
		}

		// Se imprimen los resultados obtenidos
		impCityes(dataCity)

	}

	fmt.Println("\n\n FIN DEL PROGRAMA..")
	fmt.Println("\n\nTiempo transcurrido:", time.Since(now))

}

//--IMPCITYES---------------------------------------------------------------------------
//
//  Imprime en formato los datos recuperados de la web.
//  para cada una de las ciudades.
// Se cambia el nombre de la ciudad para ser mas descriptivo.
func impCityes(cityes []dataWeb) {

	// Se ordena el Slice por el nombre de la ciudad.
	sort.Slice(cityes, func(i, j int) bool {
		return cityes[i].city < cityes[j].city
	})

	for i := 0; i < len(cityes); i++ {
		fmt.Println("No:", i+1)
		fmt.Println("Ciudad      : ", cityes[i].city)
		fmt.Println("Estado      : ", cityes[i].state)
		fmt.Println("Pais        : ", cityes[i].country)

		fmt.Println("Latitud     : ", fmt.Sprintf("%f", cityes[i].latCity))
		fmt.Println("Longitud    : ", fmt.Sprintf("%f", cityes[i].lonCity))
		fmt.Println("Tiempo proc.: ", cityes[i].tdone.Sub(cityes[i].tcreate))
		fmt.Println("Info Clima  : ", cityes[i].weather)

		fmt.Println("------------------------------------------------")
	}
} // END: func impCityes(cityes []dataWeb)

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
//
// Procesa las ciudades de una en una.
func proCitysLoc(city *dataWeb) {

	// Almacena la ubicacion recuperada de la ciudad
	// en la llamada a la funcion queryCityLocation
	var cityLoc Ubicacion

	// Se guarda el momento en que se empieza a procesar la ciudad.
	city.tcreate = time.Now()

	if city.state == "vacio" {
		var bytes []byte // bytes contiene el texto de la consulta web
		// El parametro es el nombre de la ciudad en formato ciudad,PA
		bytes, err := queryCityLocation(city.city, city.country)
		// manejo de error
		if err != nil {
			fmt.Println("Error fatal #1 en proCitys")
			log.Fatal(err)
		}
		// Si la cadena devuelta es [], significa que no se pudo hacer la consulta
		if len(bytes) < 3 {
			fmt.Println(fmt.Sprintf("La ciudad: [%s] no existe o el nombre esta mal escrito.", city.city))
			return // se sale de la funcion ya que no hay contenido
		}
		//--APLICANDO UNMARSHAL-------------------------
		// bytes recupera [] al inicio y final, no se toman en cuenta.
		objetostring2 := string(bytes)[1 : len(bytes)-1]
		json.Unmarshal([]byte(objetostring2), &cityLoc)
		// Luego de aplicar Unmarshal, la variable cityLoc
		// Contiene una cadena como esta:
		// {Loja -3.996845 -79.20167 EC Loja}

		// ** Se elimina todo el manejo de obtencion de valores
		// ** con expresiones regulares.

		intLat := cityLoc.Lat
		intLon := cityLoc.Lon
		// Se almacenan la latitud y la longitud en sus propios campos
		//city.latCity = string(strLat) //aInt, err := strconv.Atoi(a)
		city.latCity = intLat
		city.lonCity = intLon
		city.state = cityLoc.State
		city.tdone = time.Now()
	}

} // FIN: func proCitys(city *dataWeb, i int)

// --QUERYCITYLOCATION-------------------------------------------------------------------------------

func queryCityLocation(city string, country string) ([]byte, error) {

	client := &http.Client{Timeout: 30 * time.Second}
	// codigo nuevo
	urlFind := fmt.Sprintf("https://api.openweathermap.org/geo/1.0/direct?q=%s,,%s&limit=1&appid=3582889a6bd6c9bd3d2867e51f420116", city, country)
	bytes, err := getWebBytes(client, urlFind)
	// manejo de error
	if err != nil {
		fmt.Println("Error fatal en : queryCityLocation()")
		log.Fatal(err)
	}
	return bytes, nil
} // --FIN QUERYCITYLOCATION-------------------------------------------------------------------------

// --QUERYCITYWEATHER-------------------------------------------------------------------------------

func queryCityWeather(latitud string, longitud string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	urlFind := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?units=metric&lat=%s&lon=%s&appid=3582889a6bd6c9bd3d2867e51f420116", latitud, longitud)
	bytes, err := getWebBytes(client, urlFind)
	// manejo de error
	if err != nil {
		fmt.Println("Error fatal en : queryCityWeather()")
		log.Fatal(err)
	}
	return bytes, nil
}

// --PROCCITYWEATHER-------------------------------------------------------------------------

func proCityWeather(city *dataWeb) {
	// Se obtiene la latitud y longitud de la ciudad
	latitud := fmt.Sprintf("%f", city.latCity)
	longitud := fmt.Sprintf("%f", city.lonCity)
	// Se llama a la funcion para consultar el clima en ese lugar
	bytes, err := queryCityWeather(latitud, longitud)
	// manejo de error
	if err != nil {
		fmt.Println("Error fatal en : proCityWeather()")
		log.Fatal(err)
	}
	// Si la cadena devuelta es [], significa que no se pudo hacer la consulta
	if len(bytes) < 3 {
		fmt.Println(fmt.Sprintf("La ciudad: [%s] no existe o el nombre esta mal escrito.", city.city))
		fmt.Println(fmt.Sprintf("o la latitud : [%s] no es la correcta.", city.latCity))
		fmt.Println(fmt.Sprintf("o la longitud: [%s] no es la correcta", city.lonCity))
		return // se sale de la funcion ya que no hay contenido
	}

	city.weather = procWeatherRecovered(bytes)

} //--FIN--PROCCITYWEATHER------------------------------------------------------------------

//--PROCWEATHERRECOVERED--------------------------------------------------------------------
func procWeatherRecovered(cadena []byte) (wproceced string) {

	// Estructura para almacenar el clima de una ciudad

	// la INFORMACION DEL CLIMA EN UNA CIUDAD
	type coord struct {
		Lon float32 `json:"lon"`
		Lat float32 `json:"lat"`
	}
	type weather struct {
		Id          int
		Main        string
		Description string
		Icon        string
	}
	type Mein struct {
		Temp       float32 `json:"temp"`
		Feels_like float32 `json:"feels_like"`
		Temp_min   float32 `json:"temp_min"`
		Temp_max   float32 `json:"temp_max"`
		Pressure   int     `json:"pressure"`
		Humidity   int     `json:"humidity"`
	}

	type wind struct {
		Speed float32 `json:"speed"`
		Deg   int     `json:"deg"`
	}

	type clouds struct {
		All int `json:"all"`
	}

	type sys struct {
		Type    int    `json:"type"`
		Id      int16  `json:"id"`
		Country string `json:"country"`
		Sunrise int32  `json:"sunrise"`
		Sunset  int32  `json:"sunset"`
	}

	type Clima struct {
		Coord      coord   `json:"coord"`
		Weather    weather `json:"weather"`
		Base       string  `json:"base"`
		Main       Mein    `json:"main"`
		Visibility int     `json:"visibility"`
		Wind       wind    `json:"wind"`
		Clouds     clouds  `json:"clouds"`
		Dt         int32   `json:"dt"`
		Sys        sys     `json:"sys"`
		Timezone   int32   `json:"timezone"`
		Id         int32   `json:"id"`
		Name       string  `json:"name"`
		Cod        int     `json:"cod"`
	}

	var vClima Clima

	/////////////////////////////////////////////////////////////////////////////

	//--APLICANDO UNMARSHAL-------------------------")
	// bytes recupera [] al inicio y final, no se toman en cuenta.
	objetostring := string(cadena) //[1 : len(bytes)-1]
	json.Unmarshal([]byte(objetostring), &vClima)

	wproceced = fmt.Sprintf("\n\tTemperatura   : %f Grados", vClima.Main.Temp)
	//wproceced = wproceced + fmt.Sprintf("\n\tSensacion Term: %f", vClima.Main.Feels_like)
	//wproceced = wproceced + fmt.Sprintf("\n\tTemp Mínima   : %f", vClima.Main.Temp_min)
	//wproceced = wproceced + fmt.Sprintf("\n\tTemp Máxima   : %f", vClima.Main.Temp_max)
	wproceced = wproceced + fmt.Sprintf("\n\tPresión Atm.  : %dhPa", vClima.Main.Pressure)
	wproceced = wproceced + fmt.Sprintf("\n\tHumedad       : %d %%", vClima.Main.Humidity)
	wproceced = wproceced + fmt.Sprintf("\n\tVisibilidad   : %dmts", vClima.Visibility)

	// city.weather = fmt.Sprintf("\n\tTemperatura   : %f", vClima.Main.Temp)

	// /////////////////////////////////////////////////////////////////////////////

	return wproceced
}

//--FIN --PROCWEATHERRECOVERED---------------------------------------------------------------

//--GETWEBBYTES-----------------------------------------------------------------------------
//
// Esta funcion obtiene el contenido del sitio web, de acuerdo
// al URL que se le de como parametro.

func getWebBytes(client *http.Client, url string) ([]byte, error) {
	//fmt.Println("Entrando a la funcion getWebBytes()")
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
		fmt.Println("Error fatal #1 en getWebBytes()")
		log.Fatal(err)
	}
	// Ejecutar la consulta. El método Do ejecuta la soliciutd.
	res, err := client.Do(req)
	// Si se tiene un error en la conexion o la consulta se termina el programa
	// y se muestra el mensaje de error generado.
	if err != nil {
		fmt.Println("Error fatal #2 en getWebBytes()")
		log.Fatal(err)
	}
	// La siguiente funcion lee todo el "body" de la pagina web.
	bytes, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}
	return bytes, nil

} // end getWebBytes()
