package main

import "auth-api/internal/api"

const configDIR = "../../configs/"
const envDIR = ".env"

func main() {
	api.Run(configDIR, envDIR)
}
