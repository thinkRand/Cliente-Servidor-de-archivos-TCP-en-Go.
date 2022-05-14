package main

import (
	Ps "protocolo_simple"
	"log"
	"net"
	"os"
	"bufio"
	"strings"
	"fmt"
	// "io"
)



//6 factoes a tener en cuenta
// reliability
// performance
// responsiveness
// scalability
// security
// capacity



const (
	//buffer de lectura de archivos
	BUFFER_TAMANIO = 1024
)




//para crear listas de clientes asociados a un canal
type Clientes struct{
	lista map[Cliente]bool
}

//El canal es donde se registran clientes
//El canal debe escuchar activamente todos los mensajes que se envian a el, tal como en un chat
type Canal struct{
	
	nombre string //el nombre de este canal
	
	clientes map[*Cliente]bool //Registro de los clientes en este canal
	
	escribir chan string  //recive todos los mensajes para escribir en el canal, incluso archivos
	
	unir chan *Cliente //recive todas la peticiones para unirce a este canal

	salir chan *Cliente	//recive todas las peticiones para salir de este canal

}


//Escucha todos los mensajes que llegan a este canal y le da el tratamiento apropiado
//si es un cliente nuevo lo registra
//si un cliente quiere salirce del canal lo elimina del map
//si es un mensaje lo distribuye a todos en el canal
func (can *Canal) Iniciar(){
	log.Println("El", can.nombre," esta abierto")
	for{
		select{
		case cli := <-can.unir:
			log.Println(can.nombre,":: Cliente recivido")
			can.clientes[cli] = true //el canal conoce al cliente ahora
			cli.canal = can //el cliente conoce al canal ahora
		
		case cli := <-can.salir:
			delete(can.clientes, cli)
		
		case msg := <-can.escribir:
			for cli := range can.clientes{
				log.Println(can.nombre,"::redirigiendo",msg)
				cli.escribir<-msg
			}
		}
	}
}


//para almacenar la referencia a todos los canales disponibles
type Canales struct{
	lista map[string]Canal
}

type Cliente struct{
	
	conn net.Conn
	
	escribir chan string //Para escribir mensajes al cliente

	canales *Canales //Lista de canales que existen

	canal *Canal //El canal done este cliente está registrado 
}


//lee los mensajes provenientes del cliente y los manda al interprete de comandos
func (cli *Cliente) Leer(){
	lector := bufio.NewScanner(cli.conn)
	for lector.Scan(){
		entrada := lector.Text()
		log.Println("Recivido::", entrada)
		cli.Interpretar(entrada) //se interpreta uno a la vez, es sincrono, a diferencia de escribir que es asincrono
		//cuando el interprete retorna se vuelven a escuchar entradas desde el cliente
	}
}

//Espera una replica del cliente, un solo mensaje que sera la respuesta para una coordinación interna
//otro nombre puede ser msgCoordinacionInterna
func (cli *Cliente) respuestaCliente() (s string){
	lector := bufio.NewScanner(cli.conn)
	if lector.Scan(){
		entrada := lector.Text()
		return entrada
	}
	return "Error de lectura en replicaCliente()"
}

//Toma todos los mensaje para este cliente y se los envía
func (cli *Cliente) Escribir(){
	for msg := range cli.escribir{
		_, err := fmt.Fprintln(cli.conn, msg) //se ignoran los errores
		if err != nil{
			log.Println(err)
			continue //hay que seguir escuchando pese a los errores
		}
		log.Println("Despachado::", msg)
	}
}


//Recive los mensajes del cliente y reconocer los comandos para ofrecer el procedimiento adecuado
//deberia seraprace en dos partes ioServidor y servicios. Pero no hay tiempo por ahora.
func (cli *Cliente) Interpretar(entrada string){
	
	comando := strings.Split(entrada, " ")
	//formato: [comando] [...valores]
	switch(comando[0]){

	case Ps.C_UNIR_CANAL:
		
		if _, ok := cli.canales.lista[comando[1]]; !ok {	//true si el canal no existe
			log.Println("El canal no existe, lo que existe es", cli.canales.lista)
			cli.escribir<- Ps.S_UNIR_RECHAZADO //se responde el mensaje adecuado
		}
		cli.canales.lista[comando[1]].unir<- cli //se registra al cliente en el canal
		
		cli.escribir<- Ps.S_UNIR_ACEPTADO	//se responde al cliente el mensaje adecuado
		
	
	case Ps.C_ENVIAR_ARCHIVO:
		
		//se espera enviar canal archivo peso
		if len(comando) != 4{
			cli.escribir<- Ps.S_ENVIO_RECHAZADO
		} 

		if cli.canal == nil{
			log.Println("El cliente no esta en un canal")
			cli.escribir<- Ps.S_ENVIO_RECHAZADO
			return
		}

		recivirArchivo(cli, comando)
	
	default:
		// log.Println("Peticion invalida:", entrada)
		// cli.escribir<- Ps.S_CODIGO_INVALIDO //se responde el mensaje adecuado
	
		log.Println("Se presume entrada para el canal")
		if cli.canal != nil{
			cli.canal.escribir<- strings.ToUpper(entrada) //echo xd
		}else{
			log.Println("El cliente no esta en un canal")
			cli.escribir<- "no estas en un canal viejo..."
		}
	}
	return
}


	
func main(){
	ln, err := net.Listen("tcp",":9999")
	if err 	!= nil{
		log.Fatal(err)
	}

	canal1 := Canal{
		nombre: "canal1",
		clientes: make(map[*Cliente]bool),
		escribir: make(chan string),
		unir: make(chan *Cliente),
		salir: make(chan *Cliente),
	}	

	go canal1.Iniciar()

	canales := Canales{
		lista: make(map[string]Canal),
	}
	
	canales.lista[canal1.nombre] = canal1 
	
	log.Println("Servidor activo...")
	for{
		conn, err := ln.Accept()
		if err !=  nil{
			//ingnoro las conexines fallidas
			continue
		}
		go handlerCliente(conn, canales)
	}
}


