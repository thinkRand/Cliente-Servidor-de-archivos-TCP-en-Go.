package main

import(
	"log"
)


type archivo struct{
	
	nombre string //el nombre del archivo que se va a redirigir 
	
	peso string //el peso del arhcivo
	
	ruta string //la ruta del archivo

}


//Un canal es una entidad con capacidad para guardar referencias de clientes dentro de si
//y de dar las ordenes necesarias para controlar la intereaccion de los clientes, solo debería ejercer funciones de control, 
//las funciones de gestión las debería hacer el intermediario, aqui llamadao ioServidor. 
//Es un mandamás, o mejor dicho es quién manda. Se encarga de dar ordenes al intermediario y de atender solicitudes solo cuando esta desocupado, si esta ocupado
//responde con OCUPADO, una forma simple de controlar el acceso al medio.
type Canal struct{
	
	nombre string //El nombre de este canal

	ocupado bool //OCUPADO O LIBRE, indica si el canal está OCUAPDO atendiendo o esperando un proceso o si esta LIBRE
	
	clientes map[*Cliente]bool //Registro de los clientes en este canal

	grupoReceptor map[*Cliente]bool //Un registro de clienes en el canal en un momento específico
	
	unir chan *Cliente //Recive todas la peticiones para unirce a este canal

	salir chan *Cliente	//Recive todas las peticiones para salir de este canal

	listo chan bool //Para sincronisar

	archivoParaRedirigir  *archivo  //Los datos del archivo que se va a redirigir actualmente
}


//Inicia la rutina para atender solicitudes de registro y salida de los clientes. 
//Si es un cliente nuevo lo registra. Si un cliente pidiendo salirce del canal lo elimina del registro
func (canal *Canal) Iniciar(){
	log.Println("El", canal.nombre," esta abierto")
	for{
		select{
		
		case cliente := <-canal.unir:
		
			canal.clientes[cliente] = true //Regitra al cliente en el canal
			log.Println(canal.nombre,"::", cliente.conn.RemoteAddr().String(),"se unio")
			canal.listo<- true // Al hacer esto solo se registra un cliente a la vez. No hay race condition?
		
		case cliente := <-canal.salir:
		
			delete(canal.clientes, cliente) //este canal ya no conoce al cliente
			canal.listo<- true

		}
	}
}


//Ordena al intermediario que se encargue de gestionar la redirección de un archivo
//particular sobre un grupo de clientes especificos del momento cuando se soliciato
//enviar dicho archivos desde un cliente.
//Retorna true si la orden se procesa correctamente, false en otro caso.
func (canal *Canal)  OrdenarRedireccion() bool{

	if canal.grupoReceptor != nil && canal.archivoParaRedirigir.nombre != "" { //si el archivo existe
		exito := Redirige(canal.grupoReceptor, canal.archivoParaRedirigir)
		if exito {
			return true
		}
		return false
	}else{
		return false
	} 
}



//Pregunta al canal si puede recivir un archivo en este canal.
//Si el canal esta ocupado en algun proceso retornara "OCUPADO",
//de otra forma se ocupa a atender la peticion y cuando termine retorna una cadena vacia.
//Utilisa al intermediario para ordenarle lo que debe hacer.
func (canal *Canal) GestionaRecivir(cliente *Cliente, nombreArchivo, pesoArchivo string) string {
	
	//control de acceso al medio, garantiza atender a un solo cliente a la vez
	//mu lock
	if canal.ocupado {
		//mu unlock
		return "OCUPADO"
	}
	canal.ocupado = true
	//mu unlock
	
	defer func(){
		//deja todo como estaba antes de empesar
		// os.Remove(canal.archivoParaRedirigir.ruta) //los archivos que se distribuye satisfactoriamente y los que no se eliminan
		canal.archivoParaRedirigir = &archivo{}
		canal.grupoReceptor = nil
		//mu lock, para que justo el primer cliente que pida enviar pueda enviar a partir de ahora
		canal.ocupado = false 
		//mu unlock
	}()
	log.Println(canal.nombre,":: esperando archivo desde cliente...", cliente.conn.RemoteAddr().String())
	//mu lock, algun cliente se puede estar registrando o saliendo en este preciso momento
	canal.grupoReceptor = canal.clientes  //Todos los clientes conectados en este momentos, solo ellos reciviran el archivo
	//mu unlock
	delete(canal.grupoReceptor, cliente) // tengo que quitar al cliente que envio el archivo
	exito := ReciveArchivo(canal.archivoParaRedirigir, cliente, nombreArchivo, pesoArchivo) //Negocia la subida del archivo y lo almacena en la variable
	if !exito {
		log.Println(canal.nombre,":: fallo al intentar recivir el archivo")
		return ""
	}

	log.Println(canal.nombre,":: redireccionando archivo recivido...")
	exito = canal.OrdenarRedireccion()
	if exito {
		log.Println(canal.nombre,":: archivo redirigido")
		return ""
	}
	log.Println(canal.nombre,":: fallo al intentar redirigir el archivo")
	return ""
}