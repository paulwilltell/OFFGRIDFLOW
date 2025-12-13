package main

import (
    "context"
    "flag"
    "fmt"
    "log"

    "github.com/example/offgridflow/internal/db"
)

func main() {
    email := flag.String("email", "test@offgridflow.com", "email to lookup")
    flag.Parse()

    dsn := "postgresql://offgridflow:changeme@localhost:5432/offgridflow?sslmode=disable"
    database, err := db.Connect(context.Background(), db.Config{DSN: dsn})
    if err != nil {
        log.Fatalf("connect failed: %v", err)
    }
    defer database.Close()

    var id, tenantID, emailOut, name, passwordHash, role, roles string
    var isActive bool
    query := `SELECT id, tenant_id, email, name, password_hash, role, roles, is_active FROM users WHERE email = $1`
    err = database.QueryRowContext(context.Background(), query, *email).Scan(&id, &tenantID, &emailOut, &name, &passwordHash, &role, &roles, &isActive)
    if err != nil {
        log.Fatalf("query failed: %v", err)
    }

    fmt.Println("id:", id)
    fmt.Println("tenant_id:", tenantID)
    fmt.Println("email:", emailOut)
    fmt.Println("name:", name)
    fmt.Println("password_hash:", passwordHash)
    fmt.Println("password_hash_len:", len(passwordHash))
    fmt.Println("role:", role)
    fmt.Println("roles:", roles)
    fmt.Println("is_active:", isActive)
}
