package util

import (
	"testing"
)

func TestPkg_Util_Ptr(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		wantVal int
		wantNil bool
	}{
		{"returns pointer to int", 42, 42, false},
		{"returns pointer to zero int", 0, 0, false},
		{"returns pointer to negative int", -10, -10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Ptr(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Errorf("Ptr() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Error("Ptr() returned nil")
					return
				}
				if *got != tt.wantVal {
					t.Errorf("Ptr() = %v, want %v", *got, tt.wantVal)
				}
			}
		})
	}
}

func TestPkg_Util_Ptr_String(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantVal string
		wantNil bool
	}{
		{"returns pointer to string", "hello", "hello", false},
		{"returns pointer to empty string", "", "", false},
		{"returns pointer to long string", "a very long string value", "a very long string value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Ptr(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Errorf("Ptr() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Error("Ptr() returned nil")
					return
				}
				if *got != tt.wantVal {
					t.Errorf("Ptr() = %v, want %v", *got, tt.wantVal)
				}
			}
		})
	}
}

func TestPkg_Util_Ptr_Bool(t *testing.T) {
	tests := []struct {
		name    string
		input   bool
		wantVal bool
		wantNil bool
	}{
		{"returns pointer to true", true, true, false},
		{"returns pointer to false", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Ptr(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Errorf("Ptr() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Error("Ptr() returned nil")
					return
				}
				if *got != tt.wantVal {
					t.Errorf("Ptr() = %v, want %v", *got, tt.wantVal)
				}
			}
		})
	}
}

func TestPkg_Util_Ptr_Struct(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	tests := []struct {
		name    string
		input   Person
		wantNil bool
	}{
		{"returns pointer to struct with data", Person{Name: "John", Age: 30}, false},
		{"returns pointer to zero struct", Person{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Ptr(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Errorf("Ptr() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Error("Ptr() returned nil")
					return
				}
				if *got != tt.input {
					t.Errorf("Ptr() = %v, want %v", *got, tt.input)
				}
			}
		})
	}
}

func TestPkg_Util_Ptr_PointerIndependence(t *testing.T) {
	tests := []struct {
		name        string
		firstInput  int
		secondInput int
	}{
		{"different int values", 10, 20},
		{"same int values", 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p1 := Ptr(tt.firstInput)
			p2 := Ptr(tt.secondInput)

			if p1 == nil || p2 == nil {
				t.Error("Ptr() returned nil for one or both pointers")
				return
			}

			*p1 = 999

			if *p2 != tt.secondInput {
				t.Errorf("Modifying one pointer affected the other: p2 = %v, want %v", *p2, tt.secondInput)
			}
		})
	}
}
