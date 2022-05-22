package main

import (
	"log"
	"net"
)


//6 factoes a tener en cuenta
// reliability
// performance
// responsiveness
// scalability
// security
// capacity




//para crear listas de clientes asociados a un canal
type Clientes struct{
	lista map[Cliente]bool
}

//para almacenar la referencia a todos los canales existentes
type Canales struct{
	lista map[string]Canal
}

func main(){
	ln, err := net.Listen("tcp",":9999")
	if err 	!= nil{
		log.Fatal(err)
	}

	canal1 := Canal{
		nombre: "canal1",
		ocupado:false,
		clientes: make(map[*Cliente]bool),
		grupoReceptor:make(map[*Cliente]bool),
		unir: make(chan *Cliente),
		salir: make(chan *Cliente),
		listo: make(chan bool),
		archivoParaRedirigir:&archivo{},
	}	
	go canal1.Iniciar()

	
	canales := Canales{
		lista: make(map[string]Canal),
	}
	canales.lista[canal1.nombre] = canal1 
	
	
	log.Println("Servidor activo...")
	for{
		conn, err := ln.Accept()
		if err != nil{
			continue //ingnoro las conexines fallidas
		}
		go handlerCliente(conn, canales)
	}
}


//Inicia a un nuevo cliente
func handlerCliente(conn net.Conn, canales Canales){

	cliente := Cliente{
		conn: conn,
		escribir: make(chan string),
		recivir: make(chan string),
		archivo: make(chan *archivo),
		canales: &canales,
		listo:make(chan bool),
		canal: nil ,
	}
	defer cliente.conn.Close() 	
	
	log.Println("Nuevo cliente conectado:",cliente.conn.RemoteAddr().String())
	go cliente.EnviarArchivos()
	go cliente.Escribir()
	cliente.Leer()
	close(cliente.escribir) //Cerrar este canal termina con la subrutina Escribir()
	close(cliente.archivo) //Cerrar este canal termina la subrutina RecivirArchivo()
	log.Println("El cliente",cliente.conn.RemoteAddr().String(),"se desconecto")

}
