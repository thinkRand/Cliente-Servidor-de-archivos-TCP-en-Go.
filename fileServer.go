package main 

import (
	"net"
	"log"
	"bufio"
	"io"
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


func hCliente(conn net.Conn){
	//Creo un reader sobre la conexion, no un scaner
	//bueno pero cómo leo los datos de la conexion
	var msg string
	reader := conn.Read()
	for {
		err := io.Copy(msg, reader) //si copy requiere un lector y un escritor le pasaer os.Stdout como el destino
		if err != nil{
			log.Println(err)
			return
		}
		log.Println(msg) 
	}
	conn.Close()
}
