package main

import(
	"net"
	"log"
	"bufio"
	"os"
)



func main(){
	conn, err := net.Dial("tcp",":8080")
	if err != nil{
		log.Fatal(err)
	}
	log.Println("Cliente activo...")
	//el bufio funciona diferente con las conexiones que con la entrada de la terminal
	thisCli := bufio.NewReader(os.Stdin)
	//lo que pasa aquí es que la variable mensaje recive la linea, pero si ocurre un error msg tiene el error
	for {
		msg, _, _ := thisCli.ReadLine()
		conn.Write(msg)
	}
	conn.Close()
}