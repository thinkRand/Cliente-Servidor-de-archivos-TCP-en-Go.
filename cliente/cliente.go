package main

import(
	Ps "protocolo_simple"
	"net"
	"log"
	"bufio"
	"os"
	"io"
	"strings"
	"fmt"
)

var estado string //los estados posibles del cliente CONECTADO, UNIDO_A_CANAL, ACORDANDO_TRANSMISION

//Una structura a manera de abstracción para representar un conjunto de funciones que generan salidas
//a la red y reciven entradas. Son procesos intermediario entre el servidor y el cliente que 
//recive una orden y determina la forma de usar el protocolo simple para cumplirla. 
//Los resultados varian dependiendo de la orden recivida.
type IoCliente struct{
	conn net.Conn
}


//Utilisa el protocolo para crear la peticion adecuada y enviarla. Despues espera la respuesta
//y notifica el resultado obtenido
func (ioCli *IoCliente) UnirCanal(nombreCanal string)(exito bool, error string){
	//estado = "CONECTADO"
	
	var campos = map[string]string{"nombreCanal":nombreCanal} //campos necesarios para esta petición

	peticion, serr := Ps.NuevaPeticion(ioCli.conn, "unirCanal", campos)
	if serr != "" {
		log.Println("DEBUG:", serr)
		return
	}
	//la peticion paso la prueba
	serr = peticion.Enviar() 
	if serr != "" {
		log.Println("DEBUG:", serr)
		return 
	}

	rsp, err := peticion.RecivirRespuesta()
	if err != ""{
		log.Println("DEBUG:", err)
		return 
	}

	if rsp[0] == Ps.S_UNIR_ACEPTADO{
		return true, "" //exito
	}else if rsp[0] == Ps.S_UNIR_RECHAZADO{
		return false,"" //no exito xd
	}else{
		log.Println("DEBUG: Se recivio una respuesta incoerente")
		return false, "La respuesta recivida no es acorde a la peticon"
	}

}

//Funcion de prueba para enviar archivos al canal
func (ioC *IoCliente) EnviarArchivo(rutaArchivo, nombreArchivo, pesoArchivo, nombreCanal string){
	
	//primero sin usar el protocolo a manera de prueba
	// 1. enviar el archivo al servidor
	// 2. enviar al canal
	// 3. enviar al servidor y hacer echo al cliente sobre el canal lul!
 	peticion := Ps.C_ENVIAR_ARCHIVO + " " + nombreCanal + " " + nombreArchivo + " " + pesoArchivo + "\n"
	
 	n, err := ioC.conn.Write([]byte(peticion))
 	if err != nil{
 		log.Println(err)
 		return
 	}
 	log.Println("Petición para enviar archivo enviada al servidor, escrito ", n)


 	//verifico si acepto o rechazo el archivo
 	var buff [512]byte //da igual si lo hago un slice de una vez
 	n, err = ioC.conn.Read(buff[:])
 	if err != nil{
 		log.Println("Error al leer la respuesta del servidor:", err)
 		return
 	}

 	temp := string(buff[:n])
 	rsp := strings.Trim(temp, "\n")
 	rsps := strings.Split(rsp," ")
 	if rsps[0] == Ps.S_ENVIO_RECHAZADO{
 		log.Println("El servidor no acepto la transferecia")
 		return
 	}

	if rsps[0] != Ps.S_ENVIO_APROBADO{
		log.Println("Respuesta invalida. El servidor respondio con:", rsp)
		return
	}

	log.Println("El servidor acepto la transferencia del archivo")	

	//Le envio el archivo
	ar, err := os.Open(rutaArchivo)
	if err != nil{
		log.Println(err)
		return
	}
	defer ar.Close()

	bc, err := io.Copy(ioC.conn, ar)
	if err != nil{
		
		log.Println("El archivo no se pudo enviar.")
		log.Println(err)
		return

	}else{

		//espera confirmación de recepcion de archivo
		n, err = ioC.conn.Read(buff[:])
	 	if err != nil{
	 		log.Println("Error al leer la respuesta del servidor:", err)
	 		return
	 	}
	 	rsp := string(buff[:n])
	 	rsp = strings.Trim(rsp, "\n") 
	 	rsps := strings.Split(rsp," ")
	 	if rsps[0] != Ps.S_ARCHIVO_RECIVIDO{
	 		log.Println("Erro, el servidor respondion con:", rsp)	
	 		return
 		}
 		log.Println("El servidor indica que recivio el archivo y pide confirmacion")
	 	log.Println("El mensaje del servidor es:", rsps[1:])

		
		//Notifico al servidor que efectivamente envie todo el archivo
		n, err = ioC.conn.Write([]byte(Ps.C_ARCHIVO_ENVIADO+"\n"))
 		if err != nil {
 			log.Println(err)
 			return
 		}

 		log.Println("Confirmacion enviada::", Ps.C_ARCHIVO_ENVIADO)
		log.Println("Bytes enviados:", bc)
		log.Println("Lado del cliente termino")
	}
	//hasta aquí la primera prueba, que llegue al servidor
}


