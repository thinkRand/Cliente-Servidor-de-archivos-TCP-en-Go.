package main

import(
	"net"
	"log"
	"bufio"
	"os"
	"io"
	"strings"
)



func main(){
	conn, err := net.Dial("tcp",":8080")
	if err != nil{
		log.Fatal(err)
	}
	log.Println("Cliente activo...")
	archivo := "pic.png"
	//en caso de un nombre de archivo con espasios en el los elimino con TrimSpace
	ar, err := os.Open(strings.TrimSpace(archivo))
	if err != nil{
		log.Println("El archivo no se pudo leer")
	}
	defer ar.Close()
	// arInfo, err := ar.Stat()
	//evio informacion del archivo al servidor
	// conn.Write([]byte(arInfo.Name())) 
	// conn.Write([]byte(string(arInfo.Size()))) //el error de impresion tiene que ver con UTF-8

	//Envio el archivo
	_ , err = io.Copy(conn, ar)
	if err != nil{
		log.Println("El archivo no se pudo enviar.")
	}
	log.Println("Archivo enviado")
	thisCli := bufio.NewReader(os.Stdin)
	for {
		msg, _, _ := thisCli.ReadLine()
		conn.Write(msg)
	}
	conn.Close()
}