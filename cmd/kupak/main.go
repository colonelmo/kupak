package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"git.cafebazaar.ir/alaee/kupak"

	"github.com/codegangsta/cli"
	"github.com/ghodss/yaml"
)

var manager *kupak.Manager

func main() {
	kubectl, err := kupak.NewKubectlRunner()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	manager, err = kupak.NewManager(kubectl)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	app := cli.NewApp()
	app.Name = "kupak"
	app.Usage = "Kubernetes Package Manager"
	app.Version = "0.1"
	app.Commands = []cli.Command{
		{
			Name:    "paks",
			Aliases: []string{"p"},
			Usage:   "list all available paks of specified repo",
			Action:  paks,
		},
		{
			Name:    "install",
			Aliases: []string{"i"},
			Usage:   "install the specified pak (full url or a plain name that exists in specified repo)",
			Action:  install,
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list all installed packages",
			Action:  list,
		},
		{
			Name:    "spec",
			Aliases: []string{"s"},
			Usage:   "print specification of a pak",
			Action:  spec,
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "repo, r",
			Value:  "src/kupak/example/index.yaml",
			Usage:  "specify repo url",
			EnvVar: "KUPAK_REPO",
		},
		cli.StringFlag{
			Name:   "namespace",
			Value:  "default",
			Usage:  "namespace",
			EnvVar: "KUPAK_NAMESPACE",
		},
	}
	app.Run(os.Args)
}

func install(c *cli.Context) {
	pakURL := c.Args().First()
	valuesFile := c.Args().Get(1)
	if pakURL == "" {
		fmt.Fprintln(os.Stderr, "please specify the pak")
		os.Exit(-1)
	}

	pak, err := kupak.PakFromURL(pakURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// read value file
	var valuesData []byte
	if valuesFile == "" {
		valuesData, err = ioutil.ReadAll(os.Stdin)
	} else {
		valuesData, err = ioutil.ReadFile(valuesFile)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	values := make(map[string]interface{})
	err = yaml.Unmarshal(valuesData, &values)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	_, err = manager.Install(pak, c.GlobalString("namespace"), values)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}

func paks(c *cli.Context) {
	repo, err := kupak.RepoFromURL(c.GlobalString("repo"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	for i := range repo.Paks {
		fmt.Println("- Name:", repo.Paks[i].Name)
		fmt.Println("  Version:", repo.Paks[i].Version)
		fmt.Println("  URL:", repo.Paks[i].URL)
		if len(repo.Paks[i].Tags) > 0 {
			fmt.Println("  Tags:", "["+strings.Join(repo.Paks[i].Tags, ", ")+"]")
		}
		fmt.Println(" ", strings.Trim(repo.Paks[i].Description, "\n"))
	}
}

func list(c *cli.Context) {
	paks, err := manager.Installed(c.GlobalString("namespace"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	for i := range paks {
		fmt.Printf("Pak URL:  %s\n", paks[i].PakURL)
		fmt.Printf("Group: %s\n", paks[i].Group)
		fmt.Printf("Objects:\n")
		for j := range paks[i].Objects {
			obj := paks[i].Objects[j]
			md, _ := obj.Metadata()
			fmt.Printf("\t(%s) %s\n", md.Kind, md.Name)
			if md.Kind == "Pod" {
				status, _ := obj.Status()
				fmt.Printf("\t  State:     %s\n", status.Phase)
				fmt.Printf("\t  Pod IP:    %s\n", status.PodIP)
				if status.Reason != "" {
					fmt.Printf("\t  Reason:  %s\n", status.Reason)
				}
				if status.Message != "" {
					fmt.Printf("\t  Message: %s\n", status.Message)
				}
			}
		}
		fmt.Println()
	}
}

func spec(c *cli.Context) {
	pakURL := c.Args().First()
	if pakURL == "" {
		fmt.Fprintln(os.Stderr, "please specify the pak")
		os.Exit(-1)
	}
	pak, err := kupak.PakFromURL(pakURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: ", err)
		os.Exit(-1)
	}
	fmt.Println("Name:", pak.Name)
	fmt.Println("Version:", pak.Version)
	if len(pak.Tags) > 0 {
		fmt.Println("Tags:", "["+strings.Join(pak.Tags, ", ")+"]")
	}
	fmt.Println(strings.Trim(pak.Description, "\n"))

	fmt.Println("\nProperties:")
	for i := range pak.Properties {
		property := pak.Properties[i]
		fmt.Println(" - Name:", property.Name)
		fmt.Println("   Description:", strings.Trim(property.Description, "\n"))
		fmt.Println("   Type:", property.Type)
		if property.Default != nil {
			fmt.Println("   Default:", property.Default)
		}

	}
}
