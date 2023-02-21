package main

import "github.com/dnitsch/reststrategy/kubebuilder-controller/cmd/controller"

func main() {
	// init loggerHere or in init function
	controller.Execute()
}
