package main

import(
	Ps "protocoloSimple"
	"io"
	"os"
	"net"
	"bufio"
	"log"
	"fmt"
	"strings"
)

//Cliente representa un solo cliente conectado al servidor
type Cliente struct{
	
	conn net.Conn
	
	escribir chan string //Para escribir mensajes al cliente

	recivir chan string //Para recivir mensaje del cliente, se usa cuando el lado del servidor juega el papel de cliente y se comunica con el cliente (que juega el papel de servidor)

	archivo chan *archivo //Para recivir los datos de un archivo que se debe enviar a un cliente

	listo chan bool //Transmite un valor boleano cada ves que termina una operacion, este valor es recojido por otras subrutinas para conocer que el proceso se termino. Es para coordinación.

	canales *Canales //Lista de canales que existen

	canal *Canal //El canal done este cliente está registrado 
}


//lee los mensajes provenientes del cliente y los manda al interprete de códigos
//esta rutina termina si el la conexion del cliente se desase
func (cli *Cliente) Leer(){
	lector := bufio.NewScanner(cli.conn)
	for lector.Scan(){
		entrada := lector.Text()
		log.Println("Recivido::", entrada)
		Interpretar(cli, entrada) //se interpreta uno a la vez, es sincrono, a diferencia de escribir que es asincrono
		// log.Println("Escuchando nuevas peticiones...")
	}
}


//Toma todos los mensaje para este cliente y se los envía
//esta rutina termina si el canal para escribir se cierra
func (cli *Cliente) Escribir(){

	for msg := range cli.escribir{
		n, err := fmt.Fprintln(cli.conn, msg) //se ignoran los errores
		if err != nil{
			if n == 0 && err == io.EOF{
				cli.SalirCanal()
				return
			} 
			// log.Println(err)
			cli.listo<- true //terminado
			continue //hay que seguir pese a los errores
		}
		log.Println("Despachado::", msg)
		cli.listo<- true //terminado
	}

}


//Elimina los datos de este cliente que tengan que ver con un canal.
//Lo saca del canal y elimina su referencia a el.
func (cli *Cliente) SalirCanal(){
	if cli.canal != nil{
		referenciaCanal := cli.canal
		referenciaCanal.salir<-cli //El canal elimina al cliente de su registro y tambien elimina la referencia del canal en el cliente
		<-referenciaCanal.listo
		cli.canal = nil
	}
}

//Termina por completo la sesión con este cliente.
func (cli *Cliente) Close(){

	cli.SalirCanal()
	cli.conn.Close()
	//debería cerrar los canales aquí?
}


//Recive los archivos que se debe enviar a este cliente
func (cli *Cliente) EnviarArchivos(){
	for archivo := range cli.archivo {
		cli.enviarArchivo(archivo)
	}
}


