package helpers

import (
	"github.com/Benbentwo/utils/util"
	"io"
	"io/ioutil"
)

func PrintBody(response io.ReadCloser, err error) {
	defer response.Close()

	body, err := ioutil.ReadAll(response)
	if err != nil {
		util.Logger().Fatalf("ERROR: %s", err)
	}

	util.Logger().Infof("%s", body)
}
