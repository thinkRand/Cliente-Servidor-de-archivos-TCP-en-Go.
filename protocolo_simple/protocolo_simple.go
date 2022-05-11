package protocolo_simple

import (
	"net"
	"fmt"
	"bufio"
	"log"
)


const(
	//RESPUESTAS DEL SERVIDOR
	S_UNIR_ACEPTADO = "canalaprobado"
	S_UNIR_RECHAZADO = "canalanoprobado"
	S_SALIR_APROBADO = "saliraprobado"
	S_CONEXION_APROBADO = "conexionaprobada"
	S_ENVIO_APROBADO = "envioaprobado"
	S_ENVIO_RECHAZADO = "envionoaprobado"
	S_ERROR_CMD = "El comando es invalido"
	S_MSG = "msg" //para crear mensajes estandar sin relevancia para la coordinación, su destion es la pantalla del cliente


	//PETICIONES DEL CLIENTE
	C_UNIR_CANAL = "unir"
	C_SALIR_CANAL = "salir"
	C_CONEXION = "establecerconexion"
	C_ENVIAR_ARCHIVO = "enviararchivo"
	C_ERROR_TERMINAR_ENVIO = "terminar"
)




//El traductor toma una forma de comunicacion y la transforman a otra forma de comunicacion.
//En este sentido esta función recive una cadena de texto y transforma esa información al formato del 
//protocolo simple
//los valores posibles son:
//comando eviar : nombre-archivo, peso-archivo, nombre-canal
//comando unir : nombre-canal
func Traducir(cmd string, parametros map[string]string)(mensaje string, error string){
	switch(cmd){

	case "unir":
		//El formato de este mensaje es: código de unir <espacio> nombre-canal 
		if _, ok := parametros["canal"]; !ok{
			return "", "El indice -canal- no existe"
		}
		return C_UNIR_CANAL +" "+ parametros["canal"], ""

	
	case "enviar":
		//El formato de este mensaje es: código de enviar <espacio> nombre-canal <espacio> nombre-archivo <espacio> peso-archivo
		return C_ENVIAR_ARCHIVO +" "+ parametros["nombreCanal"] +" "+ parametros["nombreArchivo"] +" "+ parametros["pesoArchivo"], ""
	
	case "desconectar":
	

	case "salir-canal":
	

	default:
		return "","Comando desconocido."
	}
	return "", "Comando no especificado" 
}

//Es el proceso inverso a Traducir.
//Recive un mensaje con formato de protocolo simple y lo traduce a un mensaje mas manejable
func Interpretar(psmsj string)(msjComun string){
	msjComun = psmsj
	return msjComun
}

//Recive una peticion en formato de protocolo simple y la envia al servidor
//segun el protocolo simple la comunicacion se hace con strings separados por espacios
//retorna el string exito el la primera variable si todo sale bien
//si hay un error retorna la descripción del error en la segunda variable
func HacerPeticion(peticion string, conn net.Conn)(ok string, error string){
	//validar la peticon
	_ , err := fmt.Fprintf(conn, peticion)
	if err != nil {
		return "", "No se pudo enviar la petición"
	}
	return "existo", ""
}

//recive una respuesta desde el servidor en formato de protocolo simple para el tipo de peticion
//a la que responde
//las respuestas se reciven como texto
func RecivirRespuesta(conn net.Conn)(rsp string, error string){
	lector := bufio.NewScanner(conn)
	if lector.Scan(){
		respuesta := lector.Text() //lee hasta \n
		return respuesta, ""
	}
	return	"", "no se pudo leer la respuesta del servidor"
}



//Empaquetar toma un mensaje con formato de protocolo siple y lo comprime.
//Por ahora hace nada.
func Empaquetar(msj string)(mensaje string){
	return msj
}


//Desenpaqueta una respuesta del servidor para que sea legible mas adelante.
//el resultado es una respuesta con formato del protocolo simple.
//Por ahora hace nada.
func Desempaquetar(msj string)(mensaje string){
	return msj
}



type Peticion struct{

	conn net.Conn //lo conexion por donde se envía la petición y se espera la respuesta

	orden string

	parametros map[string]string
}

//Envía la petición y retorna éxito o fracaso segun sea el caso
//la implementacion de la peticion sabe que tipo de peticion se envia y sabe que respuesta esperar
func (p *Peticion) Enviar()(rsp string){
	lector := bufio.NewScanner(p.conn)
	//if p.orden es unir entonces se espera unir aprobado
	//if p.orden es sali-canal entonces se espera salir aprobado o rechazado
	//...
	
	traduccion, serr := Traducir(p.orden, p.parametros)
	if serr != ""{
		log.Println("DEBUG:", serr)
		return "fracaso"
	}

	paquete := Empaquetar(traduccion) //hace nada, es para demostrar un comportamiento esperado mas adelante
	_, err := fmt.Fprintf(p.conn, paquete) //con el protocolo simple envio y recivo strings
	if err != nil{
		log.Println("DEBUG:", err)
		return "fracaso"
	}

	if lector.Scan(){
		rsp := lector.Text()
		rsp = Desempaquetar(rsp)
		rsp = Interpretar(rsp)
		if rsp != ""{ //esto depende de la peticon echa
			return "exito"	
		}
		return "fracaso"
	}else{
		//end of input o error
		err := lector.Err()
		if err != nil{
			log.Println("DEBUG:", err)
			return "fracaso"
		}else{
			//io.EOF, desconecion
			log.Println("DEBUG: servidor desconectado")
		}
	}

	return "fracaso"
}


func NuevaPeticion(orden string, parametros map[string]string, conn net.Conn)(p Peticion, error string){
	p = Peticion{
		orden:orden,
		parametros:parametros,
		conn:conn,
	}
	return p, ""
}




/*
	Siento que la implementación de enviar petición esta haciendo demaciado trabajo
	> traducir > empaquetar > enviar >------RED------< recivir > desempacar > interpretar

	Siento que la gestión de todo ese proceso debería ser parte de un proceso de nivel superio

	pero debe ser parte del protocolo y no de la implementación del cliente.
	no debe ser parte del intermediario pero debe responder a él.

	Que nombre deberia darle?¡
	gestor de peticion
	manejador de conexion
	true false getter








*/

