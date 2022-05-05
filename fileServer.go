package main

import (
	"log"
	"net"
	"os"
	"bufio"
	"strings"
	"fmt"
)

/*
	Me toca crear un canal donde los clientes se registren
	Se me ocurre un contenedor de tipo canal, que almacene todos los clientes conectados
	a el, que tenga sus propias rutinas para recivir y enviar comandos

	inicialmente voy a lograr que se transmitan mensajes para luego pasar a archivos

	tan sencillo como eso, el canal es un map con todos los clientes registrados en el.

	al unirce a un canal(anotarce en el map) todos los mensajes que envien deben ser enviados al canal
	por eso debe haber una forma de saber a que canal esta conectado un cliente.
	puede ser un mapa de nombre canales map[cliente] := canal al que esta conectado
*/
const (
	//RESPUESTAS DEL SERVIDOR
	SERVIDOR_CANAL_APROBADO = "canalaprobado"
	SERVIDOR_SALIR_APROBADO = "saliraprobado"
	SERVIDOR_CONEXION_APROBADO = "conexionaprobada"
	SERVIDOR_ENVIO_APROBADO = "envioaprobado"
	SERVIDOR_ERROR = "El comando es invalido" 


	//PETICIONES DEL CLIENTE
	CLIENTE_UNIR_CANAL = "unir"
	CLIENTE_SALIR_CANAL = "salir"
	CLIENTE_CONEXION = "establecerconexion"
	CLIENTE_ENVIAR_ARCHIVO = "enviararchivo"
)



//El canal es donde se registran clientes
//El canal debe escuchar activamente todos los mensajes que se envian a el, tal como en un chat
type canal struct{
	//el nombre de este canal
	nombre string 

	//Registro de los clientes en este canal
	clientes map[cliente]bool 

	//recive todos los mensajes para escribir en el canal, incluso archivos
	escribir chan string 

	//recive todas la peticiones para unirce a este canal
	unir <-chan cliente

	//recive todas las peticiones para salir de este canal
	salir <-chan cliente
}


//iniciar comiensa la rutina de este canal que escucha todos los mensajes que llegan
func (can *canal) Iniciar(){
	log.Println("El canal1 esta activo y escuchando")
	for{
		select{
		case cli := <-can.unir:
			can.clientes[cli] = true
		
		case cli := <-can.salir:
			delete(can.clientes, cli)
		
		case msg := <-can.escribir:
			for cli := range can.clientes{
				cli.escribir<-msg
			}
		}
	}
}


type cliente struct{
	
	conn net.Conn
	
	escribir chan string //para escribir mensajes al cliente

	canal *canal //posiblemente haga cliente.canal.escribir<- msg

}


//lee los mensajes provenientes del cliente y los manda al interprete de comandos
func (cli *cliente) Leer(){
	lector := bufio.NewScanner(cli.conn)
	for lector.Scan(){
		entrada := lector.Text()
		cli.Interpretar(entrada) //un interprete por cliente
		//cuando el interprete retorna se vuelven a escuchar entradas desde el cliente
	}
}


//Toma todos los mensaje para este cliente y se los envia
func (cli *cliente) Escribir(){
	for msg := range <-cli.escribir{
		fmt.Fprintln(cli.conn, msg)
	}
}


//Recive los mensajes del cliente y reconocer los comandos para ofrecer el procedimiento adecuado
func (cli *cliente) Interpretar(entrada string){
	var comando []string
	comando = strings.Split(entrada, " ")
	//segun el protocolo la primera palabra es el comand
	//el formato es comando valor valor
	switch(comando[0]){

	case CLIENTE_UNIR_CANAL:
		//le puedo pasar un mensaje a una rutina del servidor que identifica a cada
		//canal creado

		if (estaCanalExiste(comnado[1])){
			//aqui se debe comprobar si el canal existe
			if canales[comando[1]]{

			}
			canal1.clientes[cliente] = true
			cliente<- "server: " + " Estas registrado en " + comando[1] +"\n"
		}else{
			cliente<- "server: "+ comando[1] + " no es un canal\n"
		}
		//el msg debe venir en for up espacio nombreCanal
	case CLIENTE_ENVIAR_ARCHIVO:
		log.Println("Solicitud up recivida desde cliente")
		//si el nombre del canal esta registrado en el registro de canales del cliente
		//algo como cliente.canales[comando[1]] == true
		cliente<- "server: "+ "recivida la peticion >up< en el canal "+ comando[1] + "\n"
		recivirArchivo(conn, cliente)
	default:
		log.Println(entrada, SERVIDOR_ERROR)
	}
	return
}


	
func main(){
	ln, err := net.Listen("tcp",":8080")
	if err 	!= nil{
		log.Fatal(err)
	}
	var canal1 canal
	canal1.nombre = "canal1"
	go canal1.Iniciar()

	listaCanales := make(map[string]bool) 
	listaCanales[canal1.nombre] = true

	
	log.Println("Servidor activo...")
	for{
		conn, err := ln.Accept()
		if err !=  nil{
			//ingnoro las conexines fallidas
			continue
		}
		go handlerCliente(conn, listaCanales)
	}
}


//handler del cliente 
func handlerCliente(conn net.Conn, listaCanales map[string]bool){
	var cli cliente
	cli.conn = conn
	cli.escribir = make(chan string)
	defer cli.conn.Close() //me aseguro de cerra la conexion en cualquier caso	
	log.Println("Cliente conectado: ", cli.conn.RemoteAddr().String())

	//me progunto como puedo agregar la lista de canales como una variable del cliente
	//cli.listaCanales = ListaCanales
	//se me ocurre que todos los cliente entren por defecto a un canl default o sala de espera
	//en la sala de espera se conoce la lista de todos los canales, entonces aqui el interprete
	//de la sala de espera puede puede procesar la peticion de union de los clientes
	//con el uso de if listaCanales[comando] == true 

	go cli.Escribir() //para enviar los mensajes al cliente
	cli.Leer() //se mantiene escuchado al cliente
	log.Println("El cliente ",cli.conn.RemoteAddr().String(),"se desconecto")
}



func recivirArchivo(conn net.Conn, cliente chan<- string){
		//aquí estamos en fase de comunicación interna. Ningun mensaje comiensa con >server:<
			archivo, err := os.Create("archivo")
			if err != nil{
				log.Println("No se pudo crear la ruta local para el archivo")
				cliente <- "error"
				return
			}
			defer archivo.Close() 
			BUFFER_TAMANIO := 1024 //1024 bytes
			buffer := make([]byte, BUFFER_TAMANIO)
			//servidor listo para recivir archivo
			cliente <- "ok"
			var cuenta int 
			for {
				n, err := conn.Read(buffer)
				cuenta+=n

				if err != nil{
					//posible desconexion
					//se debe eliminar el archivo
					log.Fatal(err)
				}
			
				_, aerr := archivo.Write(buffer[:n])
				if aerr != nil{
					//eliminar el archivo porque no se pudo crear correctamente
					log.Println(aerr)
					break
				}

				if n < BUFFER_TAMANIO {
					log.Println("Archivo recivido")
					break
				}
		}
			//los bytes recividos deben coincidir con los enviados
			log.Println("Bytes recividos", cuenta)
			//protocolo de mensajes nomal + msg
			cliente<- "server:" + " archivo recivido"
}

//para crear nombres temporales para el archivo en caso de ser necesario
func nombreTemp()(nombre string){
	return "archivo"
}
