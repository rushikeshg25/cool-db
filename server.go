package main

import (
	"log"
	"net"
)

func main(){
	conn,err:=net.Listen("tcp",":8989")
	if err!=nil{
		log.Fatalf("Error listening on port %s: %s\n","8989",err)
	}
	_, err = conn.Accept()
	if err != nil {
		log.Fatalf("Error accepting connection: %s\n", err)
	}
}