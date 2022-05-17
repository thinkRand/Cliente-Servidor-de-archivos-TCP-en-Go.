package main

import (
	Ps "protocoloSimple"
	"log"
	"net"
	"os"
	"bufio"
	"strings"
	"fmt"
	"io"
)



//6 factoes a tener en cuenta
// reliability
// performance
// responsiveness
// scalability
// security
// capacity



const (
	BUFFER_RCV_ARCHIVO = 1024 //buffer de lectura de archivos
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

	listo chan bool //para coordinar

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
		
			can.clientes[cli] = true //el canal conoce al cliente ahora
			cli.canal = can //el cliente conoce al canal ahora //esto beria ser responsabilidad del cliente, no del canal
			log.Println(can.nombre,"::Se registro un cliente")
			can.listo<- true //terminado, esto paralisa al canal, y como varios clientes estan en el mismo canal, si el canal se paralisa procesando a un cliente todos los demas clientes se bloquean tambien. corregir
		
		case cli := <-can.salir:
		
			delete(can.clientes, cli) //el canal ya no conoce al cliente
			cli.canal =  nil //el cliente ya no conoce a un canal //esto debe ser responsabilidad del cliente, no del canal
			can.listo<- true //terminado, esto paralisa al canal, y como varios clientes estan en el mismo canal, si el canal se paralisa procesando a un cliente todos los demas clientes se bloquean tambien. corregir
		
		case msg := <-can.escribir:

			log.Println(can.nombre,"redirigiendo",msg)
			for cli := range can.clientes{
				cli.escribir<- can.nombre+"::"+msg
				<-cli.listo //terminado, esto paralisa al canal, y como varios clientes estan en el mismo canal, si el canal se paralisa procesando a un cliente todos los demas clientes se bloquean tambien. corregir
			}
			can.listo<-true
			log.Println(can.nombre,"msg redirigido!")
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

	listo chan bool //para sincronisar 

	canales *Canales //Lista de canales que existen

	canal *Canal //El canal done este cliente está registrado 
}


//lee los mensajes provenientes del cliente y los manda al interprete de comandos
//esta rutina termina si el la conexion del cliente se desase
func (cli *Cliente) Leer(){
	lector := bufio.NewScanner(cli.conn)
	for lector.Scan(){
		entrada := lector.Text()
		log.Println("Recivido::", entrada)
		cli.Interpretar(entrada) //se interpreta uno a la vez, es sincrono, a diferencia de escribir que es asincrono
		log.Println("Escuchando nuevas peticiones...")
	}
}


//Toma todos los mensaje para este cliente y se los envía
//esta rutina termina si el canal para escribir se cierra
func (cli *Cliente) Escribir(){

	for msg := range cli.escribir{
		_, err := fmt.Fprintln(cli.conn, msg) //se ignoran los errores
		if err != nil{
			log.Println(err)
			cli.listo<- true //terminado
			continue //hay que seguir escuchando pese a los errores
		}
		log.Println("Despachado::", msg)
		cli.listo<- true //terminado
	}

}


