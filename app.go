package main

import (
	"fmt"
	"github.com/Packetify/packetify/cmd"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"log"
)

func main() {
	cmd.Execute()

	//go Radius()
	//hst := EapHostapd{AuthServer_addr: "127.0.0.1", AuthServer_port: "2021", AuthServer_secret: "secret"}
}

func Radius() {
	handler := func(w radius.ResponseWriter, r *radius.Request) {
		username := rfc2865.UserName_GetString(r.Packet)
		password := rfc2865.UserPassword_GetString(r.Packet)

		fmt.Println(username, password)

		var code radius.Code
		if username == "victor" && password == "12345678" {
			code = radius.CodeAccessAccept
		} else {
			code = radius.CodeAccessReject
		}
		log.Printf("Writing %v to %v", code, r.RemoteAddr)
		w.Write(r.Response(code))
	}

	server := radius.PacketServer{
		Handler:      radius.HandlerFunc(handler),
		SecretSource: radius.StaticSecretSource([]byte(`secret`)),
	}

	log.Printf("Starting server on :1812")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
