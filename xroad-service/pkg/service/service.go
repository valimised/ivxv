package service

import (
	"crypto/tls"
	"crypto/x509"
	stderrors "errors"
	"fmt"
	"net/rpc/jsonrpc"
	"os"
	"xroad/pkg/conf"
	"xroad/pkg/errors"
)

type EHSService struct {
	batchMaxSize int
	elections    []conf.EHS
}

type Elections struct {
	Elections []Election `json:"elections"`
}

type Election struct {
	Name string `json:"name"`
}

func NewEHSService(elections []conf.EHS, batchMaxSize int) EHSService {
	return EHSService{batchMaxSize: batchMaxSize, elections: elections}
}

func (e EHSService) GetElections() Elections {
	elections := []Election{}
	for _, election := range e.elections {
		elections = append(elections, Election{Name: election.Name})
	}
	return Elections{elections}
}

type ElectionSeqNo struct {
	Name  string `json:"electionName"`
	SeqNo int    `json:"lastSeqNo"`
}

type SeqNo struct {
	SeqNo     int
	SessionID string
}

func (e EHSService) GetSeqNo(electionName string) (ElectionSeqNo, error) {
	var ehs conf.EHS
	var ok bool
	if ehs, ok = e.isCorrectName(electionName); !ok {
		return ElectionSeqNo{}, errors.FieldError{
			Code:  errors.ErrNotFound.Error(),
			Field: "electionId",
			Value: electionName}.ToErr()
	}
	conn, err := e.ehsConn(ehs.Address, ehs.RootCA, ehs.ClientCert, ehs.ClientKey, ehs.ServerName)
	if err != nil {
		return ElectionSeqNo{}, err
	}
	defer conn.Close()
	client := jsonrpc.NewClient(conn)

	var resp SeqNo
	err = client.Call("RPC.VotesSeqNo", nil, &resp)
	if err != nil {
		if stderrors.Is(err, errors.ErrVotingEnd) {
			return ElectionSeqNo{}, err
		}
		return ElectionSeqNo{}, fmt.Errorf("EHS: %s", err)
	}
	return ElectionSeqNo{electionName, resp.SeqNo}, nil
}

type ElectionBatch struct {
	Name          string       `json:"electionName"`
	SeqNo         int          `json:"fromSeqNo"`
	BatchMaxSize  int          `json:"batchMaxSize"`
	EVotingsBatch BatchRecords `json:"eVotingsBatch"`
}
type BatchRecords struct {
	BatchRecords []Batch `json:"batchRecords"`
}

type Batch struct {
	SeqNo               int    `json:"seqNo"`
	IdCode              string `json:"idCode"`
	VoterName           string `json:"voterName"`
	KovCode             string `json:"kovCode"`
	ElectoralDistrictNo int    `json:"electoralDistrictNo"`
}

type VotesArgs struct {
	VotesFrom    int
	BatchMaxSize int
}

func (e EHSService) GetBatch(electionName string, seqNo int) (ElectionBatch, error) {
	var ehs conf.EHS
	var ok bool
	if ehs, ok = e.isCorrectName(electionName); !ok {
		return ElectionBatch{}, errors.FieldError{
			Code:  errors.ErrNotFound.Error(),
			Field: "electionId",
			Value: electionName}.ToErr()
	}
	conn, err := e.ehsConn(ehs.Address, ehs.RootCA, ehs.ClientCert, ehs.ClientKey, ehs.ServerName)
	if err != nil {
		return ElectionBatch{}, err
	}
	defer conn.Close()
	client := jsonrpc.NewClient(conn)

	var resp BatchRecords
	err = client.Call("RPC.Votes", VotesArgs{VotesFrom: seqNo, BatchMaxSize: e.batchMaxSize}, &resp)
	if err != nil {
		switch {
		case stderrors.Is(err, errors.ErrBadRequest):
			err = errors.FieldError{
				Code:  errors.ErrNotFound.Error(),
				Field: "fromSeqNo",
				Value: seqNo}.ToErr()
			return ElectionBatch{}, err
		case stderrors.Is(err, errors.ErrVotingEnd):
			return ElectionBatch{}, err
		default:
			return ElectionBatch{}, fmt.Errorf("EHS: %s", err)
		}
	}

	return ElectionBatch{Name: electionName, SeqNo: seqNo, BatchMaxSize: e.batchMaxSize, EVotingsBatch: resp}, nil
}

func (e EHSService) isCorrectName(electionName string) (conf.EHS, bool) {
	for _, election := range e.elections {
		if election.Name == electionName {
			return election, true
		}
	}
	return conf.EHS{}, false
}

func (e EHSService) ehsConn(addr string, certLoc string, clientcert string, clientkey string, serverName string) (*tls.Conn, error) {
	cert, err := os.ReadFile(certLoc)
	if err != nil {
		return nil, fmt.Errorf("EHS cert %s: ", err)
	}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(cert); !ok {
		return nil, fmt.Errorf("unable to parse EHS cert %s", cert)
	}
	clientCert, err := tls.LoadX509KeyPair(clientcert, clientkey)
	if err != nil {
		return nil, fmt.Errorf("EHS client cert %s: ", err)
	}
	conn, err := tls.Dial("tcp", addr, &tls.Config{RootCAs: certPool,
		Certificates: []tls.Certificate{clientCert}, ServerName: serverName})

	if err != nil {
		return nil, fmt.Errorf("EHS connection error: %s", err)
	}
	return conn, nil
}
