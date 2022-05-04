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
	puedo tener un canal global para recivir a todos los clientes
	que se conectan al servidor y estos se registren.

	Se me ocurre un contenedor de tipo canal, que almacene todos los clientes conectados
	a el, que tenga sus propias rutinas para recivir y enviar comandos

	inicialmente voy a lograr que se transmitan mensajes para luego pasar a archivos

	entonces como lo hago?

	primero quién crea el canal? 
	primero lo voy a hacer de forma manual

	canal1 sera su nombre y  lo probare con tres clientes, dos clientes en el canal y uno fuera de el.

	ahora como creo el canal?
		Los cliente piden al servidor conectarce, luego solicitan unirce al canal 1
		entonces los mensajes de este usuario son dirigidos al canal que solicito

		cuando el servidor recive una peticion de un cliente para unirse a un canal lo registra en ese
		canal el cual es un map[cli]bool o un mapa de clientes con valores booleanos

		tan sencillo como eso, el canal es un map con todos los clientes registrados en el

		y clave está en su funcionamiento.

		al unirce a un canal(anotarce en el map) todos los mensajes que envien deben ser enviados al canal
		por eso debe haber una forma de saber a que canal esta conectado un cliente.

		map[cliente] := canal al que esta conectado


		el servidor puede hacer map[cliente] para saber cual es el canal al que enviara los mensajes.
		En este caso el canal 1

		si el cliente no esta registrado en un canal, el servidor responde de forma estandar con
			<comando invalido
			<solo puedes enviar archivos en canales
			<error de conexion

*/

/*
	con lo que tengo hasta ahora puedo decir que

	si el cliente pide unirce a un canal aria

	unir canal1

	el servirdor interpreta el comando y efectivamente hace
		canal1.clientes[cliente.escribir] = true //le paso el canal para escribir al cliente

	de esta forma cuando el handler de este cliente reciva una peticion 

	enviar canal1 "hola a todos"

	el servidor interpreta enviar y lansa lo siguiente
		if cliente.canales[canal1] == true //si esta registrado en el canal que solicito
			//entonces comunico el mensaje recivido a todos en el canal 
			//canal1.escribir <- msg



			}

type canal struct{
	escribir chan string

}

canal1 debe tener una rutina broadcaster para escribir
lo de aquí abajo sirve para hacer go canal1.escribir() que activa la subrutina que escucha cada mensaje de los clientes 
en el canal
func (*canal) escribir(){
	for msg := range <-canal.escribir{
		for cliente := range canal.clientes{
			cliente.escribir <- msg 
		}
	}	
}


*/



type cliente struct{
	
	leer chan string //para leer mensajes del cliente
	
	escribir chan string //para escribir mensajes al cliente

	canales map[struct{}]bool  //un mapa de canales donde esta registrado
}


type canal struct{
	clientes map[chan string]bool //el canal para escribir de los clientes
}

//canal1 es conocido por todos pero no esta inicialisado, tiene todos sus valores en blanco
var canal1 canal

func main(){
	ln, err := net.Listen("tcp",":8080")
	if err 	!= nil{
		log.Fatal(err)
	}
	canal1.clientes = make(map[chan string]bool)
	log.Println("Servidor activo...")
	for{
		conn, err := ln.Accept()
		if err !=  nil{
			//ingnoro las conexines fallidas
			continue
		}
		go hCliente(conn)
	}
}

//para crear nombres temporales para el archivo en caso de ser necesario
func nombreTemp()(nombre string){
	return "archivo"
}

//runtina para comunicarce con el cliente
func hCliente(conn net.Conn){
	defer conn.Close() //me aseguro de cerra la conexion en cualquier caso
	log.Println("Cliente conectado: ", conn.RemoteAddr().String())
	cliente := make(chan string) //se reciven mensajes de las rutinas para eviarlas al cliente
	go responderCliente(conn, cliente) //para enviar las respuestas del servidor al cliente
	lector := bufio.NewScanner(conn)
	for lector.Scan(){
		entrada := lector.Text()
		interprete(entrada, conn, cliente)
		//cuando el interprete retorna se vuelven a escuchar entradas desde el cliente
	}
	log.Println("Servidor cerrado")
}

func responderCliente(conn net.Conn, cliente <-chan string){
	for msg := range cliente{
		//esta función formatea el string de msg de forma estandar y lo escribe en la conexion
		//le evio el msg al cliente
		fmt.Fprintln(conn, msg) //ingnoro los errores
		log.Println("---->", msg)
	}
}

func interprete(entrada string, conn net.Conn, cliente chan string){
	//primer paso, cuando el servidor reconoce el comando unir debe registrar al canal del cliente
	//en el canal correspondiendte

	var comando []string
	comando = strings.Split(entrada, " ")
	
	switch(comando[0]){
	case "unir":
		if (comando[1] == "canal1"){
			canal1.clientes[cliente] = true
			cliente<- "server: " + " Estas registrado en " + comando[1] +"\n"
		}else{
			cliente<- "server: "+ comando[1] + " no es un canal\n"
		}
		//el msg debe venir en for up espacio nombreCanal
	case "up":
		log.Println("Solicitud up recivida desde cliente")
		//si el nombre del canal esta registrado en el registro de canales del cliente
		//algo como cliente.canales[comando[1]] == true
		cliente<- "server: "+ "recivida la peticion >up< en el canal "+ comando[1] + "\n"
		recivirArchivo(conn, cliente)
	case "obtener":
		log.Println("comando obtener recivido")
	default:
		log.Println(entrada, "no es un comando valido")
	}
	return
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