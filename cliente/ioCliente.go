package main

import (
	Ps "protocoloSimple"
	"fmt"
	"net"
	"io"
	"os"
	"strings"
	// "time"
)


//Una structura a manera de abstracción para representar un conjunto de funciones que generan salidas
//a la red y reciven entradas. Son procesos intermediario entre el servidor y el cliente que 
//recive una orden y determina la forma de usar el protocolo simple para cumplirla. 
//Los resultados varian dependiendo de la orden recivida.
type IoCliente struct{
	
	Conn net.Conn

	// tiempoLimiteRespuesta time.Time //Tiempo límite para esperar una respuesta
  
	// tiempoLimitePeticion time.Time //Tiempo límite para esperar escribir en la conexion

}

//Inicia las variables
func (ioC *IoCliente) Iniciar(conn net.Conn){
	
	ioC.Conn = conn
	// ioC.Conn.SetDeadline()
}


//Utilisa el protocolo para crear la peticion de unirca a un canal y enviarla. 
func (ioCli *IoCliente) PeticionUnirCanal(nombreCanal string)(exito bool, serr string){
	//estado = "CONECTADO"
	var campos = map[string]string{"nombreCanal":nombreCanal} //campos necesarios para esta petición
	peticion, serr := Ps.NuevaPeticion(ioCli.Conn, "unirCanal", campos)
	if serr != "" {
		return
	}

	serr = peticion.Enviar() 
	if serr != "" {
		if serr == "conndead"{
			return false, "conndead"
		}
		return false, serr
	}

	clienteEsperaRespuesta = true
	tempRsp := <-respuestasServidor
	if tempRsp == ordenRecivida{
		return false, "Atendiendo orden"
	}
	rsp := Ps.Desformatear(tempRsp)
	clienteEsperaRespuesta = false
	if rsp[0] == Ps.S_UNIR_ACEPTADO{
		return true, "" //exito
	}else if rsp[0] == Ps.S_UNIR_RECHAZADO{
		return false, "" //no exito
	}else{
		return false, "La respuesta recivida no es acorde a la peticon:"+ fmt.Sprintf("%s", rsp)
	}
}



