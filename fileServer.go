package main 

import (
	"net"
	"log"
	"bufio"
)

//primero voy a lograr establecer una conexion entre el servidor y el cliente para luego mandar un texto
func main(){
	ln, err := net.Listen("tcp",":8080")
	if err 	!= nil{
		log.Fatal(err)
	}

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
	entrada := bufio.NewScanner(conn)
	for entrada.Scan(){
		log.Println(entrada.Text())
	}
	conn.Close()
}
