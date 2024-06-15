package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
)

func runCmd(w io.Writer, args ...string) () {
	cmd := exec.Command("./tesla-cmd", args...)
	stdout, err := cmd.Output()

	if err != nil {
		e := err.(*exec.ExitError)

		fmt.Fprintf(w, string(stdout))
		fmt.Fprintf(w, string(e.Stderr))
		fmt.Fprintf(w, err.Error())
		return
	}

	// Print the output
	fmt.Fprintf(w, string(stdout))
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/charging-set-amps/{amps}", func(w http.ResponseWriter, r *http.Request) {
		amps := r.PathValue("amps")
		runCmd(w, "charging-set-amps", amps)
	})

	mux.HandleFunc("/charging-start", func(w http.ResponseWriter, r *http.Request) {
		runCmd(w, "charging-start")
	})

	mux.HandleFunc("/charging-stop", func(w http.ResponseWriter, r *http.Request) {
		runCmd(w, "charging-stop")
	})

	mux.HandleFunc("/charging-start-stop/{value}", func(w http.ResponseWriter, r *http.Request) {
		value := r.PathValue("value")
		if value == "start" || value == "true" || value == "on" || value == "1" {
			runCmd(w, "charging-start")
		} else {
			runCmd(w, "charging-stop")
		}
	})

	//Charging port open/close
	mux.HandleFunc("/charge-port-open", func(w http.ResponseWriter, r *http.Request) {
		runCmd(w, "charge-port-open")
	})

	mux.HandleFunc("/charge-port-close", func(w http.ResponseWriter, r *http.Request) {
		runCmd(w, "charge-port-close")
	})

	//Charging port open/close
	mux.HandleFunc("/charge-port-open-close/{value}", func(w http.ResponseWriter, r *http.Request) {
		value := r.PathValue("value")
		if value == "open" || value == "true" || value == "on" || value == "1" {
			runCmd(w, "charge-port-open")
		} else {
			runCmd(w, "charge-port-close")
		}
	})

	http.ListenAndServe(":8090", mux)
}
