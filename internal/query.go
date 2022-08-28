package internal

import (
	"context"
	"io"

	"github.com/dstotijn/go-notion"
	"github.com/samber/lo"
)

type Query struct {
	client   *notion.Client
	id       string
	query    *notion.DatabaseQuery
	cur      *string
	finished bool
}

func NewQuery(client *notion.Client, id string, query *notion.DatabaseQuery) *Query {
	return &Query{client: client, id: id, query: query}
}

func (s *Query) Next(ctx context.Context) ([]notion.Page, error) {
	if s.finished {
		return nil, io.EOF
	}
	r, err := s.client.QueryDatabase(ctx, s.id, s.options())
	if err != nil {
		return nil, err
	}
	s.cur = r.NextCursor
	s.finished = !r.HasMore
	return r.Results, nil
}

func (s *Query) options() *notion.DatabaseQuery {
	return &notion.DatabaseQuery{
		Filter:      s.query.Filter,
		Sorts:       s.query.Sorts,
		StartCursor: lo.FromPtr(s.cur),
		PageSize:    s.query.PageSize,
	}
}
