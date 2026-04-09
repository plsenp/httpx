package stringx

import (
	"reflect"
	"testing"
)

type Test[T Convertable] struct {
	name    string
	args    string
	want    T
	wantErr bool
}

type TestSlice[T Convertable] struct {
	name    string
	args    []string
	want    []T
	wantErr bool
}

type MyInt int

func TestTo_UndelyingInt(t *testing.T) {
	tests := []Test[MyInt]{
		{
			name:    "int",
			args:    "123",
			want:    MyInt(123),
			wantErr: false,
		},
		{
			name:    "int",
			args:    "-123",
			want:    MyInt(-123),
			wantErr: false,
		},
		{
			name:    "int",
			args:    "1232sdf",
			want:    MyInt(0),
			wantErr: true,
		},
		{
			name:    "int",
			args:    "12345678900000000000000",
			want:    MyInt(0),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := To[MyInt](tt.args)
			t.Logf("To() = %v, wantErr %v", got, err)
			if (err != nil) != tt.wantErr {
				t.Errorf("To() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("To() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTo_int64(t *testing.T) {
	tests := []Test[int]{
		{
			name:    "int",
			args:    "123",
			want:    123,
			wantErr: false,
		},
		{
			name:    "int",
			args:    "12345678900000000000000",
			want:    0,
			wantErr: true,
		},
		{
			name:    "int",
			args:    "-123",
			want:    -123,
			wantErr: false,
		},
		{
			name:    "int",
			args:    "1232sdf",
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := To[int](tt.args)
			t.Logf("To() = %v, wantErr %v", got, err)
			if (err != nil) != tt.wantErr {
				t.Errorf("To() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("To() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTo_int32(t *testing.T) {
	tests := []Test[int32]{
		{
			name:    "int",
			args:    "123",
			want:    123,
			wantErr: false,
		},
		{
			name:    "int",
			args:    "12345678900000000000000",
			want:    0,
			wantErr: true,
		},
		{
			name:    "int",
			args:    "-123",
			want:    -123,
			wantErr: false,
		},
		{
			name:    "int",
			args:    "1232sdf",
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := To[int32](tt.args)
			t.Logf("To() = %v, wantErr %v", got, err)
			if (err != nil) != tt.wantErr {
				t.Errorf("To() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("To() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTo_uint(t *testing.T) {
	tests := []Test[uint]{
		{
			name:    "int",
			args:    "123",
			want:    123,
			wantErr: false,
		},
		{
			name:    "int",
			args:    "12345678900000000000000",
			want:    0,
			wantErr: true,
		},
		{
			name:    "int",
			args:    "-123",
			want:    0,
			wantErr: true,
		},
		{
			name:    "int",
			args:    "1232sdf",
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := To[uint](tt.args)
			t.Logf("To() = %v, wantErr %v", got, err)
			if (err != nil) != tt.wantErr {
				t.Errorf("To() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("To() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTo_float64(t *testing.T) {
	tests := []Test[float64]{
		{
			name:    "float",
			args:    "123.456",
			want:    123.456,
			wantErr: false,
		},
		{
			name:    "float",
			args:    "123.456sdf",
			want:    0,
			wantErr: true,
		},
		{
			name:    "float",
			args:    "-123.456",
			want:    -123.456,
			wantErr: false,
		},
		{
			name:    "float",
			args:    "1232sdf",
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := To[float64](tt.args)
			t.Logf("To() = %v, wantErr %v", got, err)
			if (err != nil) != tt.wantErr {
				t.Errorf("To() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("To() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToSlice_Uint(t *testing.T) {
	tests := []TestSlice[uint]{
		{
			name:    "slice_int",
			args:    []string{"123", "456"},
			want:    []uint{123, 456},
			wantErr: false,
		},
		{
			name:    "slice_int",
			args:    []string{"-123"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "slice_int",
			args:    []string{"123.456"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "slice_int",
			args:    []string{"1232sdf"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToSlice[uint](tt.args)
			t.Logf("ToSlice() = %v, wantErr %v", got, err)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