//Efectua la coordinación con un cliente para que reciva un archivo, es análoga a la funcion 
//del cliente con las mismas caracteristicas.
//devuelve true si el proceso termina con exito, false de otro modo.
//Este extremo juega el papel de cliente en este caso, mientras que el otro extremo juega el papel de servidor.
func (cli *Cliente) enviarArchivo(archivo *archivo) bool{
		//Primero se notifica la intención de enviar un archivo
		nombreCanal := cli.canal.nombre
		nombreArchivo := archivo.nombre
		pesoArchivo := archivo.peso
		rutaArchivo := archivo.ruta
	 	peticion := Ps.S_ORDEN +" "+ Ps.C_ENVIAR_ARCHIVO + " " + nombreCanal + " " + nombreArchivo + " " + pesoArchivo
	 	cli.escribir<- peticion
	 	<-cli.listo
	 	tempRsp := <-cli.recivir	
	 	rsp := strings.Split(tempRsp, " ")
	 	// rsp, serr := Ps.ReplicaRecive(peticion, cli.conn)
	 	// if serr != "" {
	 	// 	if serr == "conndead" { //segun el protocolo un error de conexion debe matar al cliente
	 	// 		return false 
	 	// 	}
	 	// 	return false 
	 	// }
	 	// log.Println("Respuesta recivida")
	 	//evaluo la respuesta, solo puede ser <aprobado> o <reachazado>
	 	if rsp[1] == Ps.S_ENVIO_RECHAZADO{	
	 		log.Println("Envio rechazdo por el servidor")
	 		return false
	 	}

	 	if rsp[1] != Ps.S_ENVIO_APROBADO{
	 		log.Println("El servidor devolvio una respuesta invalida:"+rsp[0])
	 		return false
	 	}
	 	//rsp = S_ENVIO_APROBADO, puedo continuar
		log.Println("El servidor acepto la transferencia del archivo")	



		//El servidor está esperando el archivo. Le envio el archivo
		ar, err := os.Open(rutaArchivo)
		if err != nil{

			log.Println(err)

			//el cliente intenta sacar al servidor del bucle de recepción
			rsp, serr := Ps.ReplicaRecive(Ps.C_TERMINAR_PROCESO, cli.conn)
			if serr != "" {
				if serr == "conndead"{ //perdida de conexión
					return false
				}
				//otro tipo de error, por ahora se maneja terminando el programa
				return false 
			}
			//el mensaje se envio y se recivio una respuesta
			if rsp[0] != Ps.C_TERMINAR_PROCESO{
				cli.Close() //panico
				log.Println("El servidor devolvio una respuesta invalida:"+rsp[0])
				return false
			}
			//rsp = C_TERMINAR_PROCESO, puedo continuar
			return false
		
		}
		defer ar.Close()


		log.Println("Enviando archivo...")
		//estado = "TRANSMITIENDO_ARCHIVO"
		_, err = io.Copy(cli.conn, ar)
		if err != nil{ //error durante la transmisión, posible desconexión
			return false
		}else{
			//despues de terminar la transmisión el cliente se queda esperando el mensaje <archivo recivido> o un error
			//estado = "NEGOCIANDO_CIERRE"
			tempRsp := <-cli.recivir
			rsp := strings.Split(tempRsp, " ")
			// rsp, serr := Ps.ReciveTermina(cli.conn)
			// if serr != ""{
			// 	if serr == "conndead"{ //perdida de conexion
			// 		return false
			// 	}
			// 	//otro tipo de error, por ahora lo manejo igual
			// 	return false
			// }
			//mensaje recivido
			// log.Println("En gestionando orden, recivido:", tempRsp)
			if rsp[1] != Ps.S_ARCHIVO_RECIVIDO {
				
				if rsp[1] == Ps.S_TERMINAR_PROCESO { //servidor indica que el proceso debe terminar por un error
					serr := Ps.ReplicaTermina(Ps.S_TERMINAR_PROCESO, cli.conn)
					if serr != ""{
						if serr == "conndead" { //perdida de conexion
							return false
						}
						//otro tipo de error
						return false
					}
					//termina el proceso actual, depues el cliente queda listo para hacer nuevas peticiones
					log.Println("El servidor notifico un error:" + fmt.Sprintf("%s",rsp[1:]))
					return false
				}
				//la respuesta es distinta de <terminar proceso>, pero no hay más respuestas posibles
				log.Println("Respuesta invalida desde el servidor:", fmt.Sprintf("%s",rsp))
				cli.Close()
				return false //pánico
			
			}
			//rsp[0] = S_ARCHIVO_RECIVIDO, puedo continuar
			//el servidor espera <archivo enviado>
			cli.escribir <- Ps.C_ARCHIVO_ENVIADO
			<-cli.listo
			tempRsp = <-cli.recivir
			rsp = strings.Split(tempRsp, " ")
			// rsp, serr := Ps.ReplicaRecive(Ps.C_ARCHIVO_ENVIADO, cli.conn)
			// if serr != "" {
			// 	cli.SalirCanal()
			// 	if serr == "conndead"{
			// 		return false 
			// 	}
			// 	//otro tipo de error
			// 	return false 
			// }
			//mensaje enviado
			//el lado del cliente no ha terminado todavía, la única forma que tiene el cliente de saber
			//si el archivo fue procesado con exito es reciviendo el mensaje <fin de transmisión>
			//puede pasar que el servidor procese el archivo con exito y emita <fin de transmision> pero el
			//cliente no logra recivir el mensaje. En ese caso el servidor continua pero el cliente no está al tanto.
			if rsp[1] != Ps.S_FIN_TRANSMISION{
				log.Println("Respuesta invalida desde el servidor:", fmt.Sprintf("%s",rsp))
				cli.Close()
				return false //pánico
			}
			log.Println("Este lado termino...")
			return true //por fin todo salio bien. Cliente listo para hacer nuevas peticiones
		}
}