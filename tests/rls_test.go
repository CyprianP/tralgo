package tests

import (
	"context"
	"testing"

	"tralgo/config"
	"tralgo/tenantized"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestProviderBCannotAccessProviderACourse(t *testing.T) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, config.DatabaseUrl)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	// Course 1 belongs to provider 1
	// Provider 2 cannot be able to read it
	err = tenantized.WithTenant(ctx, pool, 2, func(tx pgx.Tx) error {
		var name string
		return tx.QueryRow(ctx,
			"select course_name from courses where course_id = 1",
		).Scan(&name)
	})

	if err != pgx.ErrNoRows {
		t.Fatalf("provider 2 must not see provider 1's course; expected ErrNoRows, got: %v", err)
	}
}

func TestProviderCannotInsertCourseForAnotherProvider(t *testing.T) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, config.DatabaseUrl)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	// Tenant context is provider 1, we try to insert a row owned by provider 2.
	// RLS check must reject it.
	err = tenantized.WithTenant(ctx, pool, 1, func(tx pgx.Tx) error {
		_, e := tx.Exec(ctx,
			`insert into courses (course_name, course_description, provider_id)
			 values ($1, $2, $3)`,
			"should not", "be accepted", 2,
		)
		return e
	})

	if err == nil {
		t.Fatal("expected RLS to reject inserting a course for another provider, but insert succeeded")
	}
}