//handler del cliente 
func handlerCliente(conn net.Conn, canales Canales){
	cli := Cliente{
		conn: conn,
		escribir: make(chan string),
		canales: &canales,
		//canal //no registrado en un canal por ahora
	}

	defer cli.conn.Close() 	
	log.Println("Cliente conectado: ", cli.conn.RemoteAddr().String())

	go cli.Escribir() //para enviar los mensajes al cliente
	cli.Leer() //para recivir los mensajes del cliente
	log.Println("El cliente ",cli.conn.RemoteAddr().String(),"se desconecto")
}


//Funcion de prueba que maneja la subida de un archivo desde el cliente
func recivirArchivo(cli *Cliente, entrada []string){
	//El formato es enviar canal archivo peso
			// canal := entrada[1]
			nombreAr := entrada[2]
			pesoAr := entrada[3] //el valor se recive en string

			archivo, err := os.Create("servidor_archivos/"+nombreAr)
			if err != nil{
				log.Println(err)
				cli.escribir<- Ps.S_ENVIO_RECHAZADO
				return
			}
			defer archivo.Close() 
			buffer := make([]byte, BUFFER_TAMANIO)
			
			cli.escribir<- Ps.S_ENVIO_APROBADO 

			//tengo que escuchar la respuesta del cliente
			log.Println("Esperando el archivo...")
			
			var cuenta int 
			for {
				n, err := cli.conn.Read(buffer)
				cuenta+=n

				if err != nil{
					//EOF ocurre si hay una perdida de conexión
					defer os.Remove("servidor_archivos/"+nombreAr) //se debe eliminar el archivo
					log.Println(err)
					return
				}
			
				_, aerr := archivo.Write(buffer[:n])
				if aerr != nil{
					defer os.Remove("servidor_archivos/"+nombreAr) //eliminar el archivo porque no se pudo crear correctamente
					log.Println(aerr)
					return
				}

				if n < BUFFER_TAMANIO {
					log.Println("Trasmision terminada, se presume recepcion completada")
					// log.Println("Todos los bytes", buffer[:n])
					archivo.Close()
					break
				}
		}

		//notifico al cliente que se recivo todo el archivo, se espera confirmacion para saber si el archivo fue enviado
		cli.escribir<- Ps.S_ARCHIVO_RECIVIDO + " " + "Archivo recivido, bytes totales: " + fmt.Sprintf("%d",cuenta)
			
		
		//Esperar que el cliente envie la confirmacion de haber enviado el archvio completo
		log.Println("Esperando confirmacion del cliente...")
		
		b := make([]byte, 128)
		n, err := cli.conn.Read(b)
		if err != nil{
			os.Remove("servidor_archivos/"+nombreAr)
			log.Println(err)
			return
		}

		// log.Println("Debug:: bytes recividos", b[:n])
		temp := string(b[:n])
		// log.Println("Debug::temp:", temp)
		recivido := strings.Trim(temp, "\n")
		// log.Println("Debug::con triem", recivido)
		if recivido != Ps.C_ARCHIVO_ENVIADO {
			log.Println("El cliente envio el siguiente mensaje:", recivido)
			os.Remove("servidor_archivos/"+nombreAr)
			return
		}
		log.Println("El cliente confirmo que envio el archivo")

		//Verificar la integridad del archivo recivido
		log.Println("Bytes esperados,", pesoAr)
		log.Println("Bytes recividos", cuenta)
		if pesoAr != fmt.Sprintf("%d", cuenta) { //comparación de strings
				
			log.Println("Error. Los bytes recividos no son iguales a los bytes enviados por el cliente")
			cli.escribir<- Ps.S_IMPOSIBLE_CONTINUAR + " " + "Los bytes no coinciden" + fmt.Sprintf("%d",cuenta)
			os.Remove("servidor_archivos/"+nombreAr) //mejor usar archivo.path() o algo asi
			
			return
		}

		log.Println("Lado del servidor cerrado. Escuchando nuevas peticiones")
			//fin de prueba
}
