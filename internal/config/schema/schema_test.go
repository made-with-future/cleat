package schema

import "testing"

func TestServiceConfig_IsDocker(t *testing.T) {
	trueVal := true
	falseVal := false
	
	tests := []struct {
		name string
		svc  *ServiceConfig
		want bool
	}{
		{"nil", nil, false},
		{"nil Docker", &ServiceConfig{Docker: nil}, false},
		{"true Docker", &ServiceConfig{Docker: &trueVal}, true},
		{"false Docker", &ServiceConfig{Docker: &falseVal}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.svc.IsDocker(); got != tt.want {
				t.Errorf("ServiceConfig.IsDocker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPythonConfig_IsEnabled(t *testing.T) {
	trueVal := true
	falseVal := false
	
	tests := []struct {
		name string
		p    *PythonConfig
		want bool
	}{
		{"nil", nil, false},
		{"nil Enabled", &PythonConfig{Enabled: nil}, true},
		{"true Enabled", &PythonConfig{Enabled: &trueVal}, true},
		{"false Enabled", &PythonConfig{Enabled: &falseVal}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.IsEnabled(); got != tt.want {
				t.Errorf("PythonConfig.IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNpmConfig_IsEnabled(t *testing.T) {
	trueVal := true
	falseVal := false
	
	tests := []struct {
		name string
		n    *NpmConfig
		want bool
	}{
		{"nil", nil, false},
		{"nil Enabled", &NpmConfig{Enabled: nil}, true},
		{"true Enabled", &NpmConfig{Enabled: &trueVal}, true},
		{"false Enabled", &NpmConfig{Enabled: &falseVal}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.IsEnabled(); got != tt.want {
				t.Errorf("NpmConfig.IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
