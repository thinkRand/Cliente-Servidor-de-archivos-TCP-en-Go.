package protocolo_simple

import (
	"net"
	"fmt"
	"bufio"
)


const(
	//RESPUESTAS DEL SERVIDOR
	SERVIDOR_UNIR_APROBADO = "canalaprobado"
	SERVIDOR_UNIR_NOAPROBADO = "canalanoprobado"
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
	CLIENTE_ERROR_TERMINAR_ENVIO = "terminar"
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
		return CLIENTE_UNIR_CANAL +" "+ parametros["canal"], ""

	
	case "enviar":
		//El formato de este mensaje es: código de enviar <espacio> nombre-canal <espacio> nombre-archivo <espacio> peso-archivo
		return CLIENTE_ENVIAR_ARCHIVO +" "+ parametros["nombreCanal"] +" "+ parametros["nombreArchivo"] +" "+ parametros["pesoArchivo"], ""
	
	case "desconectar":
	

	case "salir-canal":
	

	default:
		return "","Comando desconocido."
	}
	return "", "Comando no especificado" 
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