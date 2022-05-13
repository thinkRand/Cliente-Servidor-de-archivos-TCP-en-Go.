package main

import(
	Ps "protocolo_simple"
	"net"
	"log"
	"bufio"
	"os"
	// "io"
	"strings"
	"fmt"
)

var estado string //los estados posibles del cliente CONECTADO, UNIDO_A_CANAL, ACORDANDO_TRANSMISION

//Una structura a manera de abstracción para representar un conjunto de funciones que generan salidas
//a la red y reciven entradas. Son procesos intermediario entre el servidor y el cliente que 
//recive una orden y determina la forma de usar el protocolo simple para cumplirla. 
//Los resultados varian dependiendo de la orden recivida.
type IoCliente struct{
	conn net.Conn
}

func (cli *IoCliente) UnirCanal(param map[string]string){
	//estado = "CONECTADO"
	
	peticion, err := Ps.NuevaPeticion(cli.conn, "unirCanal", param)
	if err != ""{
		log.Println("DEBUG:", err)
		return
	}

	err = peticion.Enviar() 
	if err != "" {
		log.Println("DEBUG:", err)
		return 
	}

	rsp, err := peticion.RecivirRespuesta()
	if err != ""{
		log.Println("DEBUG:", err)
		return 
	}
	
	estado = "UNIDO_A_CANAL"
	if rsp[0] == Ps.S_UNIR_ACEPTADO{ //creo que no debería ser asi, los detalles de implementación deben permanecer en el paquete
		log.Println("Te uniste al canal")
	}else if rsp[0] == Ps.S_UNIR_RECHAZADO{
		log.Println("Fallo. No estas en el canal")
	}

}


var ioCliente IoCliente


func main(){

	conn, err := net.Dial("tcp","127.0.0.1:9999")
	if err != nil{
		log.Fatal(err)
	}
	defer conn.Close()
	estado = "CONECTADO" 
	log.Println("Conexion establecida con el servidor")
	ioCliente.conn = conn

	terminal := bufio.NewScanner(os.Stdin)
	for terminal.Scan(){
		entrada := terminal.Text() //lee una cadena de texto hasta \n
		validarEntrada(entrada) //un comando a la vez
	}

}


//Recive las entradas desde la terminal para validarlas.
//Ejecuta dos niveles de validación: valida el formato, y valida los parametros
func validarEntrada(entrada string){
	divisionEntrada := strings.Split(entrada, " ")
	comando := strings.ToUpper(divisionEntrada[0]) //no case sensitive
	parametros := make(map[string]string) //lis de parametros que acompañan al comando, varian segun el comando

	switch(comando){
	case "UNIR":
		
		// se espera : UNIR <espeacio> canal.
		if len(divisionEntrada) != 2{
			log.Println("Entrada invalida. Se requiere: unir <espacio> nombre-canal")
			return
		}
		parametros["canal"] = divisionEntrada[1]
		ioCliente.UnirCanal(parametros) 


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
		pesoArchivo := fmt.Sprintf("%d", arInfo.Size()) //el peso se envÍa en string
		
		parametros["rutaArchivo"] = rutaArchivo
		parametros["nombreArchivo"] = nombreArchivo 
		parametros["pesoArchivo"] = pesoArchivo 
		parametros["canal"] = nombreCanal
		// intermediarioClienteServidor("enviar", parametros)

	case "SALIR":
		
		//Para salir hay dos opciones: salir del canal o salir completamente del programa
		if len(divisionEntrada) > 2 {
			log.Println("Demaciados parametros para el comado salir: se espera salir <espacio> nombre-canal o")
			log.Println("salir (sin parametros) para salir completamente del programa")
			return
		} 

		
		if len(divisionEntrada) == 1 {
			if estado == "UNIDO_A_CANAL"{
				//primero lo saco del canal
				//luego lo saco del programa
				return
			}else if estado == "CONECTADO"{
				log.Println("Cerrando conexion con el servidor...")
				// intermediarioClienteServidor("desconectar", parametros) //parametros vacios para esta peticion
				return
			}
		}

		
		if len(divisionEntrada) == 2 {
			if estado != "UNIDO_A_CANAL"{
				log.Println("No estas unido a un canal")
				return
			}
			parametros["canal"] = divisionEntrada[1]
			// intermediarioClienteServidor("salir-canal", parametros)
		}

	default:
		log.Println("Entrada invalida")
	}
}












// func enviarArchivo(s *Servidor, parametros []string){
// 	//se espera que entrada tenga [cmd][nombreCanal][nombreArchivo] y agrega el peso antes de
// 	//enviar la petición
	
// 	//a partir de aquí se esta en la fase de transmisión de archivos, el formato de mensajes es 
// 	//en bytes en lugar de strings
// 	canal := parametros[1]
// 	ruta := parametros[2]
	
	

// 	ar, err := os.Open(ruta)
// 	if err != nil{
// 		log.Println(err)
// 		return
// 	}
// 	defer ar.Close()
	
// 	arInfo, err := ar.Stat()
// 	if err != nil{
// 		log.Println(err)
// 		return
// 	}	
// 	peso := arInfo.Size()
// 	nombreAr := arInfo.Name()

// 	//cliente encia msg canal nombreAr peso
// 	s.peticion<- ps.CLIENTE_ENVIAR_ARCHIVO + " " + canal + " " + nombreAr + " " + fmt.Sprintf("%d",peso)
// 	if rsp := <-s.respuesta; rsp == ps.SERVIDOR_ENVIO_APROBADO{
// 		log.Println("El servidor acepto la transferencia del archivo")
		
// 	}else if rsp == ps.SERVIDOR_ENVIO_NOAPROBADO{
// 		log.Println("El servidor no acepto la transferecia")
// 		return
// 	}else{
// 		log.Println("El servidor respondio con:", rsp)
// 		return
// 	}


// 	n , err := io.Copy(s.conn, ar)
// 	if err != nil{
// 		// log.Println("El archivo no se pudo enviar.")
// 		s.peticion<- ps.CLIENTE_ERROR_TERMINAR_ENVIO
// 		log.Println(err)
// 		return
// 	}
// 	// io.Copy(conn, io.Reader(nil))
// 	log.Println("Archivo enviado")
// 	log.Println("Bytes enviados", n)
// 	// log.Println("Toda la petición fue procesada con éxito")
// }

// //anota las respuestas del servidor para que las rutinas puedan acceder a ellas
// func mostrarEnTerminal(respuesta chan<- string, conn net.Conn){
// 	lector := bufio.NewScanner(conn)
// 	for lector.Scan(){
// 		entrada := lector.Text()
// 		//divido entre las comunicaciones internas y los mensajes para la terminal
// 		rsp := strings.Split(entrada, " ")
// 		if rsp[0] == "server:"{
// 			entrada = "---->"+entrada
// 			io.Writer(os.Stdout).Write([]byte(entrada))
// 		}
// 			respuesta<- entrada 
// 	}
	
// }


