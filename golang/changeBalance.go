package main

import "encoding/json"

type ChangeBalanceStruct struct {
	AccountID     int    `json:"accountID"`
	Money         int    `json:"money"`
	TypeOperation string `json:"type"`

	encoded []byte
	err     error
}

func (cbs *ChangeBalanceStruct) ensureEncoded() {
	if cbs.encoded == nil && cbs.err == nil {
		cbs.encoded, cbs.err = json.Marshal(cbs)
	}
}

func (cbs *ChangeBalanceStruct) Length() int {
	cbs.ensureEncoded()
	return len(cbs.encoded)
}

func (cbs *ChangeBalanceStruct) Encode() ([]byte, error) {
	return cbs.encoded, cbs.err
}
