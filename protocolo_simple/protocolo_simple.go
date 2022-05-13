package protocolo_simple

import (
	"net"
	"strings"
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
	S_SALIR_ACEPTADO = "ssa"
	S_SALIR_RECHAZADO = "ssr"
	//AVISOS DEL SERVIDOR
	S_IMPOSIBLE_CONTINUAR = "sic"


	//AVISOS DEL CLIENTE
	C_IMPOSIBLE_CONTINUAR = "cic"

	//PETICIONES DEL CLIENTE
	C_UNIR_CANAL = "cun"
	C_SALIR_CANAL = "csc"
	C_DESCONECTAR = "cc"
	C_ENVIAR_ARCHIVO = "cea"

	//puedo crear un mapa para las respuestas comunes del servidor. algo como map["sda"] = "El servidor te a desconectado"
)


var LOG []string //todos los mensajes que se envian al servidor y las respuestas que se reciven


func registrarMsj(msj string){
	count := len(LOG) //cuantos registros tengo en el historial
	LOG[count] = msj //agrega el mensaje a la última posición
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
		campos: make(map[string]string),
		orden: "unirCanal",
		codigo: C_UNIR_CANAL, 
		regla: []string{"nombreCanal"}, //un solo campo obligatorio (nombreCanal) en la posición 0
		respuestas:  map[string]bool{S_UNIR_ACEPTADO:true, S_UNIR_RECHAZADO:true},
		msjFormateado: "",
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
		return pet, "" //sin errores

	}
}

