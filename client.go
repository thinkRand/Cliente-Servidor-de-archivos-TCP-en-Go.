package main

import(
	"net"
	"log"
	// "bufio"
	"os"
	"io"
	// "strings"
)



func main(){
	conn, err := net.Dial("tcp",":8080")
	if err != nil{
		log.Fatal(err)
	}
	log.Println("Cliente activo")
	// archivo := "pic.png"
	// //Si el nombre del archivo tiene espacios los elimino con TrimSpace
	// ar, err := os.Open(strings.TrimSpace(archivo))
	// if err != nil{
	// 	log.Println("El archivo no se pudo leer")
	// }
	// defer ar.Close()
	// // arInfo, err := ar.Stat()
	// //evio informacion del archivo al servidor
	// // conn.Write([]byte(arInfo.Name())) 
	// // conn.Write([]byte(string(arInfo.Size()))) //el error de impresion tiene que ver con UTF-8
	// //Envio el archivo
	// n , err := io.Copy(conn, ar)
	// if err != nil{
	// 	log.Println("El archivo no se pudo enviar.")
	// }
	// log.Println("Archivo enviado")
	// log.Println("Bytes enviados", n)

	go respuestasServidor(conn)

	//la función de aqui abajo parece crear un flujo desde el Stdin a la conexion, no es como que lea lo que hay en stdin y lo envie si no que el
	//flujo siempre esta abierto.
	_, err = io.Copy(conn, os.Stdin)
	if err != nil{
		log.Fatal(err)
	}
	conn.Close()
}


func respuestasServidor(conn net.Conn){
	//escribe las respuestas del servidor en al terminal
	io.Copy(os.Stdout, conn)
}