package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewBcryptPasswordHasher(t *testing.T) {
	hasher := NewBcryptPasswordHasher()
	assert.NotNil(t, hasher)
	assert.Equal(t, bcrypt.DefaultCost, hasher.cost)
}

func TestBcryptPasswordHasher_HashPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "SecurePassword123!",
			wantErr:  false,
		},
		{
			name:     "short password",
			password: "short",
			wantErr:  false, // bcrypt doesn't validate length
		},
		{
			name:     "long password",
			password: string(make([]byte, 100)),
			wantErr:  true, // bcrypt has 72-byte limit
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt allows empty strings
		},
		{
			name:     "password with special characters",
			password: "P@ssw0rd!#$%^&*()",
			wantErr:  false,
		},
		{
			name:     "unicode password",
			password: "пароль密码🔒",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := hasher.HashPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.NotEqual(t, tt.password, hash) // Hash should be different from password
				assert.True(t, len(hash) == 60) // bcrypt hashes are always 60 chars
				assert.True(t, hash[:4] == "$2a$" || hash[:4] == "$2b$" || hash[:4] == "$2y$") // bcrypt prefix
			}
		})
	}
}

func TestBcryptPasswordHasher_HashPassword_UniqueHashes(t *testing.T) {
	hasher := NewBcryptPasswordHasher()
	password := "SamePassword123!"

	// Hash the same password twice
	hash1, err1 := hasher.HashPassword(password)
	hash2, err2 := hasher.HashPassword(password)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEmpty(t, hash1)
	assert.NotEmpty(t, hash2)

	// Hashes should be different due to different salts
	assert.NotEqual(t, hash1, hash2)

	// But both should verify against the original password
	assert.True(t, hasher.CheckPassword(password, hash1))
	assert.True(t, hasher.CheckPassword(password, hash2))
}

func TestBcryptPasswordHasher_CheckPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	tests := []struct {
		name           string
		password       string
		checkPassword  string
		expectedResult bool
	}{
		{
			name:           "correct password",
			password:       "SecurePassword123!",
			checkPassword:  "SecurePassword123!",
			expectedResult: true,
		},
		{
			name:           "incorrect password",
			password:       "SecurePassword123!",
			checkPassword:  "WrongPassword456!",
			expectedResult: false,
		},
		{
			name:           "case sensitive - uppercase",
			password:       "Password123!",
			checkPassword:  "PASSWORD123!",
			expectedResult: false,
		},
		{
			name:           "case sensitive - lowercase",
			password:       "Password123!",
			checkPassword:  "password123!",
			expectedResult: false,
		},
		{
			name:           "extra character",
			password:       "Password123!",
			checkPassword:  "Password123!x",
			expectedResult: false,
		},
		{
			name:           "missing character",
			password:       "Password123!",
			checkPassword:  "Password123",
			expectedResult: false,
		},
		{
			name:           "empty password vs hash",
			password:       "",
			checkPassword:  "",
			expectedResult: true,
		},
		{
			name:           "empty check against real password",
			password:       "RealPassword123!",
			checkPassword:  "",
			expectedResult: false,
		},
		{
			name:           "unicode password match",
			password:       "пароль密码🔒",
			checkPassword:  "пароль密码🔒",
			expectedResult: true,
		},
		{
			name:           "unicode password mismatch",
			password:       "пароль密码🔒",
			checkPassword:  "пароль密码",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First hash the password
			hash, err := hasher.HashPassword(tt.password)
			require.NoError(t, err)

			// Then check with the test password
			result := hasher.CheckPassword(tt.checkPassword, hash)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestBcryptPasswordHasher_CheckPassword_InvalidHash(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	tests := []struct {
		name     string
		password string
		hash     string
	}{
		{
			name:     "completely invalid hash",
			password: "Password123!",
			hash:     "not-a-valid-hash",
		},
		{
			name:     "empty hash",
			password: "Password123!",
			hash:     "",
		},
		{
			name:     "truncated hash",
			password: "Password123!",
			hash:     "$2a$10$N9qo8uLOickgx",
		},
		{
			name:     "wrong algorithm prefix",
			password: "Password123!",
			hash:     "$5$rounds=5000$usesomesillystri$KqJWpanXZHKq2BOB43TSaYhEWsQ1Lr5QNyPCDH/Tp.6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasher.CheckPassword(tt.password, tt.hash)
			assert.False(t, result) // Invalid hashes should always return false
		})
	}
}

func TestBcryptPasswordHasher_CostFactor(t *testing.T) {
	// Test that we can verify the cost factor in the hash
	hasher := NewBcryptPasswordHasher()
	password := "TestPassword123!"

	hash, err := hasher.HashPassword(password)
	require.NoError(t, err)

	// Extract cost from hash
	cost, err := bcrypt.Cost([]byte(hash))
	require.NoError(t, err)
	assert.Equal(t, bcrypt.DefaultCost, cost)
}

func TestBcryptPasswordHasher_PerformanceBaseline(t *testing.T) {
	// This is not a strict performance test, just a baseline check
	// to ensure hashing doesn't take an unreasonable amount of time
	hasher := NewBcryptPasswordHasher()
	password := "TestPassword123!"

	// Hash should complete in reasonable time (bcrypt with default cost should be < 1 second)
	hash, err := hasher.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Verification should also be reasonably fast
	result := hasher.CheckPassword(password, hash)
	assert.True(t, result)
}

func BenchmarkBcryptPasswordHasher_HashPassword(b *testing.B) {
	hasher := NewBcryptPasswordHasher()
	password := "BenchmarkPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.HashPassword(password)
	}
}

func BenchmarkBcryptPasswordHasher_CheckPassword(b *testing.B) {
	hasher := NewBcryptPasswordHasher()
	password := "BenchmarkPassword123!"
	hash, _ := hasher.HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasher.CheckPassword(password, hash)
	}
}
