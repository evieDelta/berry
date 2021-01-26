package admin

import (
	"context"

	"github.com/starshine-sys/bcr"
)

func (c *Admin) addCategory(ctx *bcr.Context) (err error) {
	if err = ctx.CheckMinArgs(1); err != nil {
		_, err = ctx.Send("You need to give a category name.", nil)
		return err
	}

	var e bool
	err = c.db.Pool.QueryRow(context.Background(), "select exists (select from categories where lower(name) = lower($1))", ctx.RawArgs).Scan(&e)
	if err != nil {
		return c.db.InternalError(ctx, err)
	}
	if e {
		_, err = ctx.Send(":x :A category with that name already exists.", nil)
		return err
	}

	var id int
	err = c.db.Pool.QueryRow(context.Background(), "insert into public.categories (name) values ($1) returning id", ctx.RawArgs).Scan(&id)
	if err != nil {
		return c.db.InternalError(ctx, err)
	}
	_, err = ctx.Sendf("Added category `%v` with ID %v.", ctx.RawArgs, id)
	return
}
