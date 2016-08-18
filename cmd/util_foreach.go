package cmd

import (
	"fmt"

	"github.com/rancher/go-rancher/client"
	"github.com/urfave/cli"
)

func printErr(id string, oldErr, newErr error) error {
	if newErr != nil {
		fmt.Printf("error %s: %s\n", id, newErr.Error())
		return newErr
	}
	return oldErr
}

func forEachResourceWithClient(c *client.RancherClient, ctx *cli.Context, types []string, fn func(c *client.RancherClient, resource *client.Resource) (string, error)) error {
	types = getTypesStringFlag(ctx, types)
	w, err := NewWaiter(ctx)
	if err != nil {
		return err
	}

	var lastErr error
	for _, id := range ctx.Args() {
		resource, err := Lookup(c, id, types...)
		if err != nil {
			lastErr = printErr(id, lastErr, err)
			continue
		}

		resourceID, err := fn(c, resource)
		if resourceID == "" {
			resourceID = resource.Id
		}
		lastErr = printErr(resource.Id, lastErr, err)
		if resourceID != "" {
			w.Add(resourceID)
		}
	}

	if lastErr != nil {
		return cli.NewExitError("", 1)
	}

	return w.Wait()
}

func forEachResource(ctx *cli.Context, types []string, fn func(c *client.RancherClient, resource *client.Resource) (string, error)) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	return forEachResourceWithClient(c, ctx, types, fn)
}
