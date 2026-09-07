package helpers

import "testing"

func TestValidateLuhn(t *testing.T) {
	t.Parallel()

	t.Run("success - valid card number", func(t *testing.T) {
		result := ValidateLuhn("4111111111111111")
		if !result {
			t.Errorf("expected true, got false")
		}
	})

	t.Run("error - invalid card number", func(t *testing.T) {
		result := ValidateLuhn("4111111111111112")
		if result {
			t.Errorf("expected false, got true")
		}
	})

	t.Run("error - empty string", func(t *testing.T) {
		result := ValidateLuhn("")
		if result {
			t.Errorf("expected false, got true")
		}
	})

	t.Run("error - contains non-digit characters", func(t *testing.T) {
		result := ValidateLuhn("4111-1111-1111-1111")
		if result {
			t.Errorf("expected false, got true")
		}
	})
}
