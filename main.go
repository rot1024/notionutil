package main

import (
	"context"
	"fmt"

	"github.com/dstotijn/go-notion"
	"github.com/rot1024/notionutil/internal"
)

func main() {
	c, err := internal.LoadConfig("")
	if err != nil {
		fmt.Printf("failed to load config: %s", err)
		return
	}

	ctx := context.Background()
	client := notion.NewClient(c.APIKey)

	for i, op := range c.Ops {
		if op.Op == nil {
			continue
		}

		fmt.Printf("op %d %s start\n", i, op.Type)
		if err := op.Op.Run(ctx, client); err != nil {
			fmt.Printf("failed to run op %d: %s\n", i, err)
		}
		fmt.Printf("op %d %s finished\n", i, op.Type)
	}

	println("done")
}
