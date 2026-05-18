package surfing_test

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/surfing"
)

func ExampleURLGenerator() {
	tmlp := "https://example.com/{{vary}}/path?page={{page}}"
	params := []surfing.TemplateContext{
		{map[string]string{"vary": "some", "extra": "x"}, "some-path"},
		{map[string]string{"vary": "another", "extra": "x"}, "another-path"},
	}

	gen, err := surfing.NewURLGenerator(tmlp)
	if err != nil {
		return
	}

	for _, brParam := range params {
		brGen := gen.Branch(brParam)
		for i := range 3 {
			url, err := brGen.Page(i + 1)
			if err != nil {
				return
			}
			fmt.Println(url)
		}
	}

	// Output:
	// https://example.com/some/path?page=1
	// https://example.com/some/path?page=2
	// https://example.com/some/path?page=3
	// https://example.com/another/path?page=1
	// https://example.com/another/path?page=2
	// https://example.com/another/path?page=3
}

func TestParams_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name      string
		params    surfing.Params
		wantError bool
	}{
		{
			name: "valid_params",
			params: surfing.Params{
				ID:          1,
				Name:        "test",
				URLTemplate: "https://example.com/{{page}}",
				MaxPages:    5,
			},
			wantError: false,
		},
		{
			name: "empty_name",
			params: surfing.Params{
				Name:        "",
				URLTemplate: "https://example.com",
				MaxPages:    1,
			},
			wantError: true,
		},
		{
			name: "empty_URLTemplate",
			params: surfing.Params{
				Name:        "test",
				URLTemplate: "",
				MaxPages:    1,
			},
			wantError: true,
		},
		{
			name: "max_pages_is_zero",
			params: surfing.Params{
				Name:        "test",
				URLTemplate: "https://example.com",
				MaxPages:    0,
			},
			wantError: true,
		},
		{
			name: "max_pages_is_negative",
			params: surfing.Params{
				Name:        "test",
				URLTemplate: "https://example.com",
				MaxPages:    -1,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(&tt.params)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
