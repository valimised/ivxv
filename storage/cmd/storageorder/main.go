/*
The storageorder application adds votes to order table in the storage service.
*/
package main

import (
	"encoding/csv"
	"flag"
	"io"
	"os"

	"ivxv.ee/common/collector/command"
	"ivxv.ee/common/collector/command/exit"
	//ivxv:modules common/collector/container
	//ivxv:modules common/collector/q11n
	//ivxv:modules common/collector/storage
)

const usage = `storageorder adds votes to order table in the storage service.

It is usually not necessary to add votes to votes order table,
but use of this application might help in cases where adding vote is
unsuccessful. This might happen when too many votes arrive same time.

Input is csv file in this order:
<name>,<voterid>,<adminCode>,<district>
`

var (
	file = flag.String("file", "", "csv file")
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

	if *file == "" {
		return c.Error(exit.Usage, CmdAddVoteOrderArgError{
			Description: _STORAGEORDER_FILE}, "missing file argument")
	}

	f, err := os.Open(*file)
	if err != nil {
		return c.Error(exit.Usage, CmdAddVoteOrderFileOpenError{Err: err,
			Description: _STORAGEORDER_FILE_OPEN}, "failed to open file")
	}

	defer f.Close()

	csvReader := csv.NewReader(f)
	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return c.Error(exit.Usage, CmdAddVoteOrderLineReadError{Err: err,
				Description: _STORAGEORDER_READ_CSV}, "failed to read file")
		}
		if len(rec) != 4 {
			return c.Error(exit.Usage, CmdAddVoteOrderLineError{
				Description: _STORAGEORDER_CSV}, "wrong number of fields in line")
		}
		if err := c.Storage.AddVoteOrder(c.Ctx, rec[0], rec[1], rec[3], rec[2]); err != nil {
			return c.Error(exit.Unavailable, CmdAddVoteOrderError{Err: err,
				Description: _STORAGEORDER_ORDER},
				"failed to add vote to order table:", err)
		}
	}

	return
}
