package main

import (
	"context"
	"log"

	"github.com/Aran404/Forwarder/api/server"
	"github.com/Aran404/Forwarder/api/types"
)

func main() {
	types.Clear()
	log.Println("Listening on port 3443")

	ctx := context.Background()
	c := server.NewClient(ctx)
	defer c.Close(ctx)
	c.Listen()
}
