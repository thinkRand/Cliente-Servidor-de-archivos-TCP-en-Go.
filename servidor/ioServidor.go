package main

/*
	La intención de este modulo es servir de intermediario entre el servidor y el cliente desde el lado
	del servidor. Solo debería tener lógica para enviar y recivir, sin embargo todavía hay otras partes del 
	programa que estan haciendo io con la conexión, eso debe cambiar.
*/

import(
	Ps "protocoloSimple"
	"strings"
	"log"
	"fmt"
	"os"
	"io"
)

const (
	BUFFER_RCV_ARCHIVO = 1024 //Capacidad del buffer para recivir datos de archivos desde la red, en bytes
)

//Recive los mensajes del cliente y reconocer los códigos para ofrecer el procedimiento adecuado,
//tambien hace labores de validación de formato. 
//nota: Los formatos se deberían validar usan el paquete del protocolo, alg como verificarFormato(msj)
func Interpretar(cli *Cliente, entrada string){
	campos := strings.Split(entrada, " ")
	
	switch(campos[0]){ //El código viene de primero
	
	case Ps.C_UNIR_CANAL:
		//Se esperar C_UNIR_CANAL <espacio> nombreCanal
		if len(campos) != 2{
			cli.escribir<- Ps.S_UNIR_RECHAZADO 
			<-cli.listo //espero que cli.escribir se termine de procesar
			return
		}
		canal := campos[1]
		PeticionUnirCanal(cli, canal)
	

	case Ps.C_ENVIAR_ARCHIVO:
		//se espera el formato: enviar <espacio> canal <espacio> nombre-archivo <espacion> peso-archivo\n
		if len(campos) != 4{
			cli.escribir<- Ps.S_ENVIO_RECHAZADO
			<-cli.listo //espero que se termine de procesar
			return 
		} 
		canal := campos[1]
		nombreArchivo := campos[2]
		pesoArchivo := campos[3]
		PeticionEnviarArchivo(cli, canal, nombreArchivo, pesoArchivo)
	

	case Ps.S_ORDEN:
		cli.recivir<- entrada

	default:
		log.Println("Peticion invalida, se desconectara al cliente")
		//pánico
		cli.Close()

	}
}


//Evalua si el cliente se puede unir a el canal que solicita y lo agrega si es posible.
//nota: debería implementar metodos del paquete protocoloSimple
func PeticionUnirCanal(cli *Cliente, canal string){

	if cli.canal != nil{
		log.Println("El cliente ya esta en un canal")
		cli.escribir<- Ps.S_UNIR_RECHAZADO  
		<-cli.listo //espero que cli.escribir se termine de procesar
		return
	}
	
	if _, ok := cli.canales.lista[canal]; !ok {	//true si el canal no existe
		log.Println("El cliente pidio unirce a un canal que no existe")
		cli.escribir<- Ps.S_UNIR_RECHAZADO  
		<-cli.listo //espero que cli.escribir se termine de procesar
		return
	}

	canalObjetivo := cli.canales.lista[canal]
	canalObjetivo.unir<- cli //se registra al cliente en el canal
	<-canalObjetivo.listo //verifico si el canal termino de registrarlo
	
	cli.canal = &canalObjetivo //el cliente conoce al canal ahora 
	cli.escribir<- Ps.S_UNIR_ACEPTADO 	//se responde al cliente el mensaje adecuado
	<-cli.listo //espero que se termine de procesar 

}

//Recive la peticion para <enviar archivo> desde el cliente y la notifíca al canal. 
func PeticionEnviarArchivo(cli *Cliente, canal, nombreArchivo, pesoArchivo string){


		if cli.canal == nil{
			log.Println("El cliente no esta en un canal")
			cli.escribir<- Ps.S_ENVIO_RECHAZADO
			<-cli.listo 
			return
		}

		if cli.canal.nombre != canal{
			log.Println("El cliente no esta en ese canal")
			cli.escribir<- Ps.S_ENVIO_RECHAZADO
			<-cli.listo //espero que se termine de procesar 
			return
		}

		//se le pregunta al canal si puede gestinar la peticion de este cliente
		respuesta := cli.canal.GestionaRecivir(cli, nombreArchivo, pesoArchivo)
		if respuesta == "OCUPADO"{ 
			log.Println("El canal está ocupado")
			cli.escribir<- Ps.S_ENVIO_RECHAZADO
			<-cli.listo //espero que se termine de procesar 
		}
		//lo que suceda con la peticion es asunto del canal a partir de ahora 
}





