package main

import (
	"log"
	"net"
	"os"
	"bufio"
	"strings"
	"fmt"
)



//6 factoes a tener en cuenta
// reliability
// performance
// responsiveness
// scalability
// security
// capacity


//la implementación del protocolo se debe encargar de formatear los mensajes.
//Me refiero a que cualquier mensaje como "canal aprobado" debe ser reconocido
//si el comando es "canal apropiado" y la regla del protocolo exige que primero va el comando y
//luego va la carga entonces reconocerio mal el comando anterior.
//el protocolo debería ser un paquete aparte

type protocolo struct{
	//constantes de mensajes

	//regla del protocolo. e.j [cmd][opciones][data]

	//methodo Dial, Listen, Accept, para trabajar con un objeto protocol.Conn
}

//con estas constantes puedo cambiar el protocolo son facilidad
const (
	//RESPUESTAS DEL SERVIDOR
	SERVIDOR_CANAL_APROBADO = "canalaprobado"
	SERVIDOR_CANAL_NOAPROBADO = "canalanoprobado"
	SERVIDOR_SALIR_APROBADO = "saliraprobado"
	SERVIDOR_CONEXION_APROBADO = "conexionaprobada"
	SERVIDOR_ENVIO_APROBADO = "envioaprobado"
	SERVIDOR_ENVIO_NOAPROBADO = "envionoaprobado"
	SERVIDOR_ERROR_CMD = "El comando es invalido"
	SERVIDOR_MSG = "msg" //para crear mensajes estandar sin relevancia para la coordinación, su destion es la pantalla del cliente


	//PETICIONES DEL CLIENTE
	CLIENTE_UNIR_CANAL = "unir"
	CLIENTE_SALIR_CANAL = "salir"
	CLIENTE_CONEXION = "establecerconexion"
	CLIENTE_ENVIAR_ARCHIVO = "enviararchivo"

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
	//el nombre de este canal
	nombre string 

	//Registro de los clientes en este canal
	clientes map[*Cliente]bool

	//recive todos los mensajes para escribir en el canal, incluso archivos
	escribir chan string 

	//recive todas la peticiones para unirce a este canal
	unir chan *Cliente

	//recive todas las peticiones para salir de este canal
	salir chan *Cliente
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
			can.clientes[cli] = true
		
		case cli := <-can.salir:
			delete(can.clientes, cli)
		
		case msg := <-can.escribir:
			for cli := range can.clientes{
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
	
	escribir chan string //para escribir mensajes al cliente

	canales *Canales //para conocer a todos los canales disponibles

}



//lee los mensajes provenientes del cliente y los manda al interprete de comandos
func (cli *Cliente) Leer(){
	lector := bufio.NewScanner(cli.conn)
	for lector.Scan(){
		entrada := lector.Text()
		cli.Interpretar(entrada) //un interprete por cliente
		//cuando el interprete retorna se vuelven a escuchar entradas desde el cliente
	}
}


//Toma todos los mensaje para este cliente y se los envia
func (cli *Cliente) Escribir(){
	for msg := range <-cli.escribir{
		fmt.Fprintln(cli.conn, msg) //se ignoran los errores
	}
}


//Recive los mensajes del cliente y reconocer los comandos para ofrecer el procedimiento adecuado
func (cli *Cliente) Interpretar(entrada string){
	comando := strings.Split(entrada, " ")
	//segun el protocolo la primera palabra es el comando
	//el formato es comando valor valor
	switch(comando[0]){

	case CLIENTE_UNIR_CANAL:
		log.Println("comando",CLIENTE_UNIR_CANAL,"recivido")
		if _, ok := cli.canales.lista[comando[0]]; !ok {	//true si el canal no existe
			cli.escribir<- SERVIDOR_CANAL_NOAPROBADO
		}
		cli.canales.lista[comando[0]].unir<- cli	
	
	case CLIENTE_ENVIAR_ARCHIVO:
		log.Println("comando",CLIENTE_ENVIAR_ARCHIVO,"recivido")
		//tengo que comprobar el formato de los mensajes
		//debe hacer como mínimo comando-nombrecanal-nombreArchivo
		if len(comando) != 3{
			cli.escribir<- SERVIDOR_ENVIO_NOAPROBADO
		} 

		recivirArchivo(cli, comando)
	default:
		log.Println(entrada, SERVIDOR_ERROR_CMD)
	}
	return
}


	
func main(){
	ln, err := net.Listen("tcp",":8080")
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
	}

	//invalid memory addres or pointer nil dereference
	defer cli.conn.Close() 	
	log.Println("Cliente conectado: ", cli.conn.RemoteAddr().String())

	go cli.Escribir() //para enviar los mensajes al cliente
	cli.Leer() //para recivir los mensajes del cliente
	log.Println("El cliente ",cli.conn.RemoteAddr().String(),"se desconecto")
}