//Ejecuta todo el proceso para enviar esta peticon por la conexion. Los procesos son
//formatear la petcion > tranforma a bytes > enviar. Retorna la descripcion de un error si este ocurre.
func (p *peticion) Enviar()(error string){
	
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
//recivir > desformatear > contruir respuesta > return. Retorna una respuesta o un error si este sucede.
func (p *peticion) RecivirRespuesta()(rsp []string, error string){

	var buf [128]byte //128 es mas que suficiente para recivir una respuesta por ahora
	n, err := p.conn.Read(buf[:])
	if err != nil{
		return rsp, err.Error() //vacio y un error 
	}

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
	
	p.msjFormateado = p.codigo
	if len(p.regla) != len(p.campos){
		return "Faltan campos"
	}
	
	for k, v := range p.regla{ //la regla determina el orden exacto y los campos exactos
		
		if valorCampo, ok := p.campos[v]; !ok { //si el campo no existe
			return "El campo " + p.regla[k] + " es obligatorio"
		}else{
			p.msjFormateado = " " + valorCampo
		} 
	
	}
	p.msjFormateado += "\n" //fin de mensaje
	return "" //sin errores

}


//Transforma una respues recivida en formato de protocolo simple a un mapa
//ademas verifica si la respuesta coincide con lo esperado para la petición
func (p *peticion) desformatear(psformat string)(rsp []string, error string){

	r := strings.Split(psformat, " ")
	//r[0] = código de respuesta
	 if v, ok := p.respuestas[r[0]]; !ok && !v { //si es una respuesta inesperada para esta petición
	 	return rsp, "Respuesta inesperada" //vacio y un error
	 } 
	 //si la respuesta es valida para esta petición puedo continuar
	 return r, error //vacio

}



//Una coleccion de variables que rigen la estructura y comportamiento de las 
//respuesta del protocolo simple
type respuesta struct{

	codigo string

	campos map[string]string

	regla []string

}

func respuestaDesconectarAceptado()

func respuestaUnirAceptado()

func respuestaUnirRechazado()

func respuestaEnvioAceptado()

func respuestaEnvioRechazado()

func respuestaServidorImposibleContinuar() 


















//Recive un comado y unos parametros para transformar esa información al formato del 
//protocolo simple (cadenas de texto separadas por espacios)
//los comados y sus respesctivos parametros son:
//comando eviar : nombre-archivo, peso-archivo, nombre-canal. 
//comando unir : nombre-canal. 
//comado desconectar : (carga vacía). 
//comado salir-canal : nombre-canal. 
// func formatear(orden string, parametros map[string]string)(mensaje string, serror string){
// 	switch(orden){

// 	case "unir":
		
// 		//El formato de este mensaje es: código de unir <espacio> nombre-canal 
// 		if _, ok := parametros["canal"]; !ok{
// 			return "", "El indice [canal] no existe"
// 		}
// 		return C_UNIR_CANAL +" "+ parametros["canal"], ""

	
// 	case "enviar":
		
// 		//El formato de este mensaje es: código de enviar <espacio> nombre-canal <espacio> nombre-archivo <espacio> peso-archivo
// 		if _, ok := parametros["nombreCanal"]; !ok{
// 			return "", "El indice [nombreCanal] no existe"
// 		}

// 		if _, ok := parametros["nombreArchivo"]; !ok{
// 			return "", "El indice [nombreArchivo] no existe"
// 		}

// 		if _, ok := parametros["pesoArchivo"]; !ok{
// 			return "", "El indice [pesoArchivo] no existe"
// 		}

// 		return C_ENVIAR_ARCHIVO +" "+ parametros["nombreCanal"] +" "+ parametros["nombreArchivo"] +" "+ parametros["pesoArchivo"], ""
	
// 	case "desconectar":
		
// 		//El formato de este mensaje es: código de desconectar 
// 		return C_DESCONECTAR, ""

// 	case "salir-canal":
		
// 		//El formato de este mensaje es: código de salir-canal <espacio> nombre-canal
// 		if _, ok := parametros["canal"]; !ok{
// 			return "", "El indice -canal- no existe"
// 		}
// 		return C_SALIR_CANAL +" "+ parametros["canal"], ""

// 	case "terminar-envio":
		
// 		//El formato de este mensaje es : código de terminar envio
// 		return C_IMPOSIBLE_CONTINUAR, ""


// 	default:
// 		return "","Comando desconocido."
// 	}
// 	return "", "Comando no especificado" 
// }

// //Recive un mensaje con formato de protocolo simple (cadenas de texto separados por espacios) y
// //los transforma en datos manejables que incluyen, la respuesta del servidor y la carga si la tiene. 
// func desformatear(psmsj string)(rsp string, param map[string]string, serror string){
// 	registrarMsj(psmsj) //todos los mensajes entrantes se registran
// 	valores := strings.Split(psmsj, " ")
// 	cmd := valores[0]
// 	switch(cmd){

// 	case S_UNIR_ACEPTADO:
		
// 		//la carga debe estar vacía
// 		if len(valores) != 1{
// 			return "", param, "El formato recivido no ovedece las reglas del protocolo"
// 		}
// 		//retorna param y error vacios
// 		return  "Peticion unir aceptada por el servidor", param, ""

// 	case S_UNIR_RECHAZADO:
		
// 		//la carga debe estar vacía
// 		if len(valores) != 1{
// 			return "", param, "El formato recivido no ovedece las reglas del protocolo"
// 		}
// 		//retorna param y error vacios
// 		return  "Peticion unir rechazada por el servidor", param, ""

// 	case S_DESCONECTAR_ACEPTADO: 

// 		//la carga debe estar vacía
// 		if len(valores) != 1{
// 			return "", param, "El formato recivido no ovedece las reglas del protocolo"
// 		}
// 		//retorna param y error vacios
// 		return  "El servidor te ha desconectado", param, ""

// 	case S_ENVIO_APROBADO:

// 		//la carga debe estar vacía
// 		if len(valores) != 1{
// 			return "", param, "El formato recivido no ovedece las reglas del protocolo"
// 		}
// 		//retorna param y error vacios
// 		return  "El servidor esta esperando la transmision del archivo", param, ""

// 	case S_ENVIO_RECHAZADO: 
		
// 		//la carga debe estar vacía
// 		if len(valores) != 1{
// 			return "", param, "El formato recivido no ovedece las reglas del protocolo"
// 		}
// 		//retorna param y error vacios
// 		return  "El servidor rechazo la transmision del archivo", param, ""

// 	} 



// 	return cmd, param, "El servidor no envio una respuesta conocida"
// }


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
