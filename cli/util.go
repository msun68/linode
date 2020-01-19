package cli

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"sort"
	"text/template"

	"github.com/cheynewallace/tabby"
	"github.com/linode/linodego"
	"github.com/sethvargo/go-password/password"
)

type PrintInstancesOptions struct {
	Type    string
	Options map[string]*string
}

type bootstrapInfo struct {
	Label          string
	Login          string
	AuthorizedKeys []string
}

func GenerateBootstrapScript(label string, login string, authorizedKey []string) (string, error) {
	tmpl := template.Must(template.New("bootstrap").Parse(`#!/bin/bash
echo '{{.Label}}' > /etc/hostname
sed -i 's/#PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config
useradd -d /home/{{.Login}} -m {{.Login}}
mkdir /home/{{.Login}}/.ssh
{{range .AuthorizedKeys -}}
echo '{{.}}' >> /home/{{$.Login}}/.ssh/authorized_keys
{{- end}}
chmod 0700 /home/{{.Login}}/.ssh
chmod 0600 /home/{{.Login}}/.ssh/authorized_keys
chown -R {{.Login}}:{{.Login}} /home/{{.Login}}/.ssh
echo '{{.Login}} ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
reboot
`))

	var script bytes.Buffer

	if err := tmpl.Execute(&script, bootstrapInfo{
		Label:          label,
		Login:          login,
		AuthorizedKeys: authorizedKey,
	}); err != nil {
		return "", err
	}

	return script.String(), nil
}

func GeneratePassword() (string, error) {
	generator, err := password.NewGenerator(&password.GeneratorInput{
		LowerLetters: "abcdefghijklmnopqrstuvwxyz",
		UpperLetters: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		Symbols:      "!\"#$%&'()*+,-./:;<=>?@[]^_`{|}~\\",
		Digits:       "0123456789",
	})

	if err != nil {
		return "", err
	}

	return generator.Generate(40, 10, 10, false, true)
}

func GetIPAddresses() map[int][]linodego.InstanceIP {
	ips := make(map[int][]linodego.InstanceIP)

	if resp, err := linodeClient.ListIPAddresses(context.Background(), &linodego.ListOptions{}); err == nil {
		for _, ip := range resp {
			ips[ip.LinodeID] = append(ips[ip.LinodeID], ip)
		}
	}

	return ips
}

func Contains(a []string, x string) bool {
	for _, i := range a {
		if x == i {
			return true
		}
	}
	return false
}

func PrintInstances(instances []linodego.Instance, ips map[int][]linodego.InstanceIP, options *PrintInstancesOptions) {
	if options != nil {
		if options.Type == "ansible" {
			printInstancesAnsible(instances, ips, options.Options)
		} else {
			printInstancesTable(instances, ips)
		}
	} else {
		printInstancesTable(instances, ips)
	}
}

func printInstancesAnsible(instances []linodego.Instance, ips map[int][]linodego.InstanceIP, options map[string]*string) {

	var usePrivateIP bool

	if _, ok := options["use-private-ip"]; ok {
		usePrivateIP = true
	}

	var useIPv6 bool

	if _, ok := options["use-ipv6"]; ok {
		useIPv6 = true
	}

	invalidGroupChars := regexp.MustCompile(`[^\w]`)

	groups := make(map[string][]string)

	for _, instance := range instances {
		if !Contains(instance.Tags, "ansible") {
			continue
		}
		if instanceIPs, ok := ips[instance.ID]; ok {
			var address string
			for _, ip := range instanceIPs {
				if usePrivateIP {
					if ip.Type == linodego.IPTypeIPv4 && !ip.Public {
						address = ip.Address
						break
					}
				} else if ip.Public {
					if useIPv6 {
						if ip.Type == linodego.IPTypeIPv6 {
							address = ip.Address
							break
						}
					} else if ip.Type == linodego.IPTypeIPv4 {
						address = ip.Address
						break
					}
				}
			}

			label := invalidGroupChars.ReplaceAllString(instance.Label, "_")

			fmt.Println("[" + label + "]")
			fmt.Println(address)
			fmt.Println()

			groups[invalidGroupChars.ReplaceAllString(instance.Region, "_")] = append(groups[instance.Region], label)

			for _, tag := range instance.Tags {
				if tag != "ansible" {
					groups[invalidGroupChars.ReplaceAllString(tag, "_")] = append(groups[tag], label)
				}
			}
		}
	}

	var names []string

	for name := range groups {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		fmt.Println("[" + name + ":children]")
		for _, label := range groups[name] {
			fmt.Println(label)
		}
		fmt.Println()
	}
}

func printInstancesTable(instances []linodego.Instance, ips map[int][]linodego.InstanceIP) {
	t := tabby.New()
	t.AddHeader("ID", "LABEL", "REGION", "TYPE", "IMAGE", "STATUS", "PUBLIC IP", "PRIVATE IP", "TAG")
	for _, instance := range instances {
		var public []string
		var private []string
		if instanceIPs, ok := ips[instance.ID]; ok {
			for _, ip := range instanceIPs {
				if ip.Public {
					public = append(public, ip.Address)
				} else if ip.Type == linodego.IPTypeIPv4 {
					private = append(private, ip.Address)
				}
			}
		}
		t.AddLine(instance.ID, instance.Label, instance.Region, instance.Type, instance.Image, instance.Status, public, private, instance.Tags)
	}
	t.Print()
}
