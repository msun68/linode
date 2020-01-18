package cli

import (
	"bytes"
	"text/template"

	"github.com/cheynewallace/tabby"
	"github.com/linode/linodego"
	"github.com/sethvargo/go-password/password"
)

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

func PrintInstances(instances ...linodego.Instance) {
	t := tabby.New()
	t.AddHeader("ID", "LABEL", "REGION", "TYPE", "IMAGE", "STATUS", "IP")
	for _, instance := range instances {
		var ips []string
		for _, ip := range instance.IPv4 {
			ips = append(ips, ip.String())
		}
		ips = append(ips, instance.IPv6)
		t.AddLine(instance.ID, instance.Label, instance.Region, instance.Type, instance.Image, instance.Status, ips)
	}
	t.Print()
}
