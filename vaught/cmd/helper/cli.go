/*
Copyright © 2024 Ayobami Bamigboye <ayo@greystein.com>
*/
package helper

import "os/exec"

const (
	ParentCommand  = "vault"
	InitSubCommand = "init"
)

func Init(args ...string) ([]byte, error) {
	fullArgs := append([]string{InitSubCommand}, args...)
	cmd := exec.Command(ParentCommand, fullArgs...)
	return cmd.Output()
}
