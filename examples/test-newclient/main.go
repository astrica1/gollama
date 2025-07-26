package main

import (
	"fmt"
	"github.com/astrica1/gollama"
)

func main() {
	// Test the new NewClient signature
	client1, _ := gollama.NewClient()
	fmt.Println("Default client:", client1.BaseURL())
	
	client2, _ := gollama.NewClient("http://test:8080")
	fmt.Println("Custom client:", client2.BaseURL())
	
	client3, _ := gollama.NewClient("")
	fmt.Println("Empty client:", client3.BaseURL())
}
