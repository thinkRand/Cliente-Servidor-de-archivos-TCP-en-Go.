package main

import (
	"log"
	"net"
	"os"
	"bufio"
	"strings"
	"fmt"
)


//primero voy a lograr establecer una conexion entre el servidor y el cliente para luego mandar un texto
func main(){
	ln, err := net.Listen("tcp",":8080")
	if err 	!= nil{
		log.Fatal(err)
	}
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
	//en
	for msg := range cliente{
		//esta función formatea el string de msg de forma estandar y lo escribe en la conexion
		//le evio el msg al cliente
		fmt.Fprintln(conn, msg) //ingnoro los errores
		log.Println("Se escribio el mensaje",msg,"en la conexion")
	}
}

func interprete(entrada string, conn net.Conn, cliente chan string){
	//esto lo are un for con un canal para sincronisar
	var comando []string
	//msg guarda el texto mientras recorro todo el contenido en el cliente
	log.Println(entrada)
	comando = strings.Split(entrada, " ")
	switch(comando[0]){
	case "up":
		log.Println("Solicitud up recivida desde cliente")
		recivirArchivo(conn, cliente)
	case "obtener":
		log.Println("comando obtener recivido")
	default:
		log.Println(entrada, "no es un comando valido")
	}
	return
}


func recivirArchivo(conn net.Conn, cliente chan<- string){
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
		cliente<- "server: archivo recivido"
}