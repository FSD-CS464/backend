package main

import (
    "fsd-backend/internal/routers"
)

func main() {
    r := routers.SetupRouter()

    if err := r.Run(":8080"); err != nil {
        panic(err)
    }
}