package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/dstotijn/go-notion"
	"github.com/samber/lo"
	"github.com/thlib/go-timezone-local/tzlocal"
)

type MergeDateOp struct {
	DatabaseID      string `json:"databaseId,omitempty"`
	StartDatePropID string `json:"startDatePropId,omitempty"`
	StartTimePropID string `json:"startTImePropId,omitempty"`
	EndDatePropID   string `json:"endDatePropId,omitempty"`
	EndTimePropID   string `json:"endTimePropId,omitempty"`
	ToPropID        string `json:"toPropId,omitempty"`
	DryRun          bool   `json:"-"`
	Quiet           bool   `json:"-"`
	s               *Query `json:"-"`
	tzname          string `json:"-"`
}

func (op *MergeDateOp) Run(ctx context.Context, client *notion.Client) error {
	for {
		pages, err := op.find(ctx, client)
		if err != nil {
			if errors.Is(io.EOF, err) {
				break
			}
			return err
		}

		if !op.Quiet {
			fmt.Printf("%d pages found\n", len(pages))
		}

		for i, p := range pages {
			if params, err := op.update(&p); err != nil {
				return fmt.Errorf("failed to update page for id %s: %w", p.ID, err)
			} else if params != nil {
				if op.DryRun {
					fmt.Printf("update page %s: %+v\n", p.ID, params.DatabasePageProperties[op.ToPropID].Date)
				} else {
					if _, err := client.UpdatePage(ctx, p.ID, *params); err != nil {
						return err
					}

					fmt.Printf("updated %d/%d\n", i+1, len(pages))
				}
			}
		}
	}

	return nil
}

func (op *MergeDateOp) find(ctx context.Context, client *notion.Client) ([]notion.Page, error) {
	if op.s == nil {
		op.s = NewQuery(client, op.DatabaseID, &notion.DatabaseQuery{
			Filter: &notion.DatabaseQueryFilter{
				And: []notion.DatabaseQueryFilter{
					{
						Property: op.ToPropID,
						Date: &notion.DateDatabaseQueryFilter{
							IsEmpty: true,
						},
					},
					{
						Property: op.StartDatePropID,
						Date: &notion.DateDatabaseQueryFilter{
							IsNotEmpty: true,
						},
					},
				},
			},
		})
	}
	return op.s.Next(ctx)
}

func date(datep notion.DatabasePageProperty, timep notion.DatabasePageProperty) (t *notion.DateTime, err error) {
	if datep.Date == nil {
		return
	}

	datet := datep.Date.Start
	timestr := lo.Reduce(timep.RichText, func(s string, t notion.RichText, _ int) string {
		return s + t.PlainText
	}, "")

	if timestr != "" {
		datestr := datet.Format("2006-01-02")
		res, err2 := time.Parse("2006-01-02 15:04:05", datestr+" "+timestr)
		if err2 != nil {
			err = fmt.Errorf("failed to paese time %s %s: %w", datestr, timestr, err2)
			return
		}
		return lo.ToPtr(notion.NewDateTime(res, true)), nil
	}

	return &datet, nil
}

func (op *MergeDateOp) dateFrom(start *notion.DateTime, end *notion.DateTime) (t *notion.Date, _ error) {
	if start == nil {
		return nil, errors.New("start time should not be null")
	}

	if op.tzname == "" {
		tzname, err := tzlocal.RuntimeTZ()
		if err != nil {
			return nil, fmt.Errorf("failed to load timezone: %w", err)
		}
		op.tzname = tzname
	}

	return &notion.Date{
		Start:    *start,
		End:      end,
		TimeZone: &op.tzname,
	}, nil
}

func (op *MergeDateOp) update(p *notion.Page) (*notion.UpdatePageParams, error) {
	if p == nil {
		return nil, nil
	}

	properties, ok := p.Properties.(notion.DatabasePageProperties)
	if !ok {
		return nil, errors.New("invalid properties type")
	}

	sd, ok := properties[op.StartDatePropID]
	if !ok {
		return nil, errors.New("date prop is not found")
	}

	st, ok := properties[op.StartTimePropID]
	if !ok {
		return nil, errors.New("time prop is not found")
	}

	ed, ok := properties[op.EndDatePropID]
	if !ok {
		return nil, errors.New("date prop is not found")
	}

	et, ok := properties[op.EndTimePropID]
	if !ok {
		return nil, errors.New("time prop is not found")
	}

	start, err := date(sd, st)
	if err != nil {
		return nil, fmt.Errorf("failed to create a start time: %w", err)
	}

	end, err := date(ed, et)
	if err != nil {
		return nil, fmt.Errorf("failed to create a end time: %w", err)
	}

	date, err := op.dateFrom(start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to create a date: %w", err)
	}

	return &notion.UpdatePageParams{
		DatabasePageProperties: notion.DatabasePageProperties{
			op.ToPropID: notion.DatabasePageProperty{
				Date: date,
			},
		},
	}, nil
}
