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


const (
	//RESPUESTAS DEL SERVIDOR
	SERVIDOR_CANAL_APROBADO = "canalaprobado"
	SERVIDOR_CANAL_NOAPROBADO = "canalanoprobado"
	SERVIDOR_SALIR_APROBADO = "saliraprobado"
	SERVIDOR_CONEXION_APROBADO = "conexionaprobada"
	SERVIDOR_ENVIO_APROBADO = "envioaprobado"
	SERVIDOR_ENVIO_NOAPROBADO = "envionoaprobado"
	SERVIDOR_ERROR_CMD = "El comando es invalido"
	SERVIDOR_MSG = "msg" //para crear mensajes estandar sin relevancia para la coordinación, su destion es la pantalla del cliente


	//PETICIONES DEL CLIENTE
	CLIENTE_UNIR_CANAL = "unir"
	CLIENTE_SALIR_CANAL = "salir"
	CLIENTE_CONEXION = "establecerconexion"
	CLIENTE_ENVIAR_ARCHIVO = "enviararchivo"
	CLIENTE_ERROR_CMD = "El comando es invalido"
)




//para agrupar los dos canales del servidor, peticiones y respuestas
type Servidor struct{

	conn net.Conn

	peticion chan string

	respuesta chan string
}

//inicia la lectura sobre la conexion de este servidor
func (s *Servidor) Leer(){
	lector := bufio.NewScanner(s.conn)
	for lector.Scan(){
		entrada := lector.Text()
		//debo filtrar la entrada, o la entrada debe ir directo al interprete
		//y el interprete desempaquetara la respuesta de la forma correspondiente

		s.respuesta<- entrada //las respuestas las atienden las rutinas o la rutina que escribe en la terminal
	}
}

//Permite escribir sobre la conexion con el servidor todos los mensaje entrantes por el canal
//de peticiones
func (s *Servidor) Escribir(){
	for msg := range s.peticion{
		fmt.Fprintln(s.conn, msg) //msg es formateado como cade de texto y enviado por la conexion, termina con  \n
	}
}



func main(){
	conn, err := net.Dial("tcp",":8080")
	if err != nil{
		log.Fatal(err)
	}
	defer conn.Close()
	log.Println("Cliente activo")

	servidor := Servidor{
		conn:conn,
		peticion: make(chan string),
		respuesta: make(chan string),
	}


	go servidor.Leer()
	go servidor.Escribir()
	//la función de aqui abajo parece crear un flujo desde el Stdin a la conexion, no es como que lea lo que hay en stdin y lo envie si no que el
	//flujo siempre esta abierto.
	terminal := bufio.NewScanner(os.Stdin)
	for terminal.Scan(){
		entrada := terminal.Text()
		interpretar(entrada, &servidor) //un comando a la vez
	}
}


func interpretar(entrada string, s *Servidor){
	
		//El interprete deberia desempaquetar las repuestas del ciente?
		//Para despues enviarlas al lugar que corresponda
		// if rsp[0] == "server:"{
		// 	entrada = "---->"+entrada
		// 	io.Writer(os.Stdout).Write([]byte(entrada))
		// }

		comando := strings.Split(entrada, " ")
		switch(comando[0]){
			case "unir":
				//debe existir un mapa entre leguaje comun y el comando, map[unir] = comando/protocolo
				//el formato debe ser unir canal
				if len(comando) != 2{
					log.Println(CLIENTE_ERROR_CMD)
					return
				}
				s.peticion<- CLIENTE_UNIR_CANAL //se queda esperando que alguna rutina lo reciva
			
			case "subir":
				//el formato es subir canal archivo
				if len(comando) != 3{
					log.Println(CLIENTE_ERROR_CMD)
					return
				}
				//deveria ser algo como formaterMsg(protoMsg, carga) para facilitar las cosas y evitar errores
				s.peticion<- CLIENTE_ENVIAR_ARCHIVO + " " + comando[1] + " " + comando[2]
			case "salir":
				log.Println("No puedes salir de este programa !=0")
			default:
				log.Println(comando[0], CLIENTE_ERROR_CMD)
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
func mostrarEnTerminal(respuesta chan<- string, conn net.Conn){
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