//Recive los mensajes del cliente y reconocer los comandos para ofrecer el procedimiento adecuado
//deberia seraprace en dos partes ioServidor y servicios. Pero no hay tiempo por ahora.
//cana comando deberia tener su propia funsion en el ioServidor.
func (cli *Cliente) Interpretar(entrada string){
	
	comando := strings.Split(entrada, " ")
	//formato: [comando] [...valores]
	switch(comando[0]){

	case Ps.C_UNIR_CANAL:
		
		if _, ok := cli.canales.lista[comando[1]]; !ok {	//true si el canal no existe
			log.Println("El cliente pidio unirce a un canal que no existe")
			cli.escribir<- Ps.S_UNIR_RECHAZADO 
			<-cli.listo //espero que cli.escribir se termine de procesar
			return
		}
		
		cli.canales.lista[comando[1]].unir<- cli //se registra al cliente en el canal
		<-cli.canales.lista[comando[1]].listo //verifico si el canal termino de registrarlo
		cli.escribir<- Ps.S_UNIR_ACEPTADO	//se responde al cliente el mensaje adecuado
		<-cli.listo //espero que se termine de procesar 
	
	case Ps.C_ENVIAR_ARCHIVO:
		
		//se espera el formato: enviar <espacio> canal <espacio> nombre-archivo <espacion> peso-archivo
		if len(comando) != 4{
			cli.escribir<- Ps.S_ENVIO_RECHAZADO
			<-cli.listo //espero que se termine de procesar
			return 
		} 

		if cli.canal == nil{
			log.Println("El cliente no esta en un canal")
			cli.escribir<- Ps.S_ENVIO_RECHAZADO
			<-cli.listo //espero que se termine de procesar 
			return
		}

		recivirArchivo(cli, comando)
	
	default:
		// log.Println("Peticion invalida:", entrada)
		// cli.escribir<- Ps.S_CODIGO_INVALIDO //se responde el mensaje adecuado
	
		log.Println("Se presume entrada para el canal")
		if cli.canal != nil{
			cli.canal.escribir<- strings.ToUpper(entrada) //echo xd
			<-cli.canal.listo //espero que se termine de procesar 
		}else{
			log.Println("El cliente no esta en un canal")
			cli.escribir<- "no estas en un canal viejo..."
			<-cli.listo //espero que se termine de procesar 
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
		listo: make(chan bool),
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
		listo:make(chan bool),
		//canal //no registrado en un canal por ahora
	}

	defer cli.conn.Close() 	
	log.Println("Cliente conectado: ", cli.conn.RemoteAddr().String())

	go cli.Escribir() //para enviar los mensajes al cliente
	cli.Leer() //para recivir los mensajes del cliente, se cierra si la conexion se pierde
	close(cli.escribir) //cerra el canal de escribir termina con la subrutina Leer()
	log.Println("El cliente ",cli.conn.RemoteAddr().String(),"se desconecto")
	//en este punto la variable cli deja de existir
}


//Funcion de prueba que maneja la subida de un archivo desde el cliente
func recivirArchivo(cli *Cliente, entrada []string){
	//se espera el formato: enviar <espacio> canal <espacio> nombre-archivo <espacion> peso-archivo
	//a partir de ahora la comunicacion necesita sincronisacion
	//no se usaran los canales de escuha del servidor
			// canal := entrada[1]
			nombreAr := entrada[2]
			pesoAr := entrada[3] //el valor se recive en string
			archivo, err := os.Create("servidor_archivos/"+nombreAr)
			if err != nil{
				log.Println(err)
				cli.conn.Write([]byte(Ps.S_ENVIO_RECHAZADO))
				return
			}
			defer archivo.Close() 
			
			buffer := make([]byte, BUFFER_RCV_ARCHIVO) 
			
			cli.conn.Write([]byte(Ps.S_ENVIO_APROBADO)) //debe ser persistente, para eso hay que modificar al bucle que escribe asi que se ará más adelante

			log.Println("Esperando el archivo...")
			var cuenta int 
			
			//ahora el servidor está listo para recivir el archivo
			//El servidor notificara <imposible continuar> si ocurre un error de lectura apartir de ahora
			for {
				n, err := cli.conn.Read(buffer)
				cuenta+=n

				//en caso de error al leer de la conexión
				if err != nil{
					// if true{ //forzar el error
					//EOF ocurre si hay una perdida de conexión
					if err == io.EOF {
						
						archivo.Close() 
						os.Remove("servidor_archivos/"+nombreAr) //La conexion se cerror de forma inesperada asi que borro el archivo
						log.Println(err)
					
						//Borro al cliente completamente
						//Primero lo elimino del canal
						referenciaCanal := cli.canal
						cli.canal.salir<-cli //Elimino el registro del cliente en el canal. Debería ser algo como canal.salir<- cli. Aparte se esta proceso cli.canal = nil, algo que debe ser natural para el cli, no el canal
						<-referenciaCanal.listo //Espera la notificacón del canal sobre sacar al cliente
						//Luego cierro su conexion
						cli.escribir<- Ps.S_IMPOSIBLE_CONTINUAR //No puedo leer de la conexion, pero puedo escribir. Esto es TCP
						<-cli.listo
						cli.conn.Close() //Cerrar la conexion termina las runitinas para Leer() y Escribir del cliente
						return
					}
					
					//Si el error no es por perdida de conexion entonces es un error del lado del servidor
					//Como este error ocurre mientras el cliente sigue enviando datos
					//cada stream que el cliente envie es invalido a partir de ahora 
					//hasta que el cliente y el servidor entiendan que deden iniciar un nuevo proceso
					//por ahora solventare este problema cerrando la conexión por completo

					//Borro al cliente completamente
					//Primero lo elimino del canal
					referenciaCanal := cli.canal
					cli.canal.salir<-cli //Elimino el registro del cliente en el canal. Debería ser algo como canal.salir<- cli. Aparte se esta proceso cli.canal = nil, algo que debe ser natural para el cli, no el canal
					<-referenciaCanal.listo //Espera la notificacón del canal sobre sacar al cliente
					
					//notifíco al cliente del error en el lado del servidor
					cli.escribir<- Ps.S_IMPOSIBLE_CONTINUAR +" "+ "no puedo escuchar"  //No puedo leer de la conexion, pero puedo escribir. Esto es TCP
					<-cli.listo
					cli.conn.Close() //Cerrar la conexion termina las runitinas para Leer() y Escribir() del cliente
					return 
				}


				//en caso de error al escribir en el archivo creado
				_, werr := archivo.Write(buffer[:n])
				if werr != nil{
					log.Println(werr)
					archivo.Close()
					os.Remove("servidor_archivos/"+nombreAr) //eliminar el archivo porque no se pudo crear correctamente
					cli.escribir<- Ps.S_TERMINAR_PROCESO //se espera reply desde el cliente, debe haber deadline
					<-cli.listo
					//descargar todo lo que se reciva hasta que el cliente notifíque un <terminar proceso>
					//por ahora matare la conexion y cerrare todo, luego hay que manejar el error
					referenciaCanal := cli.canal
					cli.canal.salir<-cli //Elimino el registro del cliente en el canal. Debería ser algo como canal.salir<- cli. Aparte se esta proceso cli.canal = nil, algo que debe ser natural para el cli, no el canal
					<-referenciaCanal.listo //Espera la notificacón del canal sobre sacar al cliente
					cli.conn.Close() //Cerrar la conexion termina las runitinas para Leer() y Escribir() del cliente
					return
				}

				//si no se puede leer suficientes bytes es que finalisó la transmisión. Qué pasa si n == BUFFER_RCV_ARCHIVO pero la transmisión termino?
				if n < BUFFER_RCV_ARCHIVO {
					log.Println("Trasmision terminada, se presume recepcion completada")
					archivo.Close()
					break
				}
		}
		//Recepción de archivo terminada, sin embargo el cliente pudo terminar la transmision a propósito
		//para indicar que no pudo continuar durante la transmision. En tal caso el cliente envió información basura
		//para cerrar el bucle de lectura. 
		
		//el caso anterior se verifica con dos pasos: si bytes recividos == bytes esperados
		//y si cliente envió el mensaje <imposible continuar>

		//Comienso verificando la integridad del archivo recivido
		log.Println("Bytes esperados,", pesoAr)
		log.Println("Bytes recividos", cuenta)
		if pesoAr != fmt.Sprintf("%d", cuenta) { //comparación de strings
			
			//el archivo se debe borrar pero antes hay que comunicar al cliente del error y este
			//debe devolver el mensaje <terminar proceso>	
			log.Println("Error. Los bytes recividos no son iguales a los bytes enviados por el cliente")
			cli.escribir<- Ps.S_TERMINAR_PROCESO + " " + "Los bytes no coinciden " + fmt.Sprintf("%d",cuenta)
			<-cli.listo //espero que se termine de procesar 
			log.Println("Esperando respuesta...")
			//ahora se espera la respuesta del cliente con un deadline // cli.conn.SetReadDeadline(time)
			rspCli := make([]byte, 128)
			n, err := cli.conn.Read(rspCli)
			if err != nil{
				//Debería usar persistencia hasta 3 intentos, si no se puede leer al cliente, le intento escribir <no puedo escuchar>

				os.Remove("servidor_archivos/"+nombreAr)
				log.Println(err)
				return 
			}
			temp := string(rspCli[:n])
			rsp := strings.Split(temp, " ")
			log.Println("recivido", temp)
			if rsp[0] == Ps.C_IMPOSIBLE_CONTINUAR{
				//confirmado que hubo un error desde el cliente
				//borrar todo y dejar de esuchar
				log.Println("El cliente confirmo que no pudo continuar")
				os.Remove("servidor_archivos/"+nombreAr)
				log.Println("Lado del servidor termino")
				return 
			}else{
				log.Println("Respuesta inesperada del ciente",temp)
				os.Remove("servidor_archivos/"+nombreAr)
			}

			return
		}



		//se le pregunta al cliente si logro enviar el 
		//archivo correctamente. Si el cliente no puedo terminar enviara el mensaje <imposible continuar>

		
		//notifico al cliente que se reciviio la transmision, se espera confirmacion para saber si el archivo fue enviado
		cli.escribir<- Ps.S_ARCHIVO_RECIVIDO + " " + "Archivo recivido, bytes totales: " + fmt.Sprintf("%d",cuenta)
		<-cli.listo //espero que se termine de procesar 
		log.Println("msg enviado...")	
		
		//Esperar que el cliente envíe la confirmación de haber enviado el archvio completo
		log.Println("Esperando confirmacion del cliente...")
		
		b := make([]byte, 128)
		n, err := cli.conn.Read(b)
		if err != nil{
			//persistencia hasta 3 intentos, si no se puede leer al cliente, le intento escribir <no puedo escuchar>
			//si se escribe cierro la conexion
			//el cliente debe dejar de envair cualquier mensaje tambien y cerrar
			//esto debe pasar si algún estremo no puede escuchar

			os.Remove("servidor_archivos/"+nombreAr)
			log.Println(err)
			return 
		}

		//durante el proceso de recepción de archivos no es necesario el \n
		temp := string(b[:n])
		recivido := strings.Trim(temp, "\n")
		if recivido != Ps.C_ARCHIVO_ENVIADO {
			//si el cliente no envio un <archivo enviado> esto puede ser un
			//<imposible continuar> o un <no puedo escuchar>
			//en caso de imposible continuar cierro todo
			log.Println("El cliente envio el siguiente mensaje:", recivido)
			if recivido == Ps.C_IMPOSIBLE_CONTINUAR{
				log.Println("El cliente no pudo enviar el archivo")
				log.Println("Lado del servidor termino") //mejor un defer
				//el archivo se va a borrar, dejo todo y termino este extremo
				return
			}
			os.Remove("servidor_archivos/"+nombreAr)
			return
		}
		//fin de prueba
		log.Println("El cliente confirmo que envio el archivo")

		

		log.Println("Lado del servidor termino")
			//fin de prueba
}
