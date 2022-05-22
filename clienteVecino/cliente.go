package main

import(
	Ps "protocoloSimple"
	"net"
	"bufio"
	"os"
	"strings"
	"fmt"
)


var (
	
	kill = false //True si algun proceso indica que este cliente debe dejar de funcionar

	estado string //los estados posibles del cliente CONECTADO, UNIDO_A_CANAL, ACORDANDO_TRANSMISION, ATENDIENDO_ORDEN

	ioCliente IoCliente //input ouput del cliente. Su deber es presentar las funciones para enviar a la red y recivir de la red, usando el <protoco simple>.

	canalActual string //El nombre del canal actual al que está unido el cliente

	respuestasServidor = make(chan string) //Aquí se reciven las repuestas del servidor

	clienteEsperaRespuesta = false //indica si el cliente está esperando una respuesta del servidor

	ordenRecivida = "ordenrcv" //mensaje utilisado por las subrutinas para sabe si se recivio una orden y debe dejar de esperar una respuesta del servidor
)

func main(){

	conn, err := net.Dial("tcp","127.0.0.1:9999") //9999 es el puerto estandar para el <protocolo simple>
	if err != nil{
		fmt.Println(err)
		return
	}
	defer conn.Close()
	estado = "CONECTADO" 
	fmt.Println("Conexion establecida con el servidor")
	ioCliente.Conn = conn
	go escucharRespuestaOrdenes()

	fmt.Println("Comandos:")
	fmt.Println("unir <espacio> canal")
	fmt.Println("enviar <espacio> canal <espacio> ruta-archivo")

	//Comienso a recivir las entradas desde la terminal
	terminal := bufio.NewScanner(os.Stdin)
	for terminal.Scan(){
		entrada := terminal.Text() //lee una cadena de texto hasta \n
		validarEntrada(entrada) //un comando a la vez. Es sincrona.
		if kill {
			break
		}
	}
	fmt.Println("Conexion perdida, bye...")
}


//Recive las entradas desde la terminal para validarlas.
//Cada comando debería tener su funcion de validación.
//dependiendo de la entrada se pueden tener dos niveles de validacion. Validación de formato y validación de
//parametros
func validarEntrada(entrada string){
	if estado == "ATENDIENDO_ORDEN"{
		fmt.Println("Se esta atendiendo una orden en este momento")
		return
	}
	divisionEntrada := strings.Split(entrada, " ")
	comando := strings.ToUpper(divisionEntrada[0]) //no case sensitive
	campos := make(map[string]string) //lista de campos que acompañan al comando, varian segun el comando, esto debería ser una global.

	switch(comando){
	case "UNIR":
		// se espera : UNIR <espeacio> canal.
		if len(divisionEntrada) != 2{
			fmt.Println("Entrada invalida. Se requiere: unir <espacio> nombre-canal")
			return 
		}
		campos["nombreCanal"] = divisionEntrada[1]
		gestionar("unir", campos)


	case "ENVIAR":
		//solo se puede enviar si se esta en el estado de unido a canal
		if estado != "UNIDO_A_CANAL"{
			fmt.Println("No estas unido a un canal")
			return 
		}

		// se espera : ENVIAR <espeacio> canal <espacio> ruta-archivo. Esto es independiente del protocolo simple
		if len(divisionEntrada) != 3{
			fmt.Println("Entrada invalida. Se requiere: enviar <espacio> nombre-canal <espacio> ruta-archivo")
			return 
		}

		//nivel 2 de validacion: se comprueba que el archivo existe y es legible.
		rutaArchivo := divisionEntrada[2]
		ar, err := os.Open(rutaArchivo)
		if err != nil{
			fmt.Println("Error al intentar abrir la ruta del archivo")
			fmt.Println(rutaArchivo)
			fmt.Println(err)

			return 
		}
		arInfo, err := ar.Stat()
		if err != nil{
			fmt.Println("Error al intentar leer la estructura del archivo")
			return 
		}
		ar.Close()
		//validación completada

		nombreCanal := divisionEntrada[1]
		nombreArchivo := arInfo.Name()
		pesoArchivo := fmt.Sprintf("%d", arInfo.Size()) //el peso se envía en string
		
		campos["rutaArchivo"] = rutaArchivo
		campos["nombreArchivo"] = nombreArchivo 
		campos["pesoArchivo"] = pesoArchivo 
		campos["canal"] = nombreCanal
		gestionar("enviar", campos)

	case "SALIR": 

		fmt.Println("Opcion no habilidata por ahora")
		if true{
			return
		}
		//Para salir hay dos opciones: salir del canal o salir completamente del programa
		if len(divisionEntrada) > 2 {
			fmt.Println("Demaciados campos para el comado salir: se espera salir <espacio> nombre-canal o")
			fmt.Println("salir (sin campos) para salir completamente del programa")
			return 
		} 

		
		if len(divisionEntrada) == 1 {
			if estado == "UNIDO_A_CANAL"{
				//primero lo saco del canal
				//luego lo saco del programa
				return 
			}else if estado == "CONECTADO"{
				fmt.Println("Cerrando conexion con el servidor...")
				return 
			}
		}

		
		if len(divisionEntrada) == 2 {
			if estado != "UNIDO_A_CANAL"{
				fmt.Println("No estas unido a un canal")
				return 
			}
			campos["canal"] = divisionEntrada[1]
	
		}


	case "DEBUG":
		MuestraRegistro()


	default:
		// campos["msg"] = entrada
		// gestionar("toChan", campos)
		// return 
		fmt.Println("Entrada invalida. Prueba unir, enviar, debug, conectar o salir")
	}
	return 
}





