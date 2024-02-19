package main

import (
	"os"
	"os/exec"

	"github.com/3WDeveloper-GM/library_app/backend/config"
)

func main() {

	app := config.NewAppObject()

	app.SetConfigFlags()
	app.SetLogger()

	app.SetDB()
	defer app.Database.DB.Close()

	app.SetModels()

	//cleaning the terminal window
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	c.Run()

	//starting server.
	err := StartServer(app)
	if err != nil {
		app.Log.Panic().Err(err).Send()
	}
}
