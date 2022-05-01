package main

import (
	"log"
	"net"
	"os"
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


//runtina para comunicarce con el cliente
func hCliente(conn net.Conn){
	log.Println("Cliente conectado")

	//creo un nombre temporal para el archivo
	arlocal, err := os.Create("archivo")
	if err != nil{
		log.Println("No se pudo crear la ruta local para el archivo")
	}
	defer arlocal.Close()
	//parece que io.Copy no funcion apropiadamente para tomar los bytes provenientes del clinte y asignarlos al archivo
	buffer := make([]byte,1024)
	// var bc int //aqui se declara y luego se vuelve a declarar abajo
	//read lee de la conexion de forma infinita, para que se detenga en caso de EOF
	FDT := "<FDT>" //Fin de Transmisión
	r := io.Reader(conn)
	for {
		n, err := r.Read(buffer)
		//<FDT> tiene 5 bytes
		if string(buffer[n-5:n]) == FDT{
			_, aerr := arlocal.Write(buffer[:n-5])
			if aerr != nil{
				log.Println(aerr)
				break
			}
			log.Println("Archivo recivido")
			break
		}

		if err != nil{
			//la condicio io.EOF no funciona, debe ser porque es una conexion y esa condicion no se envia desde el cliente
			//a persar de que no funciona la dejo aquí para ver si puedo hacer que funcione
			if n == 0 && err == io.EOF{
				//los ultimos bytes antes del erro o EOF
				// arlocal.Close()
				log.Println("Archivo recivido")
				break
			}
			//Eliminar archivo porque no se pudo cargar por completo
			//Puede ser un error de conexion
			log.Println(err)
			break
		}
	
		_, aerr := arlocal.Write(buffer[:n])
		if aerr != nil{
			log.Println(aerr)
			break
		}
		

	}





	// for {
	// 	bc, err := conn.Read(buffer) //lee bytes de 1024 en 1024
	// 	if err != nil{
	// 		log.Println(err)
	// 		//elimino el archivo temporal actual
	// 		return
	// 	}
		
	// 	n, err := arlocal.Write(buffer[:bc])
	// 	if err != nil{
	// 		// if n != len(buffer){
	// 		// 	log.Println("no se escribieron todos los bytes")
	// 		// } 
	// 		log.Println("fallo al escribir el archivo")
	// 		return
	// 	}
	// 	log.Println("bytes escritos: ", n)	
	// }

	// log.Println("Archivo recivido")
	// 	//el error de impresion tiene que ver con el UTF-8
	// 	log.Println("Mensaje: ",(string(b[:bc]))) 
	// }
	conn.Close()
}
