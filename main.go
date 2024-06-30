package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/teslamotors/vehicle-command/pkg/connector/ble"
	"github.com/teslamotors/vehicle-command/pkg/protocol"
	"github.com/teslamotors/vehicle-command/pkg/protocol/protobuf/universalmessage"
	"github.com/teslamotors/vehicle-command/pkg/vehicle"
	"golang.org/x/sync/semaphore"
)

type CarCallback func(*vehicle.Vehicle, context.Context) error

func withCarConnection(cb CarCallback, needsInfotainment bool) (error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	vin := os.Getenv("TESLA_VIN")
	privateKeyFile := os.Getenv("TESLA_KEYFILE")

	privateKey, err := protocol.LoadPrivateKey(privateKeyFile);
	if err != nil {
		fmt.Println("Error loading private key: ", privateKeyFile, err)
		return err
	}

	//fmt.Println("Connecting to car with VIN", vin)
	conn, err := ble.NewConnection(ctx, vin)
	if err != nil {
		return err
	}
	defer conn.Close()

	//fmt.Println("Creating vehicle object...")
	car, err := vehicle.NewVehicle(conn, privateKey, nil)
	if err != nil {
		return nil
	}

	//fmt.Println("Connecting to car...")
	if err = car.Connect(ctx); err != nil {
		return err
	}
	defer car.Disconnect()

	if err = car.StartSession(ctx, []universalmessage.Domain{universalmessage.Domain_DOMAIN_VEHICLE_SECURITY}); err != nil {
		return errors.New("ErrNotAvailable")
	}

	if needsInfotainment {
		if err = car.StartSession(ctx, []universalmessage.Domain{universalmessage.Domain_DOMAIN_INFOTAINMENT}); err != nil {
			return errors.New("ErrAsleep")
		}
	}

	//fmt.Println("Running callback...")
	return cb(car, ctx)
}

type Handler struct {
	sem semaphore.Weighted //.NewWeighted(int64(1))
}

func newHandler() *Handler {
	return &Handler{sem: *semaphore.NewWeighted(int64(1))}
}

func (h *Handler) handleFunc(w http.ResponseWriter, r *http.Request, cb CarCallback, needsInfotainment bool) {
	err := h.sem.Acquire(r.Context(), 1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Error: ", err)
		fmt.Fprint(w, err)
		return
	}

	defer h.sem.Release(1)
	
	err = withCarConnection(cb, needsInfotainment)

	// We need to give the BLE device a second to recover for the next connection
	// otherwise subsequent connections will fail.
	time.Sleep(1 * time.Second)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Error: ", err)
		fmt.Fprint(w, err)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Success")
	}
}

func main() {
	mux := http.NewServeMux()

	var h = newHandler()

	mux.HandleFunc("/charging-set-amps", func(w http.ResponseWriter, r *http.Request) {	
		amps, _ := io.ReadAll(r.Body)
		intAmps, err := strconv.Atoi(string(amps))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: ", err)
			return
		}

		fmt.Println("Setting amps to: ", intAmps, " ...")
		h.handleFunc(w, r, func(v *vehicle.Vehicle, ctx context.Context) error {
			return v.SetChargingAmps(ctx, int32(intAmps))
		}, true)
	})

	mux.HandleFunc("/charging-start", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Starting charge ... ")
		h.handleFunc(w, r, func(v *vehicle.Vehicle, ctx context.Context) error {
			return v.ChargeStart(ctx)
		}, true)
	})

	mux.HandleFunc("/charging-stop", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Stopping charge ...")
		h.handleFunc(w, r, func(v *vehicle.Vehicle, ctx context.Context) error {
			return v.ChargeStop(ctx)
		}, true)
	})

	mux.HandleFunc("/charging-start-stop", func(w http.ResponseWriter, r *http.Request) {
		v, _ := io.ReadAll(r.Body)
		value := string(v)
		if value == "start" || value == "true" || value == "on" || value == "1" {
			fmt.Println("Starting charge ... ")
			h.handleFunc(w, r, func(v *vehicle.Vehicle, ctx context.Context) error {
				return v.ChargeStart(ctx)
			}, true)
		}else {
			fmt.Println("Stopping charge ... ")
			h.handleFunc(w, r, func(v *vehicle.Vehicle, ctx context.Context) error {
				return v.ChargeStop(ctx)
			}, true)
		}
	})

	mux.HandleFunc("/wake", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Waking up car ...")

		h.handleFunc(w, r, func(v *vehicle.Vehicle, ctx context.Context) error {
			return v.Wakeup(ctx)
		},false)
	})

	http.ListenAndServe(":8090", mux)
}
