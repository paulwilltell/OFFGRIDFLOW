package main

import (
	"fmt"

	"github.com/example/offgridflow/internal/auth"
)

func main() {
	hash := "$2a$10$yHVLsPuFSBIHIY3ZCUxEL.pV8I1Bi1t.iJLR/4dZYGS/kjf97/lnu"
	plain := "SecurePass123!"
	ok := auth.CheckPassword(hash, plain)
	fmt.Println("check result:", ok)
}
