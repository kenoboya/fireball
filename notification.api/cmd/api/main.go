package main

import "notification-api/internal/api"

const configDIR = "../../configs/"
const envDIR = ".env"

func main() {
	api.Run(configDIR, envDIR)
}
