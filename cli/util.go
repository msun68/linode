package cli

import (
	"github.com/cheynewallace/tabby"
	"github.com/linode/linodego"
	"github.com/sethvargo/go-password/password"
)

func generatePassword() (string, error) {
	generator, err := password.NewGenerator(&password.GeneratorInput{
		LowerLetters: "abcdefghijklmnopqrstuvwxyz",
		UpperLetters: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		Symbols:      "!\"#$%&'()*+,-./:;<=>?@[]^_`{|}~\\",
		Digits:       "0123456789",
	})

	if err != nil {
		return "", err
	}

	return generator.Generate(20, 5, 5, false, true)
}

func printInstances(instances ...linodego.Instance) {
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
