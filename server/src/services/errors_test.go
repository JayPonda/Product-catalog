package services_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/JayPonda/Product-catalog/server/src/services"
	"github.com/lib/pq"
)

func pqViolation(code, constraint string) *pq.Error {
	return &pq.Error{Code: pq.ErrorCode(code), Constraint: constraint}
}

func TestIsUniqueViolation_Wrappers(t *testing.T) {
	cases := []struct {
		name  string
		err   error
		check func(error) bool
		want  bool
	}{
		{
			name:  "duplicate product name",
			err:   pqViolation("23505", "uq_products_name_active"),
			check: services.IsDuplicateProductName,
			want:  true,
		},
		{
			name:  "duplicate product name wrapped",
			err:   fmt.Errorf("insert product: %w", pqViolation("23505", "uq_products_name_active")),
			check: services.IsDuplicateProductName,
			want:  true,
		},
		{
			name:  "product check against wrong constraint",
			err:   pqViolation("23505", "uq_categories_name_active"),
			check: services.IsDuplicateProductName,
			want:  false,
		},
		{
			name:  "duplicate category name",
			err:   pqViolation("23505", "uq_categories_name_active"),
			check: services.IsDuplicateCategoryName,
			want:  true,
		},
		{
			name:  "category check against wrong constraint",
			err:   pqViolation("23505", "uq_users_email_active"),
			check: services.IsDuplicateCategoryName,
			want:  false,
		},
		{
			name:  "duplicate email",
			err:   pqViolation("23505", "uq_users_email_active"),
			check: services.IsDuplicateEmail,
			want:  true,
		},
		{
			name:  "email check against wrong constraint",
			err:   pqViolation("23505", "uq_products_name_active"),
			check: services.IsDuplicateEmail,
			want:  false,
		},
		{
			name:  "foreign key violation is not a duplicate",
			err:   pqViolation("23503", "uq_products_name_active"),
			check: services.IsDuplicateProductName,
			want:  false,
		},
		{
			name:  "non-postgres error",
			err:   errors.New("some driver failure"),
			check: services.IsDuplicateProductName,
			want:  false,
		},
		{
			name:  "nil error",
			err:   nil,
			check: services.IsDuplicateEmail,
			want:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.check(tc.err); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
