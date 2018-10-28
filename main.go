package main

import (
	"log"
	"net/http"
	"os"
	"fmt"
	"github.com/marni/goigc"
	"encoding/json"


	//"github.com/gin-gonic/gin"
	//_ "github.com/heroku/x/hmetrics/onload"
)

const root = "/igcinfo"


type Meta struct {
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
	Version string `json:"version"`
}


func metaHandler(w http.ResponseWriter, r * http.Request){
	meta := Meta{
	Uptime:  "time",//calculateDuration(time.Since(startTime)),
	Info:    "Service for IGC tracks.",
	Version: "v1"}

	m, err := json.MarshalIndent(&meta, "", "    ")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	fmt.Fprintf(w, string(m))
}


func inputHandler(w http.ResponseWriter, r * http.Request){
	fmt.Fprintln(w, "GET igc")
	igcURL := "http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc"
	track, _ := igc.ParseLocation(igcURL)

	fmt.Fprintln(w, track.Pilot)
	fmt.Fprintln(w, len(track.Pilot))
}


func argsHandler(w http.ResponseWriter, r * http.Request){
	fmt.Fprintln(w, "ARGS")
}



func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	http.HandleFunc(root + "/api", metaHandler)
	http.HandleFunc(root + "/api/igc", inputHandler)
	http.HandleFunc(root + "/api/igc/", argsHandler)
	http.ListenAndServe(":" +port, nil);
}
