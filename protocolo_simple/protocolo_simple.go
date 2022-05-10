package protocolo_simple
//la implementación del protocolo se debe encargar de formatear los mensajes.
//Me refiero a que cualquier mensaje como "canal aprobado" debe ser reconocido
//si el comando es "canal apropiado" y la regla del protocolo exige que primero va el comando y
//luego va la carga entonces reconocerio mal el comando anterior.
//el protocolo debería ser un paquete aparte

// type protocolo struct{
// 	//constantes de mensajes

// 	//regla del protocolo. e.j [cmd][opciones][data]

// 	//methodo Dial, Listen, Accept, para trabajar con un objeto protocol.Conn
// }

//con estas constantes puedo cambiar el protocolo son facilidad

const(
//RESPUESTAS DEL SERVIDOR
	SERVIDOR_CANAL_APROBADO = "canalaprobado"
	SERVIDOR_CANAL_NOAPROBADO = "canalanoprobado"
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

	//RESPUESTAS DE LA TERMINAL
	CLIENTE_ERROR_NUM_PARAMETROS = "El numero de parametros es incorrecto"
	CLIENTE_ERROR_CMD = "El comando es invalido"

)