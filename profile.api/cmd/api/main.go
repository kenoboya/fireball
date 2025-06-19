package main

import "profile-api/internal/api"

const configDIR = "../../configs/"
const envDIR = ".env"

func main() {
	api.Run(configDIR, envDIR)
}
