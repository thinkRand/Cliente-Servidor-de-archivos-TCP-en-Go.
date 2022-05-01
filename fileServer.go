package main

import (
	"log"
	"net"
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
	b := make([]byte,100) //un slice de bytes , son 100 caractares maximo por lectura
	//Porque no se reconove a la variable dentro del for?
	for {
		bc := 0 //esta variable debe estar aquí porque no funciona cuando está fuera del for, no se por que
		bc, err := conn.Read(b) //si copy requiere un lector y un escritor le pasaer os.Stdout como el destino
		if err != nil{
			log.Println(err)
			return
		}

		//el error de impresion tiene que ver con el UTF-8
		log.Println("Mensaje: ",(string(b[:bc]))) 
	}
	conn.Close()
}
