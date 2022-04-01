package mbutil

import (
	"os/exec"
	"fmt"
	"io/ioutil"
	"bytes"
	"os"
)

// executes a commannd with the standand input of bytes
func ExecCmdBytes(cmd []string,bs []byte) error {
	subProcess := exec.Command(cmd[0],cmd[1:]...) //Just for testing, replace with your subProcess
	subProcess.Stdin = bytes.NewReader(bs)
	bs,err := subProcess.Output()
	fmt.Println(string(bs))
	return err 
}


// reads from standand input 
func ReadStdin() ([]byte,error) {
	bs,err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println(err)
	}
	return bs,err
}