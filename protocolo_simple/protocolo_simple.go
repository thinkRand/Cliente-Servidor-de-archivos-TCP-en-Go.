package protocolo_simple

import (
	"net"
	"strings"
)


//reglas que describen el formato exacto que debe tener una petición
//con len(regla) se conoce el número exacto de campos que debe tener una petición
//con regla[k] = value se sabe que k es la posicion exacta que debe tener un valor
//con relag[] = value se conocen todos los campos que se requieren
var reglaUnirCanal = [1]string{"nombreCanal"} //un solo campo obligatorio (nombreCanal) en la posición 0
var reglaSalirCanal = [1]string{"nombreCanal"} //un solo campo obligatorio (nombreCanal) en la posición 0
var reglaDesconectar [0]string //no requiere campos
var reglaEnviarArchivo = [3]string{"nombreCanal", "nombreArchivo", "pesoArchivo"} //solo 3 campos obligatorios, en el orden descrito



//códigos de mensajes, son bastante sencillos y basados en texto
const(
	//RESPUESTAS DEL SERVIDOR 
	S_DESCONECTAR_ACEPTADO = "sda"
	S_UNIR_ACEPTADO = "sua"
	S_UNIR_RECHAZADO = "sur"
	S_ENVIO_APROBADO = "sea"
	S_ENVIO_RECHAZADO = "ser"

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

	campos map[string]string //lista de campos y sus valores

	regla []string //describe el orden exacto que deben tener los campos, y los campos exactos que debe tener la petición

	mFormateado string //la cadena con formato de protocolo simple (strings separados por espacios)

	respuestasEsperadas []string //lista de las respuestas esperadas para esta petición
}

//Ejecuta todo el proceso para enviar esta peticon por la conexion. Los procesos son
//formatear la petcion > tranforma a bytes > enviar. Retorna la descripcion de un error si este ocurre.
func (p *peticion) Enviar()(error string){
	
	p.formatear()
	bstream := marshal(p.mFormateado)
	_ , err := p.conn.Write(bstream)
	if err != nil{
		return err.Error()
	}		
	registrarMsj(p.mFormateado) //guardar la cadena enviada en el historial

	return "" //no hay error
}




//Recive la respuesta a una peticion ejecutando los procesos de 
//recivir > desformatear > contruir respuesta > return. Retorna una respuesta o un error si este sucede.
func (p *peticion) recivirRespuesta()(rsp respuesta, error string){

	buf := make([]byte, 128) //128 es mas que suficiente para recivir una respuesta por ahora
	n, err := p.conn.Read(buf)
	if err != nil{
		return rsp, err.Error() //vacio y un error 
	}

	umsj := unmarshal(buf[:n])
	registrarMsj(umsj) //actualisa el historial de mensajes 
	rsp, serr := p.desformatear(umsj)
	if serr != "" {
		return rsp, "No se puedo desformatear el mensje" //vacio y un error
	}

	return rsp, "" 
}


//Transformar la petición al formato del protocolo simple (cadenas de texto separadas por espacios)
//asegurandoce que la petición tiene el formato correcto
func (p *peticion) formatear()(error string){
	
	p.mFormateado = p.codigo
	if len(p.regla) != len(p.campos){
		return "Faltan campos"
	}
	
	for k, v := range p.regla{ //la regla determina el orden exacto y los campos exactos
		
		if valorCampo, ok := p.campos[v]; !ok { //si el campo no existe
			return "El campo " + p.regla[k] + " es obligatorio"
		}else{
			p.mFormateado = " " + valorCampo
		} 
	
	}
	return "" //sin errores

}

//Una coleccion de variables que rigen la estructura y comportamiento de las 
//respuesta del protocolo simple
type respuesta struct{

	codigo string //codigo de la respuesta

	carga string //el contenido del mensaje, usualmente innecesario





}

//Transforma un mensaje con formato de protocolo simple (cadenas de texto separados por espacios) en datos 
//manejables que incluyen la respuesta del servidor y la carga si la tiene.
//ademas de comprobar que el formato de la respuesta es el esperado 
//se supone que desformatear toma una cadena de texto con codigo de mensaje y carga del mensaje
//separadas por espacios y lo transforma en una respuesta valida lista para ser operada por
//el siguiente nivel de abstracción
//debería retornar un objeto respuesta valido o un error
//deberia verificar
//la cantidad de campos debe ser exacta a la cantidad de campos esperados para esta peticion
//dicha información esta contenida en los valores de la petición
//segundo cada campo debe contener un valor valido: por ejemplo se esperan siertos codigos exactos
//dependiendo de la peticion por lo que otros codigos deben arrorjar un error por código inesperado.
//el valor de los campos pues no tengo como comprobar que estám en el orden correcto, esto puede
//ser un problema más adelante, pero por ahora continuare con lo que tengo a la mano.

func (p *peticion) desformatear(psformat string)(rsp respuesta, error string){

}




//Ejecuta todos los procesos para entender una orden y gestion el envio de la misma
//reronar la cadena vacia "" si el proceso termina bien. De otra manera devuel una descripcion
//del error.
//deberia hacer un mapa del tipo map[string]func para manejar mas facil los comandos con ["enviar"](param)
//hacer envio debe hacer uso de la interface correspondiente para enviar, si es cliente o servidor
func HacerEnvio(orden string, parametros map[string]string, conn net.Conn)(error string){
	switch(orden){
		case "unir":
			//construyo la peticion
			//la envio
			//espero al respuesta
			//ejecuto acciones...

		case "enviar":
			

		case "desconectar":
		case "salir-canal":
		case "terminar-envio":
		case "crash":


		default:
			return "Orden desconocida"
	}
	return "Falta el parametro: orden"
}


//Ejecuta todos los procesa para entender una respuesta del servidor y entrega los resultados
//los procesos son recivir > unmarshal > desformatear > retornar respuesta
func LeerRespuesta(conn net.Conn)(respuesta string, parametros map[string]string, serror string){
	buf := make([]byte, 512) //limite de 511 bytes para la cadena del mensaje, el byte 1 es un codigo de respuesta
	n, err := conn.Read(buf)
	if err != nil{
		return "", parametros, err.Error() //valores vacios y un error
	}

	umsj := unmarshal(buf[0:n])
	rsp, param, serr := desformatear(umsj)
	if serr != "" {
		return "", parametros, "No se puedo desformatear el mensje" //valores vacios y un error
	}

	return rsp, param, ""
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
	//podíra usar fmt?
	bstream = []byte(msj+"\n")
	return bstream
}


//recive un stream de bytes y los estructura segun el protocolo simple (simples cadenas de texto)
func unmarshal(bstream []byte)(mensaje string){
	mensaje = string(bstream)
	return mensaje
}
