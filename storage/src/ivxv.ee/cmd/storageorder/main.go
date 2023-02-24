/*
The storageorder application adds vote to order table in the storage service.
*/
package main

import (
	"flag"
	"os"

	"ivxv.ee/command"
	"ivxv.ee/command/exit"
	//ivxv:modules container
	//ivxv:modules q11n
	//ivxv:modules storage
)

const usage = `storageorder adds vote to order table in the storage service.

It is usually not necessary to manually add vote to votes order table,
but use of this application might help in cases where adding vote is
unsuccessful. This might happen when too many votes arrive same time.
`

var (
	name      = flag.String("name", "", "voter name")
	voterid   = flag.String("voterid", "", "voter id")
	district  = flag.String("district", "", "district")
	adminCode = flag.String("admincode", "", "admin code")
)

func main() {
	// Call storageordermain in a separate function so that it can set up
	// defers and have them trigger before returning with a non-zero exit
	// code.
	os.Exit(storageordermain())
}

func storageordermain() (code int) {
	c := command.New("ivxv-storageorder", usage)
	defer func() {
		code = c.Cleanup(code)
	}()
	if *name == "" || *voterid == "" || *district == "" || *adminCode == "" {
		return c.Error(exit.Usage, CmdAddVoteOrderArgError{}, "missing vote argument(s)")
	}

	if err := c.Storage.AddVoteOrder(c.Ctx, *name, *voterid, *district, *adminCode); err != nil {
		return c.Error(exit.Unavailable, CmdAddVoteOrderError{Err: err},
			"failed to add vote to order table:", err)
	}

	return
}