//Utilisa el ioCliente para hacer las peticiones y determina que hacer con las respuestas.
func gestionar(orden string, campos map[string]string){
	switch orden{

	case "unir":
		exito, serr := ioCliente.PeticionUnirCanal(campos["nombreCanal"])
		if serr != "" {
			if serr == "conndead"{
				kill = true
				return
			}
			fmt.Println(serr)
			return
		} 

		if exito {
			estado = "UNIDO_A_CANAL"
			canalActual = campos["nombreCanal"]
			fmt.Println("Te uniste al canal")
		}else{
			//continua en estado CONECTADO
			fmt.Println("Fallo. No estas en el canal")
		}


	case "enviar":
		serr := ioCliente.PeticionEnviarArchivo(campos["rutaArchivo"], campos["nombreArchivo"], campos["pesoArchivo"], campos["canal"])
		if serr != "" {
			if serr == "conndead"{
				kill = true 
				return
			}else{
				fmt.Println(serr)
				return
			}
		}
		fmt.Println("El archivo fue recivido y aceptado")


	default:

	}

}








//Recive todos los mensaje desde el servidor y los envia al proceso de filtrar para que se filtren por
//ordenes o respuestas. 
//Si son respuestas a peticiones las enruta por el canal de respuestasServidor. 
//Si es una orden le asigna el tratamiento adecuado.
func escucharRespuestaOrdenes(){
	lector := bufio.NewScanner(ioCliente.Conn)
	for lector.Scan(){
		entrada := lector.Text()
		fmt.Println("escucharRespuestaOrdenes:", entrada)
		filtrarMensaje(entrada)
	}
}

//Filtra mensaje en ordenes o respuestas, las respuestas van a la rutina que las solicito, las ordens
//se atienden aparte y terminan cualquier rutina que espere una respuesta para indica que se debe atender
//la orden in nada más
func filtrarMensaje(msj string){
	campos := strings.Split(msj, " ")

	if campos[0] == Ps.S_ORDEN {
		//en caso de orden se espera: orden <espacio> codigo <espacio> campos...
		fmt.Println("Orden recivida")
		if clienteEsperaRespuesta{
			respuestasServidor<- ordenRecivida //este mensaje indica a cualquir rutina que este esperando una respuesta que termine su procedimiento
			return
		}
		serr := gestionarOrdenRecivida(msj)
		if serr != ""{
			if serr == "conndead"{
				kill = true
				return
			}else{
				fmt.Println(serr)
				return
			}
		}
		return
	}
	//si no es una orden entonces es una respues del servidor
	if clienteEsperaRespuesta{
		respuestasServidor<- msj
		return 
	}else{
		//panico, el servidor envio un mensaje inesperado
		fmt.Println("El servidor envio un mensaje inesperado:", []byte(msj))
		kill = true
		return
	}
}


//Identifica el tipo de orden y le asigana el tratamiento adecuado. Una orden es un mensaje con 
//formato de protocolo simple.
//Una orden indica a este cliente que inicie el proceso para procesar una solicitud específica y no atienda
//ninguna otra cosas durante ese proceso
//Todo mensaje en esta parte se debe enviar con el prefijo Ps.S_ORDEN+" "+ y con un \n al final debido a 
//que le otro extremo está usando un scan.Text()
func gestionarOrdenRecivida(orden string) (serr string){
	fmt.Println("Gestionando ordern recivida")
	estado = "ATENDIENDO_ORDEN"
	defer func(){
		estado = "UNIDO_A_CANAL"
	}() 

	campos := strings.Split(orden, " ")
	if campos[1] == Ps.C_ENVIAR_ARCHIVO {  
		//se espera el formato: S_ORDEN <espacio> enviar <espacio> canal <espacio> nombre-archivo <espacion> peso-archivo\n
		if len(campos) != 5{
			serr = Ps.ReplicaTermina(Ps.S_ORDEN+" "+Ps.S_ENVIO_RECHAZADO+"\n", ioCliente.Conn)
			if serr != ""{
				if serr == "conndead"{
					return "conndead"
				}
				return "conndead"
			}
			return 
		} 
		
		canal := campos[2]
		if canal != canalActual{
			fmt.Println("Se recivio una orden de un canal desconocido")
			serr = Ps.ReplicaTermina(Ps.S_ORDEN+" "+Ps.S_ENVIO_RECHAZADO+"\n", ioCliente.Conn)
			if serr != ""{
				if serr == "conndead"{
					return "conndead"
				}
				return "conndead"
			}
			return
		}

		nombreArchivo := campos[3]
		pesoArchivo := campos[4]
		serr = ioCliente.OrdenReciveArchivo(nombreArchivo, pesoArchivo)
		if serr != "" {
			if serr == "conndead"{
				kill = true //termina con este cliente
				return
			}else{
				fmt.Println(serr)
				return
			}
		}
		return "Orden <recivir archivo> procesada satisfactoriamente"
	}else{
		return "Orden desconocido: "+orden

	}

}
