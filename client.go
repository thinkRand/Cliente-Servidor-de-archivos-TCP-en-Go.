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
	log.Println("Cliente activo")
	archivo := "pic.png"
	//Si el nombre del archivo tiene espacios los elimino con TrimSpace
	ar, err := os.Open(strings.TrimSpace(archivo))
	if err != nil{
		log.Println("El archivo no se pudo leer")
	}
	defer ar.Close()
	// arInfo, err := ar.Stat()
	//evio informacion del archivo al servidor
	// conn.Write([]byte(arInfo.Name())) 
	// conn.Write([]byte(string(arInfo.Size()))) //el error de impresion tiene que ver con UTF-8
	var count int64
	//Envio el archivo
	n64 , err := io.Copy(conn, ar)
	count+=n64
	if err != nil{
		log.Println("El archivo no se pudo enviar.")
	}
	FDT := "<FDT>" //Fin de Transmisión
	//si le envio al servidor la cantidad exacta de bytes que debe esperar entonces sabra cuando terminar
	n, _ := conn.Write([]byte(FDT))
	count+=int64(n)
	log.Println("Archivo enviado")
	log.Println("Bytes enviados", count)
	thisCli := bufio.NewReader(os.Stdin)
	for {
		msg, _, _ := thisCli.ReadLine()
		conn.Write(msg)
	}
	conn.Close()
}