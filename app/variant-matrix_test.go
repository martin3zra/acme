package app

import (
	"testing"
)

// TestBuildVariantSignature verifies that variant signatures are built consistently
func TestBuildVariantSignature(t *testing.T) {
	tests := []struct {
		name      string
		selection map[int]int
		expected  string
	}{
		{
			name:      "empty selection",
			selection: map[int]int{},
			expected:  "",
		},
		{
			name:      "single attribute",
			selection: map[int]int{1: 10},
			expected:  "1:10",
		},
		{
			name:      "multiple attributes sorted",
			selection: map[int]int{3: 30, 1: 10, 2: 20},
			expected:  "1:10|2:20|3:30",
		},
		{
			name:      "attributes must be sorted by ID",
			selection: map[int]int{5: 50, 2: 20, 8: 80},
			expected:  "2:20|5:50|8:80",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildVariantSignature(tt.selection)
			if result != tt.expected {
				t.Errorf("buildVariantSignature() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestBuildVariantSignatureConsistency ensures the same input always produces the same output
func TestBuildVariantSignatureConsistency(t *testing.T) {
	selection := map[int]int{3: 30, 1: 10, 2: 20, 5: 50}

	// Build signature multiple times
	sig1 := buildVariantSignature(selection)
	sig2 := buildVariantSignature(selection)
	sig3 := buildVariantSignature(selection)

	if sig1 != sig2 || sig2 != sig3 {
		t.Errorf("buildVariantSignature() produced inconsistent results: %v, %v, %v", sig1, sig2, sig3)
	}

	// Verify the signature is in sorted order
	expected := "1:10|2:20|3:30|5:50"
	if sig1 != expected {
		t.Errorf("buildVariantSignature() = %v, want %v", sig1, expected)
	}
}

// TestVariantSignatureOrdering verifies that attribute order is preserved
func TestVariantSignatureOrdering(t *testing.T) {
	// Different input order, same attribute:value pairs
	selection1 := map[int]int{1: 10, 2: 20, 3: 30}
	selection2 := map[int]int{3: 30, 1: 10, 2: 20}
	selection3 := map[int]int{2: 20, 3: 30, 1: 10}

	sig1 := buildVariantSignature(selection1)
	sig2 := buildVariantSignature(selection2)
	sig3 := buildVariantSignature(selection3)

	// All should produce the same signature
	if sig1 != sig2 || sig2 != sig3 {
		t.Errorf("Signature order mismatch: %v, %v, %v", sig1, sig2, sig3)
	}
}

// TestVariantMatrixGeneration verifies cartesian product generation
// This test documents the expected behavior of the frontend matrix generation
func TestVariantMatrixGeneration(t *testing.T) {
	// Example: Color (Red=1, Blue=2) x Size (Small=10, Medium=11)
	// Should generate 4 combinations:
	// - Red/Small: 1:1|2:10
	// - Red/Medium: 1:1|2:11
	// - Blue/Small: 1:2|2:10
	// - Blue/Medium: 1:2|2:11

	colorAttr := 1
	sizeAttr := 2

	colors := []int{1, 2}  // Red, Blue
	sizes := []int{10, 11} // Small, Medium

	expectedSignatures := []string{
		"1:1|2:10",  // Red/Small
		"1:1|2:11",  // Red/Medium
		"1:2|2:10",  // Blue/Small
		"1:2|2:11",  // Blue/Medium
	}

	var generatedSignatures []string
	for _, colorValue := range colors {
		for _, sizeValue := range sizes {
			selection := map[int]int{
				colorAttr: colorValue,
				sizeAttr:  sizeValue,
			}
			sig := buildVariantSignature(selection)
			generatedSignatures = append(generatedSignatures, sig)
		}
	}

	if len(generatedSignatures) != len(expectedSignatures) {
		t.Fatalf("Expected %d signatures, got %d", len(expectedSignatures), len(generatedSignatures))
	}

	// Verify all expected signatures are present
	sigMap := make(map[string]bool)
	for _, sig := range generatedSignatures {
		sigMap[sig] = true
	}

	for _, expected := range expectedSignatures {
		if !sigMap[expected] {
			t.Errorf("Expected signature %v not found in generated signatures", expected)
		}
	}
}
