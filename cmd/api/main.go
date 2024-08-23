package main

import "log"

func main() {

	auth, err := InitializeAuthService()
	if err != nil {
		log.Fatal(err)
		return
	}

	if err = auth.LoadPolicy(); err != nil {
		log.Fatal(err)
		return
	}

}