//Elabora una peticion para enviar un archivo y la envia, si es aceptada gestiona todo el proceso para que 
//el servidor reciva el archivo
func (ioC *IoCliente) PeticionEnviarArchivo(rutaArchivo, nombreArchivo, pesoArchivo, nombreCanal string)(serror string){

		
	//Debido a los cambios que hice de ultimo minuto ioC.ReplicaRecive es un sustituto para Ps.ReplicaRecive(), esto agrega algo de confución
 	//Primero se notifica al servidor la intención de enviar un archivo
  	peticion := Ps.C_ENVIAR_ARCHIVO + " " + nombreCanal + " " + nombreArchivo + " " + pesoArchivo + "\n" //las "peticiones" llevan \n a diferencia de los "repli", que no llevan.
  	rsp, serr := ioC.ReplicaRecive(peticion)
  	if serr != "" {
  		if serr == "conndead" { //segun el protocolo un error de conexion debe matar al cliente
  			return "conndead" //indica al cliente que cierre todo
  		}
  		return //termina el proceso actual pero sigue conectado
  	}
  	fmt.Println("Respuesta recivida desde el servidor")
  	//evaluo la respuesta, solo puede ser <aprobado> o <reachazado>
  	if rsp[0] == Ps.S_ENVIO_RECHAZADO{	
  		return "Envio rechazdo por el servidor"
  	}

  	if rsp[0] != Ps.S_ENVIO_APROBADO{
  		return "El servidor devolvio una respuesta invalida:"+rsp[0]
  	}
  	//rsp = S_ENVIO_APROBADO, puedo continuar


 	//El servidor está esperando el archivo. Le envio el archivo
 	ar, err := os.Open(rutaArchivo)
 	if err != nil{

 		fmt.Println(err)

 		//el cliente intenta sacar al servidor del bucle de recepción
 		rsp, serr = ioC.ReplicaRecive(Ps.C_TERMINAR_PROCESO)
 		if serr != "" {
 			if serr == "conndead"{ //perdida de conexión
 				return "conndead" //mata a este cliente
 			}
 			//otro tipo de error, por ahora se maneja terminando el programa
 			return "conndead" //para que el caller cierre el programa
 		}
 		//el mensaje se envio y se recivio una respuesta
 		if rsp[0] != Ps.C_TERMINAR_PROCESO{
 			ioC.Conn.Close() //panico
 			return "El servidor devolvio una respuesta invalida:"+rsp[0]
 		}
 		//rsp = C_TERMINAR_PROCESO, puedo continuar
 		return //termina estre proceso, cliente puede hacer nuevas peticiones
 	
 	}
 	defer ar.Close()


 	fmt.Println("Enviando archivo...")
 	//estado = "TRANSMITIENDO_ARCHIVO"
 	_, err = io.Copy(ioC.Conn, ar)
 	if err != nil{ //error durante la transmisión, posible desconexión
 		return "conndead" //termina el programa
 	}else{
 		//despues de terminar la transmisión el cliente se queda esperando el mensaje <archivo recivido> o un error
 		//estado = "NEGOCIANDO_CIERRE"
 		rsp, serr := ioC.ReciveTermina(ioC.Conn)
 		if serr != ""{
 			if serr == "conndead"{ //perdida de conexion
 				return "conndead"
 			}
 			//otro tipo de error, por ahora lo manejo igual
 			return "conndead"
 		}
 		//mensaje recivido

 		if rsp[0] != Ps.S_ARCHIVO_RECIVIDO {
 			
 			if rsp[0] == Ps.S_TERMINAR_PROCESO { //servidor indica que el proceso debe terminar por un error
 				serr := Ps.ReplicaTermina(Ps.S_TERMINAR_PROCESO, ioC.Conn)
 				if serr != ""{
 					if serr == "conndead" { //perdida de conexion
 						return "conndead"
 					}
 					//otro tipo de error
 					return "conndead"
 				}
 				//termina el proceso actual, depues el cliente queda listo para hacer nuevas peticiones
 				return "El servidor notifico un error:" + fmt.Sprintf("%s",rsp[1:])
 			}
 			//la respuesta es distinta de <terminar proceso>, pero no hay más respuestas posibles
 			fmt.Println("Respuesta invalida desde el servidor:", fmt.Sprintf("%s",rsp))
 			return "conndead" //pánico
 		
 		}
 		//rsp[0] = S_ARCHIVO_RECIVIDO, puedo continuar
 		//el servidor espera <archivo enviado>
 		rsp, serr = ioC.ReplicaRecive(Ps.C_ARCHIVO_ENVIADO)
 		if serr != "" {
 			if serr == "conndead"{
 				return "conndead" //mata a este cliente, el servido notara la desconexion tambien
 			}
 			//otro tipo de error
 			return "conndead" 
 		}
 		//mensaje enviado
 		//el lado del cliente no ha terminado todavía, la única forma que tiene el cliente de saber
 		//si el archivo fue procesado con exito es reciviendo el mensaje <fin de transmisión>
 		//puede pasar que el servidor procese el archivo con exito y emita <fin de transmision> pero el
 		//cliente no fmtra recivir el mensaje. En ese caso el servidor continua pero el cliente no está al tanto.
 		if rsp[0] != Ps.S_FIN_TRANSMISION{
 			fmt.Println("Respuesta invalida desde el servidor:", fmt.Sprintf("%s",rsp))
 			return "conndead" //pánico
 		}
 		//NEGOCIANDO CIERRE TERMINADO. Vuelve a estado = "UNIDO_A_CANAL"
 		return "" //por fin todo sali bien. Cliente listo para hacer nuevas peticiones
 	}
}


