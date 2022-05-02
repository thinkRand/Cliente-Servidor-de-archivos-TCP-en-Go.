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

//para crear nombres temporales para el archivo en caso de ser necesario
func nombreTemp()(nombre string){
	return "archivo"
}

//runtina para comunicarce con el cliente
func hCliente(conn net.Conn){
	log.Println("Cliente conectado: ", conn.RemoteAddr().String())

	//creo el archivo que voy a llenar
	arlocal, err := os.Create("archivo")
	if err != nil{
		log.Println("No se pudo crear la ruta local para el archivo")
	}
	defer arlocal.Close() 
	//io.Copy no funcina apropiadamente para leer de la conexion y llenar el archivo
	buffer := make([]byte,1024)
	//read lee de la conexion de forma infinita, para que se detenga en caso de EOF
	FDT := "<FDT>" //Fin de Transmisión
	r := io.Reader(conn)
	var count int
	for {
		n, err := r.Read(buffer)
		count+=n
		// log.Print(n)
		//leo los ultimos 5 bytes para ver si se presenta FDT
		//<FDT> tiene 5 bytes
		if string(buffer[n-5:n]) == FDT{
			_, aerr := arlocal.Write(buffer[:n-5])
			if aerr != nil{
				//eliminar el archivo porque no se pudo crear correctamente
				log.Println(aerr)
				break
			}
			log.Println("Archivo recivido")
			break
		}

		if err != nil{
			//la condicio io.EOF no funciona. Debe ser porque en una conexion no funciona la condicion io.EOF, o esa condicion no se envia desde el cliente.
			//a persar de que no funciona la dejo aquí para ver si puedo hacer que funcione
			if n == 0 && err == io.EOF{
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
			//eliminar el archivo porque no se pudo crear correctamente
			log.Println(aerr)
			break
		}
}
	//los bytes recividos deben coincidir con los enviados
	log.Println("Bytes recividos", count)




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
	log.Println("Servidor cerrado")
}
