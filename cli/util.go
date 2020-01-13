package cli

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/cheynewallace/tabby"
	"github.com/linode/linodego"
	"github.com/sethvargo/go-password/password"
	"golang.org/x/crypto/ssh"
)

type bootstrapInfo struct {
	Label          string
	Login          string
	AuthorizedKeys []string
}

func Bootstrap(instance linodego.Instance, rootPass string, login string, authorizedKey []string) error {
	tmpl := template.Must(template.New("bootstrap").Parse(`
echo '{{.Label}}' > /etc/hostname
sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab
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
`))

	var script bytes.Buffer

	if err := tmpl.Execute(&script, bootstrapInfo{
		Label:          instance.Label,
		Login:          login,
		AuthorizedKeys: authorizedKey,
	}); err != nil {
		return err
	}

	var connection *ssh.Client

	for retries := 0; connection == nil; retries++ {
		var err error
		connection, err = ssh.Dial("tcp", instance.IPv4[0].String()+":22", &ssh.ClientConfig{
			User:            "root",
			Auth:            []ssh.AuthMethod{ssh.Password(rootPass)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			BannerCallback:  ssh.BannerDisplayStderr(),
		})
		if err != nil {
			log.Printf("%v. Retrying.", err)
		}
	}

	session, err := connection.NewSession()

	if err != nil {
		return err
	}

	defer session.Close()

	in, err := session.StdinPipe()

	if err != nil {
		return err
	}

	if err := session.RequestPty("xterm", 80, 40, ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		return err
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.Shell(); err != nil {
		return err
	}

	fmt.Fprintln(in, script.String())

	if err := session.Wait(); err != nil {
		return err
	}

	return nil
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
