package versionset

import (
	"reflect"
	"testing"
)

func TestVersionSetSort(t *testing.T) {
	vs := VersionSet{
		{Name: "c", Version: "v1.0.0", Optional: false},
		{Name: "a", Version: "v2.0.0", Optional: true},
		{Name: "b", Version: "v3.0.0", Optional: false},
	}

	vs.Sort()

	expected := VersionSet{
		{Name: "a", Version: "v2.0.0", Optional: true},
		{Name: "b", Version: "v3.0.0", Optional: false},
		{Name: "c", Version: "v1.0.0", Optional: false},
	}

	if !reflect.DeepEqual(vs, expected) {
		t.Errorf("Sort() = %v, want %v", vs, expected)
	}
}

func TestVersionSetSet(t *testing.T) {
	vs := VersionSet{}

	vs.Set(VersionInfo{Name: "a", Version: "v1.0.0", Optional: false})
	vs.Set(VersionInfo{Name: "b", Version: "v2.0.0", Optional: true})
	vs.Set(VersionInfo{Name: "a", Version: "v1.1.0", Optional: true}) // Update existing

	expected := VersionSet{
		{Name: "a", Version: "v1.1.0", Optional: true},
		{Name: "b", Version: "v2.0.0", Optional: true},
	}

	if !reflect.DeepEqual(vs, expected) {
		t.Errorf("After Set() = %v, want %v", vs, expected)
	}
}

func TestVersionSetGet(t *testing.T) {
	vs := VersionSet{
		{Name: "a", Version: "v1.0.0", Optional: false},
		{Name: "b", Version: "v2.0.0", Optional: true},
	}

	tests := []struct {
		name     string
		expected VersionInfo
		found    bool
	}{
		{"a", VersionInfo{Name: "a", Version: "v1.0.0", Optional: false}, true},
		{"b", VersionInfo{Name: "b", Version: "v2.0.0", Optional: true}, true},
		{"c", VersionInfo{}, false},
	}

	for _, tt := range tests {
		got, found := vs.Get(tt.name)
		if found != tt.found || !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("Get(%s) = (%v, %v), want (%v, %v)", tt.name, got, found, tt.expected, tt.found)
		}
	}
}

func TestVersionSetCompatibleWith(t *testing.T) {
	tests := []struct {
		name    string
		vs1     VersionSet
		vs2     VersionSet
		want    VersionSet
		wantErr bool
	}{
		{
			name: "Compatible versions",
			vs1: VersionSet{
				{Name: "a", Version: "v1.2.0", Optional: false},
				{Name: "b", Version: "v2.0.0", Optional: true},
			},
			vs2: VersionSet{
				{Name: "a", Version: "v1.1.0", Optional: false},
				{Name: "b", Version: "v2.1.0", Optional: false},
			},
			want: VersionSet{
				{Name: "a", Version: "v1.1.0", Optional: false},
				{Name: "b", Version: "v2.0.0", Optional: false},
			},
			wantErr: false,
		},
		{
			name: "Incompatible major versions",
			vs1: VersionSet{
				{Name: "a", Version: "v1.0.0", Optional: false},
			},
			vs2: VersionSet{
				{Name: "a", Version: "v2.0.0", Optional: false},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Missing required version",
			vs1: VersionSet{
				{Name: "a", Version: "v1.0.0", Optional: false},
			},
			vs2: VersionSet{
				{Name: "b", Version: "v1.0.0", Optional: false},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Extra optional version",
			vs1: VersionSet{
				{Name: "a", Version: "v1.0.0", Optional: false},
				{Name: "b", Version: "v1.0.0", Optional: true},
			},
			vs2: VersionSet{
				{Name: "a", Version: "v1.0.0", Optional: false},
			},
			want: VersionSet{
				{Name: "a", Version: "v1.0.0", Optional: false},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.vs1.CompatibleWith(tt.vs2)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompatibleWith() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CompatibleWith() = %v, want %v", got, tt.want)
			}
		})
	}
}
