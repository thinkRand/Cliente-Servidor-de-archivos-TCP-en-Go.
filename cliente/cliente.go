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
	CLIENTE_ERROR_TERMINAR_ENVIO = "terminar"


	//RESPUESTAS DE LA TERMINAL
	CLIENTE_ERROR_NUM_PARAMETROS = "El numero de parametros es incorrecto"
	CLIENTE_ERROR_CMD = "El comando es invalido"
)




//para agrupar los dos canales del servidor, peticiones y respuestas
type Servidor struct{

	conn net.Conn

	peticion chan string

	respuesta chan string
}

//inicia la lectura sobre la conexion 
func (s *Servidor) Leer(){
	lector := bufio.NewScanner(s.conn)
	for lector.Scan(){
		entrada := lector.Text()
		log.Println("Recivido: ", entrada)
		//debo filtrar la entrada, o la entrada debe ir directo al interprete
		//y el interprete desempaquetara la respuesta de la forma correspondiente

		s.respuesta<- entrada //las respuestas las atienden las rutinas o la rutina que escribe en la terminal
	}
}

//Permite escribir sobre la conexion con el servidor todos los mensaje entrantes por el canal
//de peticiones
func (s *Servidor) Escribir(){
	for msg := range s.peticion{
		fmt.Fprintln(s.conn, msg) //msg es formateado como string y enviado por la conexion, termina con  \n
		log.Println("Despachado:",msg)
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
		verificarComando(entrada, &servidor) //un comando a la vez
	}
}


func verificarComando(entrada string, s *Servidor){
	
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
					log.Println(CLIENTE_ERROR_NUM_PARAMETROS)
					return
				}
				s.peticion<- CLIENTE_UNIR_CANAL //se queda esperando que alguna rutina lo reciva
				//si la peticion es recivida en el canal del servidor entonces continuo con otra cosa
			
			case "subir":
				//El formato es subir canal archivo
				//se puede hacer un bool, err := veri(comando, "subir")
				//if err  != nil chacata
				if len(comando) != 3{
					log.Println(CLIENTE_ERROR_NUM_PARAMETROS)
					return
				}
				enviarArchivo(s, comando)

			case "salir":
				log.Println("No puedes salir de este programa !=0")
			default:
				log.Println(CLIENTE_ERROR_CMD, ":",comando[0])
			}
}

func enviarArchivo(s *Servidor, parametros []string){
	//se espera que entrada tenga [cmd][nombreCanal][nombreArchivo] y agrega el peso antes de
	//enviar la petición
	
	//a partir de aquí se esta en la fase de transmisión de archivos, el formato de mensajes es 
	//en bytes en lugar de strings
	canal := parametros[1]
	ruta := parametros[2]
	
	

	ar, err := os.Open(ruta)
	if err != nil{
		log.Println(err)
		return
	}
	defer ar.Close()
	
	arInfo, err := ar.Stat()
	if err != nil{
		log.Println(err)
		return
	}	
	peso := arInfo.Size()
	nombreAr := arInfo.Name()

	//cliente encia msg canal nombreAr peso
	s.peticion<- CLIENTE_ENVIAR_ARCHIVO + " " + canal + " " + nombreAr + " " + fmt.Sprintf("%d",peso)
	if rsp := <-s.respuesta; rsp == SERVIDOR_ENVIO_APROBADO{
		log.Println("El servidor acepto la transferencia del archivo")
		
	}else if rsp == SERVIDOR_ENVIO_NOAPROBADO{
		log.Println("El servidor no acepto la transferecia")
		return
	}else{
		log.Println("El servidor respondio con:", rsp)
		return
	}


	n , err := io.Copy(s.conn, ar)
	if err != nil{
		// log.Println("El archivo no se pudo enviar.")
		s.peticion<- CLIENTE_ERROR_TERMINAR_ENVIO
		log.Println(err)
		return
	}
	// io.Copy(conn, io.Reader(nil))
	log.Println("Archivo enviado")
	log.Println("Bytes enviados", n)
	// log.Println("Toda la petición fue procesada con éxito")
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


