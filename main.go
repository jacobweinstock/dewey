package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/gin-gonic/gin"

	"github.com/pkg/errors"
	"github.com/tinkerbell/tink/protos/events"
	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc"
)

type pixecoreResponse struct {
	Kernel  string   `json:"kernel"`
	Initrd  []string `json:"initrd,omitempty"`
	Cmdline string   `json:"cmdline,omitempty"`
	Message string   `json:"message,omitempty"`
}

func main() {
	tinkServer := os.Getenv("DEWEY_TINK_SERVER")

	r := gin.Default()
	r.GET("/v1/boot/:mac", func(c *gin.Context) {
		fullURI := c.Request.RequestURI
		mac := path.Base(fullURI)
		_, _, hClient, _, err := Setup(tinkServer)
		if err != nil {
			c.JSON(500, "could not connect to tink server")
			return
		}
		hw, err := getTinkHardwareByMac(hClient, mac)
		if err != nil {
			c.JSON(400, "not found")
			return
		}
		kernel := hw.GetNetwork().GetInterfaces()[0].GetNetboot().GetOsie().GetKernel()
		initrd := hw.GetNetwork().GetInterfaces()[0].GetNetboot().GetOsie().GetInitrd()
		fmt.Println(kernel)
		fmt.Println(initrd)
		var resp pixecoreResponse
		if initrd == "" {
			resp = pixecoreResponse{
				Kernel: kernel,
			}
		} else {
			resp = pixecoreResponse{
				Kernel:  kernel,
				Initrd:  []string{initrd},
				Cmdline: "ip=dhcp modules=loop,squashfs,sd-mod,usb-storage parch=x86_64 packet_action=workflow facility=onprem plan=c2.medium.x86 initrd=initramfs-x86_64 console=tty0 console=ttyS1,115200",
			}
		}

		c.JSON(200, resp)
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func getTinkHardwareByMac(hClient hardware.HardwareServiceClient, mac string) (*hardware.Hardware, error) {
	//fmt.Println("talking to tink")

	// get hardware
	in := &hardware.GetRequest{Mac: mac}
	resultsHardware, err := hClient.ByMAC(context.TODO(), in)
	if err != nil {
		return nil, err
	}

	return resultsHardware, nil
}

// GetConnection returns a gRPC client connection
func GetConnection(tinkGrpcAuthority string) (*grpc.ClientConn, error) {
	var dialOpts grpc.DialOption
	dialOpts = grpc.WithInsecure()

	conn, err := grpc.Dial(tinkGrpcAuthority, dialOpts)
	if err != nil {
		return nil, errors.Wrap(err, "connect to tinkerbell server")
	}
	return conn, nil
}

// Setup : create a connection to server
func Setup(tinkGrpcAuthority string) (template.TemplateServiceClient, workflow.WorkflowServiceClient, hardware.HardwareServiceClient, events.EventsServiceClient, error) {
	conn, err := GetConnection(tinkGrpcAuthority)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return template.NewTemplateServiceClient(conn), workflow.NewWorkflowServiceClient(conn), hardware.NewHardwareServiceClient(conn), events.NewEventsServiceClient(conn), nil
}
