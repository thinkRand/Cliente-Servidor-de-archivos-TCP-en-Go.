package cliente

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

	thisCli := bufio.NewScanner(os.Stdin)
	for thisCli.Scan(){
		msg := thisCli.Text()
	}
	conn.Close()
}