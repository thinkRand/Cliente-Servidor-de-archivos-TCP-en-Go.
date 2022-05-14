package protocolo_simple

import (
	"net"
	"strings"
	// "log"
)


//reglas que describen el formato exacto que debe tener una petición. 
//con len(regla) se conoce el número exacto de campos que debe tener una petición. 
//con regla[k] = value se sabe que k es la posicion exacta que debe tener un valor. 
//con relag[] = value se conocen todos los campos que se requieren.
//intente que fueran arreglos para enfatisar su caracter estricto, pero los arreglos complicaron las cosas 
var reglaDesconectar []string //no requiere campos
var reglaEnviarArchivo = []string{"nombreCanal", "nombreArchivo", "pesoArchivo"} //solo 3 campos obligatorios, en el orden descrito



//códigos de mensajes, son bastante sencillos y basados en texto
const(
	//RESPUESTAS DEL SERVIDOR 
	S_DESCONECTAR_ACEPTADO = "sda"
	S_UNIR_ACEPTADO = "sua"
	S_UNIR_RECHAZADO = "sur"
	S_ENVIO_APROBADO = "sea"
	S_ENVIO_RECHAZADO = "ser"
	S_ARCHIVO_RECIVIDO = "sar"
	S_SALIR_ACEPTADO = "ssa"
	S_SALIR_RECHAZADO = "ssr"
	S_CODIGO_INVALIDO = "sci"
	//AVISOS DEL SERVIDOR
	S_IMPOSIBLE_CONTINUAR = "sic"


	//AVISOS DEL CLIENTE
	C_IMPOSIBLE_CONTINUAR = "cic"

	//PETICIONES DEL CLIENTE
	C_UNIR_CANAL = "cuc"
	C_SALIR_CANAL = "csc"
	C_DESCONECTAR = "cd"
	C_ENVIAR_ARCHIVO = "cea"
	C_ARCHIVO_ENVIADO = "cae"
	//puedo crear un mapa para las respuestas comunes del servidor. algo como map["sda"] = "El servidor te a desconectado"
)

const bufferMsg = 128 //128 bytes maximo por mensaje es mas que suficiente por ahora
var LOG = make([]string, 0, 1000) //Guarda los mensajes que viene y van entre el cliente y el servior hasta 1000, para probar


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


//Crea una peticion del tipo orden con los campos en campos y sobre la conexion en conn
func NuevaPeticion(conn net.Conn, orden string, campos map[string]string)(p peticion, error string){

	listaPeticiones["unirCanal"] = peticionUnirCanal
	listaPeticiones["salirCanal"] = peticionSalirCanal

	if pet, ok := listaPeticiones[orden]; !ok{
		return p,"La orden no existe"
	}else{
		//las peticiones vienen preformateadas y solo necesitan la conexion y los campos
		pet.conn = conn 
		pet.campos = campos
		// log.Println("en nueva peticon. la peticion tiene forma", pet)
		return pet, "" //sin errores

	}
}


//Ejecuta todo el proceso para enviar esta peticon por la conexion. Los procesos son
//formatear la petcion > marshall > enviar. Retorna la descripcion de un error si este ocurre.
func (p *peticion) Enviar()(error string){
	
	// log.Println("En Enviar...")
	p.formatear()
	bstream := marshal(p.msjFormateado)
	_ , err := p.conn.Write(bstream)
	if err != nil{
		return err.Error()
	}		
	registrarMsj(p.msjFormateado) //guardar la cadena enviada en el historial

	return "" //no hay error
}


//Recive la respuesta a una peticion ejecutando los procesos de 
//recivir > desformatear > unmarshall > return. Retorna una respuesta o un error si este sucede.
func (p *peticion) RecivirRespuesta()(rsp []string, error string){

	//error aqui, siempre arroja que la respuesta es inesperada
	var buf [bufferMsg]byte
	n, err := p.conn.Read(buf[:])
	if err != nil{
		return rsp, err.Error() //vacio y un error 
	}
	// log.Println("En recivir respuesta, recivido en bytes:", buf[:n])
	umsj := unmarshal(buf[:n]) //de bytes a cadena de texto con formato del protocolo
	registrarMsj(umsj) //actualisa el historial de mensajes 
	r, serr := p.desformatear(umsj) //de cadena de texto a []strings
	if serr != "" {
		return rsp, serr //vacio y un error
	}

	//creo que es buena idea devolver un exito o un fracaso
	//isi la respuesta es satisfactoria para esta petición return true
	//si no lo es return false
	return r, "" 
}


//Transformar la petición al formato del protocolo simple (cadenas de texto separadas por espacios y \n )
//asegurandoce que la petición tiene el formato correcto
func (p *peticion) formatear()(error string){
	
	// log.Println("En formatear...")
	//error el mensaje no esta siendo formateado correctamente
	//solo contiene el nombre del canal
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
	// log.Println("Al terminar la peticion tiene la forma:  ", p)
	p.msjFormateado += "\n" //\n para indicar el fin del mensaje
	// log.Println("Fin de formateo...")
	return "" //sin errores

}


//Transforma una respues recivida en formato de protocolo simple a un mapa
//ademas verifica si la respuesta coincide con lo esperado para la petición
func (p *peticion) desformatear(psformat string)(rsp []string, error string){

	//rastree el error hasta aqua
	temp := strings.Trim(psformat,"\n\r") //quita los caracteres de fin de cadena
	r := strings.Split(temp," ") 
	//r[0] = código de respuesta
	// log.Println("En desformatear...")
	// log.Println("cantidad del split:",len(r))
	// log.Println("r[0]:", r[0])
	// log.Println("p.respuestas:", p.respuestas)
	 if v, ok := p.respuestas[r[0]]; !ok && !v { //si es una respuesta inesperada para esta petición
	 	return rsp, "Respuesta inesperada" //vacio y un error
	 } 
	 //si la respuesta es valida para esta petición puedo continuar
	 return r, error //vacio

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

// type respuesta struct{

// 	codigo string

// 	campos map[string]string

// 	regla []string

// }