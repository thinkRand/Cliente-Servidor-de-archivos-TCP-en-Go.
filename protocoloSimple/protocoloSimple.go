/*La intencion de este paque es proporcionar las herramientas para usar el protocolo simple, como control de
sintaxsis, uso adecuado de códigos, manejo de errores en la coordinación etc*/
package protocoloSimple

import (
	"net"
	"strings"
	"io"
)

/*
	Este protocolo no tolera perdidas de conexión aunque esto es muy común en sistemas cliente-servidor.
*/

var(
	//reglas que describen el formato exacto que debe tener una petición. 
	//con len(regla) se conoce el número exacto de campos que debe tener una petición. 
	//con regla[k] = value se sabe que k es la posicion exacta que debe tener un valor. 
	//con relag[] = value se conocen todos los campos que se requieren.
	//intente que fueran arreglos para enfatisar su caracter estricto, pero los arreglos complicaron las cosas 
	reglaDesconectar []string //no requiere campos
	
	reglaEnviarArchivo = []string{"nombreCanal", "nombreArchivo", "pesoArchivo"} //solo 3 campos obligatorios, en el orden descrito
	
	LOG = make([]string, 0, 1000) //Guarda los mensajes que vienen y van entre el cliente y el servior hasta 1000, para probar
)


const(
	//CODIGOS DE RESPUESTAS DEL SERVIDOR 
	
	// S_DESCONECTAR_ACEPTADO = "sda"
	S_UNIR_ACEPTADO = "sua"
	S_UNIR_RECHAZADO = "sur"
	S_ENVIO_APROBADO = "sea"
	S_ENVIO_RECHAZADO = "ser"
	S_ARCHIVO_RECIVIDO = "sar" //El servidor indica que recivio el archivo, normalmente se requiere una respuesta del cliente indicando que si envio el archivo.
	S_ORDEN = "so" //Indica una orden para el cliente
	S_SALIR_ACEPTADO = "ssa"
	S_SALIR_RECHAZADO = "ssr"
	// S_CODIGO_INVALIDO = "sci"
	//CÓDIGOS DE AVISOS DEL SERVIDOR
	
	// S_IMPOSIBLE_CONTINUAR = "sic" //Mata la conexion desde el servidor. Si el cliente la recive debe dejar de funcionar. Si el cliente no recive la respuesta igual debe finalisar con un deadline
	S_TERMINAR_PROCESO = "stp" //El servidor no puede continuar con el proceso actual, notifíca al cliente y espera una respuesta identica para poder continuar.
	S_FIN_TRANSMISION = "sft" //El servidor indica al cliente que este proceso termino con exito, no espera reply. Si este mensaje no llega al cliente igual el servidor coninua procesando, solo el cliente no esta al tanto.

	//CÓDIGOS DE AVISOS DEL CLIENTE
	
	// C_IMPOSIBLE_CONTINUAR = "cic"
	C_TERMINAR_PROCESO = "ctp" //El cliente no puede continuar el proceso actual, notifíca al servidor y espera un respuesta identica para poder continuar. Si el cliente no recive respuesta o no puede enviar el mensaje debe haber kill de la conexion
	
	//CÓDIGOS DE PETICIONES DEL CLIENTE
	
	C_UNIR_CANAL = "cuc"
	C_SALIR_CANAL = "csc"
	// C_DESCONECTAR = "cd"
	C_ENVIAR_ARCHIVO = "cea" 
	C_ARCHIVO_ENVIADO = "cae" //El cliente certifica que envio el archivo al servidor.

	bufferMsg = 128 //Tamaño maximo de mensajes en bytes. 128 bytes es más que suficiente por ahora
)


func registrarMsj(msj string){

	total := len(LOG) //cuantos registros tengo en el historial
	if total == cap(LOG){
		return //ya no guardes mas, esta prueba ya duro mucho xd
	}
	LOG = LOG[:total+1] //agrego un espacio más
	LOG[total] = msj //agrega el mensaje a la última posición

}


