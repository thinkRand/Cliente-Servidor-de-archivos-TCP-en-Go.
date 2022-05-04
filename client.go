package main

import(
	"net"
	"log"
	"bufio"
	"os"
	"io"
	"strings"
	"fmt"
)




func main(){
	conn, err := net.Dial("tcp",":8080")
	if err != nil{
		log.Fatal(err)
	}
	log.Println("Cliente activo")
	interpretar := make(chan string) //para comunicar las entradas al interprete de comandos
	respuesta := make(chan string) //aquí se escriben las respuestas del servidor, para que las demas rutinas las puedan ver
	peticion := make(chan string) //enviar peticiones al servidor, la rutina peticionServidor escucha este canal para enviar las peticiones al servidor
	echo := make(chan bool) //para sincronisar
	

	go interprete(interpretar, peticion, respuesta, echo, conn)
	go respuestasServidor(respuesta, conn)
	go peticionServidor(peticion, conn)
	//la función de aqui abajo parece crear un flujo desde el Stdin a la conexion, no es como que lea lo que hay en stdin y lo envie si no que el
	//flujo siempre esta abierto.
	terminal := bufio.NewScanner(os.Stdin)
	for terminal.Scan(){
		entrada := terminal.Text()
		interpretar<- entrada 
		<- echo //espero a que el mensajes se procese
	}
	conn.Close()
}


func interprete(interpretar <-chan string, peticion chan string, respuesta chan string, echo chan<- bool, conn net.Conn){
	for entrada := range interpretar{
		comando := strings.Split(entrada, " ")
		switch(comando[0]){
			case "unir":
				peticion<- entrada
			case "subir":
				log.Println("Comando subir")
				enviarArchivo(comando[1:], conn, respuesta, peticion)
			case "obtener":
				log.Println("comando obtener")
				log.Println("Toda la petición fue procesada con exito")
			case "salir":
				log.Println("No puedes salir de este progama ")
				log.Println("Toda la petición fue procesada con exito")
			default:
				log.Println(comando[0], "no es un comando valido")
			}
			echo<- true //permite que la terminal lea la siguiente entrada
		}
}

func enviarArchivo(entrada []string, conn net.Conn, respuesta <-chan string, peticion chan<- string){
	//obtengo el nombre del archivo
	//se espera que entrada tenga [nombreArchivo] [nombreCanal]
	//y esta validaciones deberian estar en el interprete
	if len(entrada) < 2{
		log.Println("Parametro faltantes: debe ser: subir nombreArchivo nombreCanal")
		return
	}
	archivo := entrada[0]
	canal := entrada[1]
	if len (canal) <= 0{
		log.Println("Parametro faltante: canal")
		return
	}
	//el nombre del archivo no debe tener espacios
	//Si el nombre del archivo tiene espacios los elimino con TrimSpace
	ar, err := os.Open(archivo)
	if err != nil{
		// log.Println("El archivo",archivo,"no se pudo leer")
		log.Println(err)
		return
	}
	defer ar.Close()
	// arInfo, err := ar.Stat()
	//evio informacion del archivo al respuesta
	// conn.Write([]byte(arInfo.Name())) 
	// conn.Write([]byte(string(arInfo.Size()))) //el error de impresion tiene que ver con UTF-8
	
	//coordino la entrega con el servidor
	peticion<- "up " + canal
	//me tengo que quedar esperando la respuesta de la conexion
	rsp := <-respuesta
	if rsp != "ok"{
		log.Println("rsp: ", rsp +"\n")
		log.Println("El servidor no está listo para recivir el archivo")
		return
	}
	//Envio el archivo
	n , err := io.Copy(conn, ar)
	if err != nil{
		// log.Println("El archivo no se pudo enviar.")
		log.Println(err)
		return
	}
	io.Copy(conn, io.Reader(nil))
	log.Println("Archivo enviado")
	log.Println("Bytes enviados", n)
	log.Println("Toda la petición fue procesada con exito")
}

//envia las peticiones al servidor
//para esta rutina el canal peticion es solo para escuchar
func peticionServidor(peticion <-chan string, conn net.Conn){
	//escribe la petición a la conexion
	for p := range peticion{
		fmt.Fprintln(conn, p) //se ignoran los errores de red como desconexion en cuyo caso esta runita se cerrara porque peticion se cerrara
	}
}

//anota las respuestas del servidor para que las rutinas puedan acceder a ellas
func respuestasServidor(respuesta chan<- string, conn net.Conn){
	lector := bufio.NewScanner(conn)
	for lector.Scan(){
		entrada := lector.Text()
		//divido entre las comunicaciones internas y los mensajes para la terminal
		rsp := strings.Split(entrada, " ")
		if rsp[0] == "server:"{
			entrada = "---->"+entrada
			io.Writer(os.Stdout).Write([]byte(entrada))
		}
			respuesta<- entrada 
	}
	
}


