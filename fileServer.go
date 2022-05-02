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

//para crear nombres temporales para el archivo en caso de ser necesario
func nombreTemp()(nombre string){
	return "archivo"
}

//runtina para comunicarce con el cliente
func hCliente(conn net.Conn){
	log.Println("Cliente conectado: ", conn.RemoteAddr().String())

	arlocal, err := os.Create("archivo")
	if err != nil{
		log.Println("No se pudo crear la ruta local para el archivo")
	}
	defer arlocal.Close() 
	
	buffer := make([]byte,1024)
	var count int
	for {
		n, err := conn.Read(buffer)
		count+=n

		if err != nil{
			//posible desconexion
			//se debe eliminar el archivo
			log.Fatal(err)
		}
	
		_, aerr := arlocal.Write(buffer[:n])
		if aerr != nil{
			//eliminar el archivo porque no se pudo crear correctamente
			log.Println(aerr)
			break
		}

		if n < 1024 {
			log.Println("Archivo recivido")
			break
		}
}
	//los bytes recividos deben coincidir con los enviados
	log.Println("Bytes recividos", count)




	conn.Close()
	log.Println("Servidor cerrado")
}
