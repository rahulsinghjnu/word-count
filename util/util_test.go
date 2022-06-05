package util

import (
	"reflect"
	"testing"
)

func TestRankByWordCount(t *testing.T) {
	type args struct {
		wordFrequencies map[string]int
	}
	tests := []struct {
		name string
		args args
		want PairList
	}{
		{"Three Elements", args{wordFrequencies: map[string]int{"A": 1, "B": 5, "C": 3}}, []Pair{{Key: "B", Value: 5}, {Key: "C", Value: 3}, {Key: "A", Value: 1}}},
		{"One Elements", args{wordFrequencies: map[string]int{"A": 1}}, []Pair{{Key: "A", Value: 1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RankByWordCount(tt.args.wordFrequencies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RankByWordCount() = %v, want %v", got, tt.want)
			}
		})
	}
}
