package main

import (
	"testing"

	"github.com/sirion/fanmi/app/configuration"
)

func Test_calculateStep(t *testing.T) {
	type args struct {
		temp      float32
		lowEntry  configuration.Entry
		highEntry configuration.Entry
	}
	tests := []struct {
		name string
		args args
		want float32
	}{
		{
			name: "Same",
			args: args{
				temp: 50,
				lowEntry: configuration.Entry{
					Temp: 45, Speed: 0.5,
				},
				highEntry: configuration.Entry{
					Temp: 55, Speed: 0.5,
				},
			},
			want: 0.5,
		},
		{
			name: "Middle",
			args: args{
				temp: 50,
				lowEntry: configuration.Entry{
					Temp: 45, Speed: 0.4,
				},
				highEntry: configuration.Entry{
					Temp: 55, Speed: 0.6,
				},
			},
			want: 0.5,
		},
		{
			name: "Min",
			args: args{
				temp: 50,
				lowEntry: configuration.Entry{
					Temp: 0, Speed: 0,
				},
				highEntry: configuration.Entry{
					Temp: 50, Speed: 0.75,
				},
			},
			want: 0.75,
		},
		{
			name: "Max",
			args: args{
				temp: 50,
				lowEntry: configuration.Entry{
					Temp: 50, Speed: 0.1,
				},
				highEntry: configuration.Entry{
					Temp: 80, Speed: 0.75,
				},
			},
			want: 0.1,
		},
		{
			name: "Random",
			args: args{
				temp: 55,
				lowEntry: configuration.Entry{
					Temp: 40, Speed: 0,
				},
				highEntry: configuration.Entry{
					Temp: 60, Speed: 0.2,
				},
			},
			want: 0.15,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateStep(tt.args.temp, tt.args.lowEntry, tt.args.highEntry); got != tt.want {
				t.Errorf("calculateStep() = %v, want %v", got, tt.want)
			}
		})
	}
}
