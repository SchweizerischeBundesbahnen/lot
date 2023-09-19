package lot_client

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type Client interface {
	client.Client
	Apply(ctx context.Context, obj client.Object, applyPatch interface{}, fieldsOwner string) error
}

type lotClient struct {
	client.Client
}

func New(cl client.Client) Client {
	return lotClient{cl}
}
func (c lotClient) Apply(ctx context.Context, obj client.Object, applyPatch interface{}, fieldsOwner string) error {
	gvk, err := apiutil.GVKForObject(obj, c.Scheme())
	if err != nil {
		return err
	}

	obj.GetObjectKind().SetGroupVersionKind(gvk)
	err = c.Client.Patch(ctx, obj, client.Apply, client.ForceOwnership, client.FieldOwner(fieldsOwner))
	if err != nil {
		return err
	}

	return nil
}