//maneja la subida de un archivo desde el cliente
func recivirArchivo(cli *Cliente, entrada []string){
	//entrada debe ser cmd / canal / nombre archivo
			archivo, err := os.Create(entrada[3])
			if err != nil{
				log.Println("No se pudo crear la ruta local para el archivo",entrada[3])
				cli.escribir<- SERVIDOR_ENVIO_NOAPROBADO
				return
			}
			defer archivo.Close() 
			buffer := make([]byte, BUFFER_TAMANIO)
			
			//servidor listo para recivir archivo
			cli.escribir<- SERVIDOR_ENVIO_APROBADO
			var cuenta int 
			for {
				n, err := cli.conn.Read(buffer)
				cuenta+=n

				if err != nil{
					//posible desconexion
					//se debe eliminar el archivo
					defer os.Remove(entrada[3])
					log.Fatal(err)
					return
				}
			
				_, aerr := archivo.Write(buffer[:n])
				if aerr != nil{
					//eliminar el archivo porque no se pudo crear correctamente
					defer os.Remove(entrada[3])
					log.Fatal(aerr)
					return
				}

				if n < BUFFER_TAMANIO {
					log.Println("Archivo recivido")
					break
				}
		}
			//los bytes recividos deben coincidir con los enviados
			log.Println("Bytes recividos", cuenta)
			//protocolo de mensajes normales: msg(cmd) + UnMensaje
			cli.escribir<- SERVIDOR_MSG + " " + "Archivo recivido, bytes totales: " + fmt.Sprintf("%T",cuenta)
}

//para crear nombres temporales para el archivo en caso de ser necesario
func nombreTemp()(nombre string){
	return "archivo"
}


//el servidor es un canal en si mismo que permite redirigir el flujo de bytes o almacena
//el archivo localmente para luego enviarlo sobre el canal, en cuyo caso lo debe almacenar.
//me gusta más la opción uno.
//lo que pasa con la opción uno es que el flujo de datos se debe copiar a cada cliente por lo que
//necesito una forma para copiar un stream y luego copiarlo a cada cliente
//al como:
// n, err := conn.read(stream)
//	for everu cli{
//		cli.escribir<- stream //escribir es un rutina aparte que escribe al cliente
//	}
//eveftivamente es un de distribucion  multiplex en el que entra un stream y sale cli*streams
//
//El siguiente problema es de saruracion del canal cuando dos cliente quieren transmitir un archivo al mismo
//tiempo. Como deberiá transportarce la información sobre el canal?
//Pues este problema no es nuevo y seguro ya tiene una solución estandar.
//cuando muchos clientes intentan enviar bytes al mismo canal se genera un cuello de botella
//El proceso es más o menos así:
//cliente>>>bytes>>>rutinaRead>>>>única Rutina readCanal>>>Única runtina writeCanal

//hay esta el mayor cuello de botella jamás antes visto


func servirArchivo(path string, canal string){
	//abro el archivo
	//lo envio al canal
	//cierro el archivo
}



func respuestaCliente(clienteWriter string, peticion string){
	//responder al cliente segun al petición
	//seguramente con cli.interpretar
}