//Funcion analoga a la funcion del mismo nombre en el lado del servidor, solo que este extremo
//juega el papel de servidor y es quien recive el archivo, para que el otro extremo filtre los 
//mensajes como es debido los mensajes deben tener el prefijo S_ORDEN y terminar con \n
func (ioC *IoCliente) OrdenReciveArchivo(nombreArchivo, pesoArchivo string) (serr string){
	//estado = "NEGOCIANDO_TRANSMISION"
	BUFFER_RCV_ARCHIVO := 1024 //bytes
	rutaLocalArchivo := "desde_canal/"+nombreArchivo
	fmt.Println("ioCliente OrdenReciveArchivo")
	archivo, err := os.Create(rutaLocalArchivo)
	if err != nil{
		serr := Ps.ReplicaTermina(Ps.S_ORDEN+" "+Ps.S_ENVIO_RECHAZADO+"\n", ioC.Conn)
		if serr != ""{
			if serr == "conndead"{ //perdida de conexion
				return "conndead"
			}
			//otro tipo de error
			return "conndead"
		}
		return err.Error()
	}
	defer archivo.Close() 
	//se creo la ruta local para el archivo


	//indico listo para recivir
	serr = Ps.ReplicaTermina(Ps.S_ORDEN+" "+Ps.S_ENVIO_APROBADO+"\n", ioC.Conn)
	if serr != ""{
		if serr == "conndead"{ //perdida de conexion
			return "conndead"
		}
		//otro tipo de error
		return "conndead"
	}
	//estado = "ESPERANDO_ARCHIVO"
	fmt.Println("envio aprobado.")


	fmt.Println("Esperando el archivo...")
	var cuenta int 
	buffer := make([]byte, BUFFER_RCV_ARCHIVO)	
	for {
		n, err := ioC.Conn.Read(buffer)
		cuenta+=n

		//en caso de error al leer de la conexión
		if err != nil{
			archivo.Close() 
			os.Remove(rutaLocalArchivo)
			
			if err == io.EOF { //EOF ocurre si hay una perdida de conexión
				return "conndead"
			}
			return err.Error()
		}


		//en caso de error al escribir en el archivo creado
		_, werr := archivo.Write(buffer[:n])
		if werr != nil{
			archivo.Close()
			os.Remove(rutaLocalArchivo) //eliminar el archivo porque no se pudo crear correctamente
			return werr.Error()
		}


		//Este caso ocurre cuando el cliente termino una transmisión o cuando envia un mensaje para terminar este bucle de recepción de datos
		//puede darce un caso n == 1024 y la transmisión aya terminado?. No, en ese caso la siguiente ejecución de Read() devolvera n = 0
		if n < BUFFER_RCV_ARCHIVO {
			if string(buffer[:n]) == Ps.C_TERMINAR_PROCESO { //el cliente indica que la transmisión debe terminar por un error
				
				archivo.Close()
				os.Remove(rutaLocalArchivo)
				
				serr := Ps.ReplicaTermina(Ps.S_ORDEN+" "+Ps.C_TERMINAR_PROCESO+"\n", ioC.Conn) //el cliente espera este mensaje para continuar
				if serr != ""{
					if serr == "conndead"{ //conexion perdida
						return "conndead" //El cliente se cerrar por completo mas adelante porque la conexion se perdio
					}
					//otro tipo de error, por ahora lo manejo igual. Además el cliente necesita este mensaje o se quedara esperando
					return "conndead"
				}

				//El mensaje <terminar proceso > se envio correctamente
				return "conndead"
			}

			fmt.Println("Trasmision terminada, se presume recepcion completada")
			archivo.Close()
			break
		} //fin del if n < BUFFER_RCV_ARCHIVO
	} //fin del bucle
	//RECEPCION FINALIZADA

	//Verificar la integridad del archivo recivido
	fmt.Println("Bytes esperados:", pesoArchivo)
	fmt.Println("Bytes recividos:", cuenta)
	
	if pesoArchivo != fmt.Sprintf("%d", cuenta) { //comparación de strings
		os.Remove(rutaLocalArchivo)

		msg := Ps.S_ORDEN+" "+Ps.S_TERMINAR_PROCESO + " " + "Los bytes no coinciden " + fmt.Sprintf("%d",cuenta)+"\n"
		rsp, serr := Ps.ReplicaRecive(msg, ioC.Conn)
		if serr != ""{
			if serr == "conndead"{ //conexion perdida
				return "conndead"
			}
			//otro tipo de error, por ahora lo manejo igual
			return "conndead"
		}
		//respuesta recivido
		if rsp[0] != Ps.S_TERMINAR_PROCESO {
			//S_TERMINAR_PROCESO es la única respuesta esperada
			//panico
			return "El cliente envio una respuesta invalida:"+fmt.Sprintf("%s",rsp)
		}
		//la respuesta recivida es S_TERMINAR_PROCESO, puedo continuar
		return "Los bytes recividos no son iguales a los bytes enviados por el cliente"
	}
	//los bytes coinciden, el cliente esta espera <archivo recivido> para continuar

	msg := Ps.S_ORDEN+" "+Ps.S_ARCHIVO_RECIVIDO + " " + "Archivo recivido, bytes totales: " + fmt.Sprintf("%d",cuenta) +"\n"
	tempRsp, serr := Ps.ReplicaRecive(msg, ioC.Conn)
	if serr != "" {
		if serr == "conndead" { //perdida de conexion
			return "conndead"//El servidor dejara de escuchar a este cliente más adelante
		}
		//otro tipo de error, por ahroa manejo el error terminando la conexion
		return "conndead"
	}
	rsp := strings.Trim(tempRsp[0],"\n") //esta respuesta viene con un \n al final
	//recive la respuesta
	if rsp != Ps.C_ARCHIVO_ENVIADO {
		//pánico
		return "El cliente envio una respuesta invalida:"+ fmt.Sprintf("%s",rsp)
	}
	//rsp = C_ARCHIVO_ENVIADO, puedo continuar
	//el cliente espera <fin de transmision>, si este mensaje no se puede enviar el servidor continua igual, solo el cliente no está al tanto
	serr = Ps.ReplicaTermina(Ps.S_ORDEN+" "+Ps.S_FIN_TRANSMISION+"\n", ioC.Conn)
	if serr != ""{
		if serr == "conndead"{ //perdida de conexion
			//el archivo se sigue procesando, pero el cliente se elimina por la perdida de conexion
			return "conndead"
		}
		//otro tipo de error, por ahora lo cierro completamente
		return "No se pudo replicar <fin de transmision> pero el servidor continua"
	}
	return "" 
}




func MuestraRegistro(){

	fmt.Printf("%s",Ps.LOG) //imprime el registro de entradas del <protocolo simple> liena a linea

}

//Funcion analoga a la del protocolo, solo que permite operar con el cambio de haber agregado
//la rutina escucharOrdenesRespuestas()
func (ioC *IoCliente) ReplicaRecive(psmsj string)(rsp []string, error string){
	bstream := []byte(psmsj)
	_ , err := ioC.Conn.Write(bstream)
	if err != nil{
		if err == io.EOF{
			//la conexion se perdio //simplemente indica un error
			return rsp, "conndead"
		}
		//otro tipo de error
		return rsp, err.Error()
	}
	//el mensaje se envio

	clienteEsperaRespuesta = true
	temp := <-respuestasServidor
	clienteEsperaRespuesta = false
	rsp = strings.Split(temp, " ")
	return rsp, ""
}



func (ioC *IoCliente) ReciveTermina(conn net.Conn)(rsp []string, serr string){

	clienteEsperaRespuesta = true
	temp := <-respuestasServidor
	clienteEsperaRespuesta = false
	rsp = strings.Split(temp, " ")
	return rsp, ""

}