//Una funcion para probar si el canal recive. Despues sera borrada
func (ioCli *IoCliente) ToChan(msg string){
	//fuera de protocolo
	n, err := ioCli.conn.Write([]byte(msg+"\n"))
	if err != nil{
		log.Println("El mensaje no se pudo enviar por el canal")
	}
	log.Println("Bytes escritos", n)

	//leer la respuesta
	var niceEcho [512]byte 
	n, err = ioCli.conn.Read(niceEcho[:])
	if err != nil{
		log.Println(err)
		return
	}
	log.Println("FROM_S:",string(niceEcho[:n]))
}





var ioCliente IoCliente


func main(){

	conn, err := net.Dial("tcp","127.0.0.1:9999")
	if err != nil{
		log.Fatal(err)
	}
	defer conn.Close()
	estado = "CONECTADO" 
	log.Println("Conexion establecida con el servidor")
	ioCliente.conn = conn

	terminal := bufio.NewScanner(os.Stdin)
	for terminal.Scan(){
		entrada := terminal.Text() //lee una cadena de texto hasta \n
		validarEntrada(entrada) //un comando a la vez
	}

}


//Recive las entradas desde la terminal para validarlas.
//Ejecuta dos niveles de validación: valida el formato, y valida los parametros
func validarEntrada(entrada string){
	divisionEntrada := strings.Split(entrada, " ")
	comando := strings.ToUpper(divisionEntrada[0]) //no case sensitive
	campos := make(map[string]string) //lis de campos que acompañan al comando, varian segun el comando

	switch(comando){
	case "UNIR":
		
		// se espera : UNIR <espeacio> canal.
		if len(divisionEntrada) != 2{
			log.Println("Entrada invalida. Se requiere: unir <espacio> nombre-canal")
			return
		}
		campos["nombreCanal"] = divisionEntrada[1]
		gestionar("unir", campos)


	case "ENVIAR":

		//solo se puede enviar si se esta en el estado de unido a canal
		if estado != "UNIDO_A_CANAL"{
			log.Println("No estas unido a un canal")
			return
		}

		// se espera : ENVIAR <espeacio> canal <espacio> ruta-archivo. Esto es independiente del protocolo simple
		if len(divisionEntrada) != 3{
			log.Println("Entrada invalida. Se requiere: enviar <espacio> nombre-canal <espacio> ruta-archivo")
			return
		}

		//nivel 2 de validacion: se comprueba que el archivo existe y es legible.
		rutaArchivo := divisionEntrada[2]
		ar, err := os.Open(rutaArchivo)
		if err != nil{
			log.Println("Error al intentar abrir la ruta del archivo")
			log.Println(rutaArchivo)
			log.Println(err)

			return
		}
		arInfo, err := ar.Stat()
		if err != nil{
			log.Println("Error al intentar leer la estructura del archivo")
			return
		}
		ar.Close()
		//validación completada

		nombreCanal := divisionEntrada[1]
		nombreArchivo := arInfo.Name()
		pesoArchivo := fmt.Sprintf("%d", arInfo.Size()) //el peso se envÍa en string
		
		campos["rutaArchivo"] = rutaArchivo
		campos["nombreArchivo"] = nombreArchivo 
		campos["pesoArchivo"] = pesoArchivo 
		campos["canal"] = nombreCanal
		gestionar("enviar", campos)

	case "SALIR":
		
		//Para salir hay dos opciones: salir del canal o salir completamente del programa
		if len(divisionEntrada) > 2 {
			log.Println("Demaciados campos para el comado salir: se espera salir <espacio> nombre-canal o")
			log.Println("salir (sin campos) para salir completamente del programa")
			return
		} 

		
		if len(divisionEntrada) == 1 {
			if estado == "UNIDO_A_CANAL"{
				//primero lo saco del canal
				//luego lo saco del programa
				return
			}else if estado == "CONECTADO"{
				log.Println("Cerrando conexion con el servidor...")
				return
			}
		}

		
		if len(divisionEntrada) == 2 {
			if estado != "UNIDO_A_CANAL"{
				log.Println("No estas unido a un canal")
				return
			}
			campos["canal"] = divisionEntrada[1]
	
		}


	case "DEBUG":
		debug()


	default:
		campos["msg"] = entrada
		gestionar("toChan", campos)

		// log.Println("Entrada invalida")
	}
}





//Utilisa el ioCliente para hacer las peticiones y determina que hacer con las respuestas.
//Es como un controller en el modelo MVC
func gestionar(orden string, campos map[string]string){
	switch orden{
	case "unir":
		
		exito, serr := ioCliente.UnirCanal(campos["nombreCanal"])
		if serr != "" {
			log.Println(serr)
			return
		} 

		if exito {
			estado = "UNIDO_A_CANAL"
			log.Println("Te uniste al canal")
		}else{
			//continua en estado CONECTADO
			log.Println("Fallo. No estas en el canal")
		}

	case "salir":


	case "toChan":		
		ioCliente.ToChan(campos["msg"])


	case "enviar":

		ioCliente.EnviarArchivo(campos["rutaArchivo"], campos["nombreArchivo"], campos["pesoArchivo"], campos["canal"])

	}

}


func debug(){
	log.Println(Ps.LOG)
}