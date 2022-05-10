package main

import(
	ps "protocolo_simple"
	"net"
	"log"
	"bufio"
	"os"
	"io"
	"strings"
	"fmt"
)


//Probablemente es mas claro decir que este debe ser un Servidor de protocolo simple
//para que se entienda que recive peticions de protocolo simple y responde de la misma forma
type Servidor struct{

	conn net.Conn

	// peticion chan string

	// respuesta chan string
}

// //inicia la lectura sobre la conexion 
// func (s *Servidor) Leer(){
// 	lector := bufio.NewScanner(s.conn)
// 	for lector.Scan(){
// 		entrada := lector.Text()
// 		log.Println("Recivido: ", entrada)
// 		//debo filtrar la entrada, o la entrada debe ir directo al interprete
// 		//y el interprete desempaquetara la respuesta de la forma correspondiente

// 		s.respuesta<- entrada //las respuestas las atienden las rutinas o la rutina que escribe en la terminal
// 	}
// }

// //Permite escribir sobre la conexion con el servidor todos los mensaje entrantes por el canal
// //de peticiones
// func (s *Servidor) Escribir(){
// 	for msg := range s.peticion{
// 		fmt.Fprintln(s.conn, msg) //msg es formateado como string y enviado por la conexion, termina con  \n
// 		log.Println("Despachado:",msg)
// 	}
// }


func main(){

	conn, err := net.Dial("tcp","127.0.0.1:9999")
	if err != nil{
		log.Fatal(err)
	}
	defer conn.Close()
	log.Println("Conexion establecida con el servidor")

	
	servidor := Servidor{
		conn:conn,
		// peticion: make(chan string),
		// respuesta: make(chan string),
	}
	// go servidor.Leer()
	// go servidor.Escribir()
	
	//para leer las entradas de la terminal
	terminal := bufio.NewScanner(os.Stdin)
	for terminal.Scan(){
		entrada := terminal.Text() //lee hasta \n
		validarEntrada(entrada, &servidor) //un comando a la vez
	}
}


//Recive las entradas desde la terminal para validarlas.
//Ejecuta dos niveles de validación: valida el formato, y valida los parametros
func validarEntrada(entrada string, s *Servidor){
	divisionEntrada := strings.Split(entrada, " ")

	switch(strings.ToUpper(divisionEntrada[0])){
	case "UNIR":
		// se espera : UNIR <espeacio> canal . Esto es independiente del protocolo simple
		if len(divisionEntrada) != 2{
			log.Println("Entrada invalida. Se requiere unir <espacio> nombre-canal")
			return
		}

		parametros := map[string]string{
			"canal":divisionEntrada[1],
		}
		msjExito, msjErr := intermediarTransaccion("unir", parametros, s) 
		if msjErr != ""{
			log.Println(msjErr)
			return
		}

		if msjExito != ""{
			log.Println(msjExito)
			return
		}

		//si se llega aquí es un error
		log.Println("ERROR DESCONOCIDO")


	case "ENVIAR":
		// se espera : ENVIAR <espeacio> canal <espacio> ruta-archivo. Esto es independiente del protocolo simple
		if len(divisionEntrada) != 3{
			log.Println("Entrada invalida. Se requiere enviar <espacio> nombre-canal <espacio> ruta-archivo")
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
		//validación completada

		nombreCanal := divisionEntrada[1]
		nombreArchivo := arInfo.Name()
		pesoArchivo := fmt.Sprintf("%d", arInfo.Size()) //el peso se envia en string
		
		parametros := map[string]string{
			"nombreArchivo":nombreArchivo, 
			"pesoArchivo":pesoArchivo, 
			"canal":nombreCanal}
		intermediarTransaccion("enviar", parametros, s)

	case "SALIR":
		//Para salir hay dos opciones: salir del canal o salir completamente del programa
		if len(divisionEntrada) > 2 {
			log.Println("Demaciados parametros para el comado salir: se espera salir <espacio> nombre-canal o")
			log.Println("salir (sin parametros) para salir completamente del programa")
			return
		} 
		
		if len(divisionEntrada) == 1 {
			log.Println("Quieres salir del programa... bye")
			log.Fatal()
			return
		}


		if len(divisionEntrada) == 2 {
			//parametros: nombre-canal
			//probablemente desde aquí me voy a ir al intermediario
			//msg := ps.Traducir("salir", parametros)
			//msg := ps.Empaquetar(msg)
			//ps.peticion(msg).enviar()
			//ps.respuesta() //recive la ultima respuesta del servidor
			return
		}
		return

	case "DESCONECTAR":
	

	default:
		log.Println("Entrada invalida")
	}
}


//Recive una orden del cliente y usa el protocolo simple para determinar la forma de comunicarce 
//con el servidor para cumplir con esa orden.
//Los resultados varian dependiendo de la orden recivida
func intermediarTransaccion(orden string, parametros map[string]string, s *Servidor)(respuesta string, error string){

	switch(orden){
	case "unir":
		//transforma la orden unir al correspondiente formato en el protocolo simple
		peticion, err := ps.Traducir("unir", parametros)
		if err != ""{
			return "", err
		}
		peticion = ps.Empaquetar(peticion) //hace nada, es para demostrar un comportamiento esperado mas adelante
		rsp, err := ps.HacerPeticion(peticion, s.conn) 
		if err != "" {
			return "", err
		}
		
		rsp = ps.Desempaquetar(rsp)
		if rsp == ps.SERVIDOR_UNIR_APROBADO{

		}else if rsp == ps.SERVIDOR_UNIR_NOAPROBADO{
			//la respuesta que se vera en la terminal
			return "", "El servidor rechaso la union al canal"
		}
	

	case "enviar":
	

	default:
		return "", "La orden recivida no se reconoce"
	}
	return "","No se recivio una orden"
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