//Una coleccion de informacion relevante para una peticion asi como sus metodos propios
type peticion struct{
	
	conn net.Conn

	codigo string //Código que representa a este mensaje según el protocolo simple 

	orden string //La forma del comando en lenguaje comun, UNIR, DESCONECTAR ETC...

	campos map[string]string //lista de campos de una peticion y sus valores

	regla []string //describe el orden exacto que deben tener los campos, y los campos exactos que debe tener la petición. intente que fuera un arreglo pero no pude

	msjFormateado string //la cadena con formato de protocolo simple (strings separados por espacios)

	respuestas map[string]bool //Lista de las repuestas posibles para esta petición

}


var peticionUnirCanal = peticion{
		// conn: 
		codigo: C_UNIR_CANAL, 
		orden: "unirCanal",
		campos: make(map[string]string),
		regla: []string{"nombreCanal"}, //un solo campo obligatorio (nombreCanal) en la posición 0
		msjFormateado: "",
		respuestas:  map[string]bool{S_UNIR_ACEPTADO:true, S_UNIR_RECHAZADO:true},
	}


var peticionSalirCanal = peticion{
	// conn: 
	campos: make(map[string]string),
	orden: "salirCanal",
	codigo: C_SALIR_CANAL, 
	regla: []string{"nombreCanal"}, //un solo campo obligatorio (nombreCanal) en la posición 0
	respuestas:  map[string]bool{S_SALIR_ACEPTADO:true, S_SALIR_RECHAZADO:true},
	msjFormateado: "",
}


var listaPeticiones = make(map[string]peticion) //guarda el registro de todas las peticiones disponibles 


//Crea una peticion estructurada de una forma única para cada tipo de peticion.
//El parametro orden es el tipo de peticion, mientros los campos son la carga del mensaje
func NuevaPeticion(conn net.Conn, orden string, campos map[string]string)(p peticion, error string){

	listaPeticiones["unirCanal"] = peticionUnirCanal
	listaPeticiones["salirCanal"] = peticionSalirCanal

	if pet, ok := listaPeticiones[orden]; !ok{
		return p,"La orden no existe"
	}else{
		//las peticiones vienen preformateadas y solo necesitan la conexion y los campos
		pet.conn = conn 
		pet.campos = campos
		return pet, "" //sin errores

	}
}


//Ejecuta todo el proceso para enviar esta peticon por la conexion. Los procesos son
//formatear la petición > marshall > enviar. Retorna la descripcion de un error si llega a suceder.
func (p *peticion) Enviar()(serr string){
	
	p.formatear()
	bstream := p.marshal()
	_ , err := p.conn.Write(bstream)
	if err != nil{
		if err == io.EOF{
			return "conndead"
		}
		return err.Error()
	}		
	registrarMsj(p.msjFormateado) //guardar la cadena enviada en el historial

	return "" //no hay error
}


//Recive la respuesta a una peticion ejecutando los procesos de 
//recivir > desformatear > unmarshall > return. Retorna una respuesta o un error si este sucede.
func (p *peticion) RecivirRespuesta()(rsp []string, serr string){

	buff := make([]byte, bufferMsg)	
	
	n, err := p.conn.Read(buff[:])
	if err != nil{
		if err == io.EOF{
			return rsp, "conndead"
		}
		return rsp, err.Error() //vacio y un error 
	}

	umsj := unmarshal(buff[:n]) //de bytes a cadena de texto con formato del protocolo
	registrarMsj(umsj) //actualisa el historial de mensajes 
	r, serr := p.desformatear(umsj) //de cadena de texto a []strings
	if serr != "" {
		return rsp, serr //vacio y un error
	}

	return r, "" //todo bien
}


//Transformar la petición al formato del protocolo simple (cadenas de texto separadas por espacios y \n )
//asegurandoce que la petición tiene el formato correcto
func (p *peticion) formatear()(error string){
	
	p.msjFormateado = p.codigo
	if len(p.regla) != len(p.campos){
		return "Faltan campos"
	}
	
	for k, v := range p.regla{ //la regla determina el orden exacto y los campos exactos
		
		if valorCampo, ok := p.campos[v]; !ok { //si el campo no existe
			return "El campo " + p.regla[k] + " es obligatorio"
		}else{
			p.msjFormateado += " " + valorCampo
		} 
	
	}
	p.msjFormateado += "\n" //\n para indicar el fin del mensaje
	return "" //sin errores

}


