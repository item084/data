package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nimezhu/data"
	"github.com/nimezhu/snowjs"
	"github.com/urfave/cli"
)

const VERSION = "0.0.0"

func main() {
	app := cli.NewApp()
	app.Version = VERSION
	app.Name = "datam"
	app.Usage = "indexed file data manager tools"
	app.EnableBashCompletion = true //TODO
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Show more output",
		},
	}

	// Commands
	app.Commands = []cli.Command{
		{
			Name:   "start",
			Usage:  "start a server",
			Action: CmdStart,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input,i",
					Usage: "input data tsv",
					Value: "data tsv",
				},
				cli.IntFlag{
					Name:  "port,p",
					Usage: "data server port",
					Value: 5000,
				},
			},
		},
	}

	app.Run(os.Args)
}
func CmdStart(c *cli.Context) error {
	uri := c.String("input")
	port := c.Int("port")
	router := mux.NewRouter()
	snowjs.AddHandlers(router, "")
	AddStaticHandle(router)
	data.Load(uri, router) //TODO using only router not manager interface.
	log.Println("Listening...")
	log.Println("Please open http://127.0.0.1:" + strconv.Itoa(port))
	http.ListenAndServe(":"+strconv.Itoa(port), router)
	return nil
}
