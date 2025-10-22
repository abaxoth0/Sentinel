package authn

import (
	"fmt"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestCompareHashAndPassword(t *testing.T) {
	// Test data
	password := "testPassword123"
	invalidPassword := "wrongPassword"

	// Generate a valid hash for testing
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to generate test hash: %v", err)
	}

	tests := []struct {
		name     string
		hash     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password comparison",
			hash:     string(hash),
			password: password,
			wantErr:  false,
		},
		{
			name:     "invalid password comparison",
			hash:     string(hash),
			password: invalidPassword,
			wantErr:  true,
		},
		{
			name:     "empty password with valid hash",
			hash:     string(hash),
			password: "",
			wantErr:  true,
		},
		{
			name:     "empty hash with password",
			hash:     "",
			password: password,
			wantErr:  true,
		},
		{
			name:     "both empty strings",
			hash:     "",
			password: "",
			wantErr:  true,
		},
		{
			name:     "malformed hash",
			hash:     "not-a-valid-hash",
			password: password,
			wantErr:  true,
		},
		{
			name:     "hash with wrong format",
			hash:     "$2a$10$",
			password: password,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CompareHashAndPassword(tt.hash, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("CompareHashAndPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("CompareHashAndPassword() unexpected error: %v", err)
			}

			if tt.wantErr && err != InvalidAuthCreditinals {
				t.Errorf("CompareHashAndPassword() expected InvalidAuthCreditinals, got: %v", err)
			}
		})
	}
}

func TestCompareHashAndPasswordWithDifferentCosts(t *testing.T) {
	password := "testPassword123"

	// Test with different bcrypt costs (excluding MaxCost as it's too slow for unit tests)
	// bcrypt.MaxCost (31) takes extremely long to compute and can freeze tests
	costs := []int{bcrypt.MinCost, bcrypt.DefaultCost, 12, 15} // Using 12, 15 as reasonable high costs

	for _, cost := range costs {
		t.Run(fmt.Sprintf("cost_%d", cost), func(t *testing.T) {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
			if err != nil {
				t.Fatalf("Failed to generate hash with cost %d: %v", cost, err)
			}

			// Valid comparison should always work regardless of cost
			authErr := CompareHashAndPassword(string(hash), password)
			if authErr != nil {
				t.Errorf("CompareHashAndPassword() should succeed for correct password with cost %d, got error: %v", cost, authErr)
			}

			// Invalid comparison should always fail regardless of cost
			authErr = CompareHashAndPassword(string(hash), "wrongPassword")
			if authErr == nil {
				t.Errorf("CompareHashAndPassword() should fail for wrong password with cost %d", cost)
			}
		})
	}
}

func TestCompareHashAndPasswordConsistency(t *testing.T) {
	password := "consistencyTestPassword"

	// Generate hash once
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to generate test hash: %v", err)
	}

	// Test multiple comparisons with same inputs should be consistent
	for i := range 10 {
		err := CompareHashAndPassword(string(hash), password)
		if err != nil {
			t.Errorf("CompareHashAndPassword() failed on iteration %d: %v", i, err)
		}
	}

	// Test that wrong password consistently fails
	for i := range 10 {
		err := CompareHashAndPassword(string(hash), "wrongPassword")
		if err == nil {
			t.Errorf("CompareHashAndPassword() should have failed on iteration %d", i)
		}
	}
}

func BenchmarkCompareHashAndPassword(b *testing.B) {
	password := "benchmarkPassword123"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		b.Fatalf("Failed to generate test hash: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CompareHashAndPassword(string(hash), password)
	}
}

func BenchmarkCompareHashAndPasswordInvalid(b *testing.B) {
	password := "benchmarkPassword123"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		b.Fatalf("Failed to generate test hash: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CompareHashAndPassword(string(hash), "wrongPassword")
	}
}
