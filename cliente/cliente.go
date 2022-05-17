package main

import(
	ioc "ioCliente"
	"net"
	"log"
	"bufio"
	"os"
	"strings"
	"fmt"
)

var kill = false //True si algun proceso indica que este cliente debe dejar de funcionar

var estado string //los estados posibles del cliente CONECTADO, UNIDO_A_CANAL, ACORDANDO_TRANSMISION

var ioCliente ioc.IoCliente //input ouput del cliente. Su deber es presentar las funciones para enviar a la red y recivir de la red, usando el <protoco simple>.


func main(){

	conn, err := net.Dial("tcp","127.0.0.1:9999") //9999 es el puerto estandar para el <protocolo simple>
	if err != nil{
		log.Fatal(err)
	}
	defer conn.Close()
	estado = "CONECTADO" 
	log.Println("Conexion establecida con el servidor")
	ioCliente.Conn = conn

	

	//Comienso a recivir las entradas desde la terminal
	terminal := bufio.NewScanner(os.Stdin)
	for terminal.Scan(){
		entrada := terminal.Text() //lee una cadena de texto hasta \n
		validarEntrada(entrada) //un comando a la vez. Es sincrona.
		if kill {
			break
		}
	}
	log.Println("Cliente cerrado, bye...")
}


//Recive las entradas desde la terminal para validarlas.
//Cada comando debería tener su funcion de validación.
//dependiendo de la entrada se pueden tener dos niveles de validacion. Validación de formato y validación de
//parametros
func validarEntrada(entrada string){
	divisionEntrada := strings.Split(entrada, " ")
	comando := strings.ToUpper(divisionEntrada[0]) //no case sensitive
	campos := make(map[string]string) //lista de campos que acompañan al comando, varian segun el comando, esto debería ser una global.

	switch(comando){
	case "UNIR":
		
		// se espera : UNIR <espeacio> canal.
		if len(divisionEntrada) != 2{
			log.Println("Entrada invalida. Se requiere: unir <espacio> nombre-canal")
			return 
		}
		campos["nombreCanal"] = divisionEntrada[1]
		gestionar("unir", campos)


	case "ENVIAR":

		//solo se puede enviar si se esta en el estado de unido a canal
		if estado != "UNIDO_A_CANAL"{
			log.Println("No estas unido a un canal")
			return 
		}

		// se espera : ENVIAR <espeacio> canal <espacio> ruta-archivo. Esto es independiente del protocolo simple
		if len(divisionEntrada) != 3{
			log.Println("Entrada invalida. Se requiere: enviar <espacio> nombre-canal <espacio> ruta-archivo")
			return 
		}

		//nivel 2 de validacion: se comprueba que el archivo existe y es legible.
		rutaArchivo := divisionEntrada[2]
		ar, err := os.Open(rutaArchivo)
		if err != nil{
			log.Println("Error al intentar abrir la ruta del archivo")
			log.Println(rutaArchivo)
			log.Println(err)

			return 
		}
		arInfo, err := ar.Stat()
		if err != nil{
			log.Println("Error al intentar leer la estructura del archivo")
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
		
		//Para salir hay dos opciones: salir del canal o salir completamente del programa
		if len(divisionEntrada) > 2 {
			log.Println("Demaciados campos para el comado salir: se espera salir <espacio> nombre-canal o")
			log.Println("salir (sin campos) para salir completamente del programa")
			return 
		} 

		
		if len(divisionEntrada) == 1 {
			if estado == "UNIDO_A_CANAL"{
				//primero lo saco del canal
				//luego lo saco del programa
				return 
			}else if estado == "CONECTADO"{
				log.Println("Cerrando conexion con el servidor...")
				return 
			}
		}

		
		if len(divisionEntrada) == 2 {
			if estado != "UNIDO_A_CANAL"{
				log.Println("No estas unido a un canal")
				return 
			}
			campos["canal"] = divisionEntrada[1]
	
		}


	case "DEBUG":
		ioc.MuestraRegistro()


	default:
		campos["msg"] = entrada
		gestionar("toChan", campos)
		return 
		// log.Println("Entrada invalida")
	}
	return 
}





//Utilisa el ioCliente para hacer las peticiones y determina que hacer con las respuestas.
//Su intención es ser como un controller en el modelo MVC
func gestionar(orden string, campos map[string]string){
	switch orden{
	case "unir":
		
		exito, serr := ioCliente.UnirCanal(campos["nombreCanal"])
		if serr != "" {
			log.Println(serr)
			return
		} 

		if exito {
			estado = "UNIDO_A_CANAL"
			log.Println("Te uniste al canal")
		}else{
			//continua en estado CONECTADO
			log.Println("Fallo. No estas en el canal")
		}

	case "salir":


	case "toChan":		
		ioCliente.ToChan(campos["msg"])


	case "enviar":

	serr := ioCliente.EnviarArchivo(campos["rutaArchivo"], campos["nombreArchivo"], campos["pesoArchivo"], campos["canal"])
	if serr == "conndead"{
		kill = true 
	}


	
	default:

	}

}