//Transforma una respues recivida en formato de protocolo simple a un mapa
//ademas verifica si la respuesta coincide con lo esperado para la petición
func (p *peticion) desformatear(psformat string)(rsp []string, error string){

	temp := strings.Trim(psformat,"\n\r") //quita los caracteres de fin de cadena
	r := strings.Split(temp," ") 
	//r[0] = código de respuesta
	 if v, ok := p.respuestas[r[0]]; !ok && !v { //si es una respuesta inesperada para esta petición
	 	return rsp, "Respuesta inesperada" //vacio y un error
	 } 
	 //si la respuesta es valida para esta petición puedo continuar
	 return r, error //vacio

}


//Recive un mensaej con formato de protocolo simple y lo transforma en un []strings
func Desformatear(psformat string)(rsp []string){
	temp := strings.Trim(psformat,"\n\r") //quita los caracteres de fin de cadena
	r := strings.Split(temp," ") 
	return r
}

//transforma el mensaje en formato de protocolo simple a un stream de bytes
func (p *peticion) marshal()(bstream []byte){
	bstream = []byte(p.msjFormateado)
	return bstream
}

//transforma un mensaje en formato de protocolo simple(simple cadena de texto) y lo transforma a un 
//stream de bytes
func marshal(msj string)(bstream []byte){
	//mejor fmt?
	bstream = []byte(msj)
	return bstream
}


//recive un stream de bytes y los estructura segun el protocolo simple (simples cadenas de texto)
func unmarshal(bstream []byte)(mensaje string){
	mensaje = string(bstream)
	return mensaje
}



//Recive un mensaje con formato de protocolo simple para enviar y devuelve la respuesta obtenida.
//Retorna la respuesta y un error vacio si todo sale bien, si no retorna la descripcion de un error como string. 
//Esta funcion es para coordinación interna del protocolo, pero por ahora se debe poder usar fuera de este paquete
func ReplicaRecive(psmsj string, conn net.Conn)(rsp []string, error string){
	bstream := []byte(psmsj)
	_ , err := conn.Write(bstream)
	if err != nil{
		if err == io.EOF{
			//la conexion se perdio //simplemente indica un error
			return rsp, "conndead"
		}
		//otro tipo de error
		return rsp, "otro"
	}
	//el mensaje se envio


	//se espera la respuesta
	buff := make([]byte, bufferMsg)
	n , err := conn.Read(buff)
	if err != nil{
		if err == io.EOF{
			//la conexion se perdio //simplemente indica un error
			return rsp, "conndead"
		}
		//otro tipo de error
		return rsp, "otro"
	}	
	//la respuesta se recivio 

	temp := string(buff[:n])
	rsp = strings.Split(temp, " ") //el formato del <protocolo simple> es strings separados por espacios
	return rsp, ""
}


//Envia un mensaje sin esperar respuesta.
//El error es un string vacio si todo sale bien
func ReplicaTermina(psmsj string, conn net.Conn)(error string){
	bstream := []byte(psmsj)
	_ , err := conn.Write(bstream)
	if err != nil{
		if err == io.EOF{
			//la conexion se perdio //simplemente indica un error
			return "conndead"
		}
		//otro tipo de error
		return "otro"
	}
	//el mensaje se envio
	return ""
}

//Espera una respuesta.
//Retorna la respuesta y un error vacio si todo sale bien. En tro caso retorna la descripcion del error en string
func ReciveTermina(conn net.Conn)(rsp []string, serr string){

	//se espera la respuesta
	buff := make([]byte, bufferMsg)
	n , err := conn.Read(buff)
	if err != nil{
		if err == io.EOF{
			//la conexion se perdio //simplemente indica un error
			return rsp, "conndead"
		}
		//otro tipo de error
		return rsp, "otro"
	}	
	//la respuesta se recivio 

	temp := string(buff[:n])
	rsp = strings.Split(temp, " ") //el formato del <protocolo simple> es strings separados por espacios
	return rsp, ""
}