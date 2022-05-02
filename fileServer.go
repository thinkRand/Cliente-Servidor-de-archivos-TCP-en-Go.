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

const BUFFER_TAMANIO int = 1024
//runtina para comunicarce con el cliente
func hCliente(conn net.Conn){
	defer conn.Close() //me aseguro de cerra la conexion en cualquier caso
	log.Println("Cliente conectado: ", conn.RemoteAddr().String())
	canal := make(chan string)
	go responderCliente(conn, canal) //para enviar las respuestas del servidor al cliente
	lector := bufio.NewScanner(conn)
	for lector.Scan(){
		entrada := lector.Text()
		interprete(entrada, conn, canal)
	}
	log.Println("Servidor cerrado")
}

func responderCliente(conn net.Conn, chanel <-chan string){
	for msg := range chanel{
		//esta función formatea el string de msg de forma estandar y lo escribe en la conexion
		log.Println("Se escribio el mensaje en al conexion")
		fmt.Fprintln(conn, msg) //ingnoro los errores
	}
}

func interprete(entrada string, conn net.Conn, canal chan string){
	var comando []string
	//msg guarda el texto mientras recorro todo el contenido en el canal
	log.Println(entrada)
	comando = strings.Split(entrada, " ")
	switch(comando[0]){
	case "subir":
		log.Println("Comando subir recivido")
		canal<- "server: Comando subir reconocido, espera mientras mientras se arregla todo en el servidor"
		subirArchivo(conn, canal)
	case "obtener":
		log.Println("comando obtener recivido")
	default:
		log.Println(entrada, "no es un comando valido")
	}
	return
}


func subirArchivo(conn net.Conn, canal chan string){
		archivo, err := os.Create("archivo")
		if err != nil{
			log.Println("No se pudo crear la ruta local para el archivo")
			canal <- "Error en el servidor, no se puede subir el archivo"
			return
		}
		defer archivo.Close() 
		canal <- "server: Listo parar recivir archivo"
		buffer := make([]byte, BUFFER_TAMANIO)
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
}