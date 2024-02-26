package initer

import "testing"

func TestInitOptions_Run(t *testing.T) {
	_ = InitOptions{
		args:            nil,
		debug:           false,
		storeStr:        "",
		gobLoc:          "",
		redisConnString: "",
		fileLoc:         "",
	}
}

func TestInitCli(t *testing.T) {

}
