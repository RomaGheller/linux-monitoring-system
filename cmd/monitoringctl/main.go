package main

import (
	"fmt"
	"os"

	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/api"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	//client, err := api.NewClient("/tmp/monitoring.sock")
	client, err := api.NewDefaultClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "services":
		if err := handleServices(client, os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

	case "reload":
		if err := client.Reload(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

	default:
		usage()
		os.Exit(1)
	}
}

func handleServices(
	client *api.Client,
	args []string,
) error {
	if len(args) == 0 {
		return fmt.Errorf("missing services command")
	}

	switch args[0] {
	case "list":
		return listServices(client)

	case "add":
		return addService(client, args[1:])

	case "ensure":
		return ensureService(client, args[1:])

	default:
		return fmt.Errorf("unknown services command %q", args[0])
	}
}

func listServices(client *api.Client) error {
	services, err := client.ListServices()
	if err != nil {
		return err
	}

	for _, service := range services {
		fmt.Printf("%-20s %s\n", service.Name, service.Type)
	}

	return nil
}

func addService(
	client *api.Client,
	args []string,
) error {
	name, serviceType, err := parseServiceArgs(args)
	if err != nil {
		return err
	}

	service := api.ServiceRequest{
		Name: name,
		Type: serviceType,
	}

	if err := client.AddService(service); err != nil {
		return err
	}

	fmt.Printf("Service %q registered\n", name)

	return nil
}

func ensureService(
	client *api.Client,
	args []string,
) error {
	name, serviceType, err := parseServiceArgs(args)
	if err != nil {
		return err
	}

	service := api.ServiceRequest{
		Name: name,
		Type: serviceType,
	}

	added, err := client.EnsureService(service)
	if err != nil {
		return err
	}

	if added {
		fmt.Printf("Service %q registered\n", name)
	} else {
		fmt.Printf("Service %q already registered\n", name)
	}

	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage:
  monitoringctl services list
  monitoringctl services add <name> --type <type>
  monitoringctl services ensure <name> --type <type>
  monitoringctl reload`)
}

func parseServiceArgs(args []string) (string, string, error) {
	var name string
	var serviceType string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--type":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("missing value for --type")
			}

			serviceType = args[i+1]
			i++

		default:
			if name != "" {
				return "", "", fmt.Errorf("unexpected argument %q", args[i])
			}

			name = args[i]
		}
	}

	if name == "" {
		return "", "", fmt.Errorf("expected service name")
	}

	if serviceType == "" {
		return "", "", fmt.Errorf("expected --type")
	}

	return name, serviceType, nil
}