//Funcion de prueba que maneja la subida de un archivo desde el cliente al servidor
//y la guardar en la variable destinoArchivo
func ReciveArchivo(destinoArchivo *archivo , cli *Cliente, nombreArchivo, pesoArchivo string) bool{
	
	archivo, err := os.Create("servidor_archivos/"+nombreArchivo)
	if err != nil{
		log.Println(err)
		serr := Ps.ReplicaTermina(Ps.S_ENVIO_RECHAZADO+"\n", cli.conn)
		if serr != ""{
			if serr == "conndead"{ //perdida de conexion
				cli.SalirCanal()
				return false
			}
			//otro tipo de error
			cli.Close()
			return false
		}
		return false
	}
	defer archivo.Close() 
	

	serr := Ps.ReplicaTermina(Ps.S_ENVIO_APROBADO+"\n", cli.conn)
	if serr != ""{
		if serr == "conndead"{ //perdida de conexion
			cli.SalirCanal()
			return false
		}
		//otro tipo de error
		cli.Close()
		return false
	}
	//estado = "ESPERANDO_ARCHIVO"
	


	log.Println("Esperando el archivo...")
	var cuenta int 
	buffer := make([]byte, BUFFER_RCV_ARCHIVO)	
	for {
		n, err := cli.conn.Read(buffer)
		cuenta+=n

		//en caso de error al leer de la conexión
		if err != nil{
			log.Println(err)
			archivo.Close() 
			os.Remove("servidor_archivos/"+nombreArchivo)
			
			if err == io.EOF { //EOF ocurre si hay una perdida de conexión
				//cierre forsoso, borrado de datos
				cli.SalirCanal()
				return false
			}
			//Otro error. Por ahora manejo el error de la siguiente forma
			cli.Close()
			return false
		}


		//en caso de error al escribir en el archivo creado
		_, werr := archivo.Write(buffer[:n])
		if werr != nil{
			log.Println(werr)
			archivo.Close()
			os.Remove("servidor_archivos/"+nombreArchivo) //eliminar el archivo porque no se pudo crear correctamente
			//puedo continuar pero el cliente no escuchara un mensaje de <terminar proceso> hasta que
			//termine de transmitir todo el archivo. 

			//Por eso tengo que descartar todo lo que se reciva hasta que el cliente replique un <terminar proceso>
			//por ahora matare la conexion y cerrare todo, luego hay que manejar el error como debe ser.
			cli.Close()
			return false
		}


		//Este caso ocurre cuando el cliente termino una transmisión o cuando envia un mensaje para terminar este bucle de recepción de datos
		//puede darce un caso n == 1024 y la transmisión aya terminado?. No, en ese caso la siguiente ejecución de Read() devolvera n = 0
		if n < BUFFER_RCV_ARCHIVO {
			if string(buffer[:n]) == Ps.C_TERMINAR_PROCESO { //el cliente indica que la transmisión debe terminar por un error
				
				archivo.Close()
				os.Remove("servidor_archivos/"+nombreArchivo)
				
				serr := Ps.ReplicaTermina(Ps.C_TERMINAR_PROCESO+"\n", cli.conn) //el cliente espera este mensaje para continuar
				if serr != ""{
					if serr == "conndead"{ //conexion perdida
						cli.SalirCanal()
						return false //El cliente se cerrar por completo mas adelante porque la conexion se perdio
					}
					//otro tipo de error, por ahora lo manejo igual. Además el cliente necesita este mensaje o se quedara esperando
					cli.Close()
					return false
				}

				//El mensaje <terminar proceso > se envio correctamente
				return false//continua a escuchar nuevas peticiones
			}

		log.Println("Trasmision terminada, se presume recepcion completada")
		archivo.Close()
		break
		} //fin del if n < BUFFER_RCV_ARCHIVO
	} //fin del bucle
	//RECEPCION FINALIZADA

	//Verificar la integridad del archivo recivido
	log.Println("Bytes esperados:", pesoArchivo)
	log.Println("Bytes recividos:", cuenta)
	
	if pesoArchivo != fmt.Sprintf("%d", cuenta) { //comparación de strings. 
		log.Println("Los bytes recividos no son iguales a los bytes enviados por el cliente")
		os.Remove("servidor_archivos/"+pesoArchivo)

		msg := Ps.S_TERMINAR_PROCESO + " " + "Los bytes no coinciden " + fmt.Sprintf("%d",cuenta)+"\n"
		rsp, serr := Ps.ReplicaRecive(msg, cli.conn)
		if serr != ""{
			if serr == "conndead"{ //conexion perdida
				cli.SalirCanal()
				return false
			}
			//otro tipo de error, por ahora lo manejo igual
			cli.Close()
			return false
		}
		//respuesta recivido
		if rsp[0] != Ps.S_TERMINAR_PROCESO {
			//S_TERMINAR_PROCESO es la única respuesta esperada
			log.Println("Panico. El cliente envio una respuesta invalida:", fmt.Sprintf("%s",rsp))
			//panico
			cli.Close()
			return false
		}
		//la respuesta recivida es S_TERMINAR_PROCESO, puedo continuar
		return false//despues el servidor queda lista para recivir nuevas peticiones
	}
	//los bytes coinciden, el cliente esta espera <archivo recivido> para continuar


	msg := Ps.S_ARCHIVO_RECIVIDO + " " + "Archivo recivido, bytes totales: " + fmt.Sprintf("%d",cuenta)+"\n"
	rsp, serr := Ps.ReplicaRecive(msg, cli.conn)
	if serr != "" {
		log.Println("Error.")
		if serr == "conndead" { //perdida de conexion
			cli.SalirCanal()
			return false//El servidor dejara de escuchar a este cliente más adelante
		}
		//otro tipo de error, por ahroa manejo el error terminando la conexion
		cli.Close()
		return false
	}
	//recive la respuesta
	if rsp[0] != Ps.C_ARCHIVO_ENVIADO{
		//pánico
		log.Println("Panico. El cliente envio una respuesta invalida:", fmt.Sprintf("%s",rsp))
		cli.Close()
		return false
	}
	//rsp = C_ARCHIVO_ENVIADO, puedo continuar
	//el cliente espera <fin de transmision>, si este mensaje no se puede enviar el servidor continua igual, solo el cliente no está al tanto
	serr = Ps.ReplicaTermina(Ps.S_FIN_TRANSMISION+"\n", cli.conn)
	if serr != ""{
		if serr == "conndead"{ //perdida de conexion
			//el archivo se sigue procesando, pero el cliente se elimina por la perdida de conexion
			cli.SalirCanal()
			return false
		}
		//otro tipo de error, por ahora lo cierro completamente
		log.Println("No se pudo replicar <fin de transmision> pero el servidor continua")
	}

	destinoArchivo.ruta = "servidor_archivos/"+nombreArchivo
	destinoArchivo.nombre = nombreArchivo
	destinoArchivo.peso = pesoArchivo
	log.Println("Lado del servidor termino")
	return true 
}



//Se encargar de avisar a cada cliente que le asignen que debe recivir un archivo
//específico. No sabe si esos cliente reciven el archivos, eso debe cambiar.
func Redirige(grupo map[*Cliente]bool, ar *archivo) bool{
	
	for cliente := range grupo {
		cliente.archivo <- ar
	}
	return true
	//podría ser algo como 
	//go enviarArchivo(cliente, archivo) para enfatisar que esta parte organiza la gestion y no dejarle eso al cliente
}

