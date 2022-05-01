package main

import (
	"log"
	"net"
	"os"
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


//runtina para comunicarce con el cliente
func hCliente(conn net.Conn){
	log.Println("Cliente conectado")
	// b := make([]byte,100) //un slice de bytes , son 100 caractares maximo por lectura
	//Porque no se reconove a la variable dentro del for?
	arlocal, err := os.Create("archivo")
	if err != nil{
		log.Println("No se pudo crear la ruta local para el archivo")
	}

	//parece que io.Copy no funcion apropiadamente para tomar los bytes provenientes del clinte y asignarlos al archivo
	//usare read
	buffer := make([]byte,1024)
	// var bc int //aqui se declara y luego se vuelve a declarar abajo
	for {
		bc, err := conn.Read(buffer) //lee bytes de 1024 en 1024
		if err != nil{
			log.Println(err)
			return
		}
		
		n, err := arlocal.Write(buffer[:bc])
		if err != nil{
			if n != len(buffer){
				log.Println("no se escribieron todos los bytes")
			} 
			log.Println("fallo al escribir el archivo")
		return
		}
		
	}

	log.Println("Archivo recivido")
	// 	//el error de impresion tiene que ver con el UTF-8
	// 	log.Println("Mensaje: ",(string(b[:bc]))) 
	// }
	conn.Close()
}
