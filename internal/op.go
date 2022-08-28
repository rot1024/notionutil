package internal

import (
	"context"
	"encoding/json"

	"github.com/dstotijn/go-notion"
)

type Op interface {
	Run(ctx context.Context, client *notion.Client) error
}

type OpJSON struct {
	Type string `json:"type"`
	Op   Op     `json:"-"`
}

func (op *OpJSON) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, op); err != nil {
		return err
	}

	if op.Type == "mergeDate" {
		var m *MergeDateOp
		if err := json.Unmarshal(data, m); err != nil {
			return err
		}
		op.Op = m
	}

	return nil
}
