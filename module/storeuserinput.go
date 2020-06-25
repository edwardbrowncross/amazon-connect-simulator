package module

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type storeUserInput flow.Module

type storeUserInputParams struct {
	Text             string
	Timeout          string
	MaxDigits        int
	TextToSpeechType string
	EncryptEntry     bool
	TerminatorDigits *string
	EncryptionKeyId  *string
	EncryptionKey    *string
}

func (m storeUserInput) Run(call CallConnector) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleStoreUserInput {
		return nil, fmt.Errorf("module of type %s being run as storeUserInput", m.Type)
	}
	pr := parameterResolver{call}
	p := storeUserInputParams{}
	err = pr.unmarshal(m.Parameters, &p)
	if err != nil {
		return
	}
	timeout, err := strconv.Atoi(p.Timeout)
	if err != nil {
		return
	}
	txt := pr.jsonPath(p.Text)
	call.Send(txt, p.TextToSpeechType == "ssml")
	terminator := '#'
	if p.TerminatorDigits != nil {
		terminator, _ = utf8.DecodeRuneInString(*p.TerminatorDigits)
	}
	entry, ok := call.Receive(p.MaxDigits, time.Duration(timeout)*time.Second, terminator)
	if !ok {
		call.SetSystem(flow.SystemLastUserInput, "Timeout")
		return m.Branches.GetLink(flow.BranchSuccess), nil
	}
	if p.EncryptEntry {
		if p.EncryptionKey == nil {
			return nil, errors.New("missing encryption key")
		}
		if p.EncryptionKeyId == nil {
			return nil, errors.New("missing encryption key ID")
		}
		enc := call.Encrypt(entry, *p.EncryptionKeyId, []byte(*p.EncryptionKey))
		entry = base64.StdEncoding.EncodeToString(enc)
	}
	call.SetSystem(flow.SystemLastUserInput, entry)
	next = m.Branches.GetLink(flow.BranchSuccess)
	return
}
