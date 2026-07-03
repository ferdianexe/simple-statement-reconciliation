package csv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBankStatement_AbsAmount(t *testing.T) {
	type args struct {
		amount float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "when_amount_is_negative_then_return_positive_value",
			args: args{amount: -110000},
			want: 110000,
		},
		{
			name: "when_amount_is_positive_then_return_same_value",
			args: args{amount: 220000},
			want: 220000,
		},
		{
			name: "when_amount_is_zero_then_return_zero",
			args: args{amount: 0},
			want: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := BankStatement{Amount: test.args.amount}
			got := b.AbsAmount()
			assert.Equal(t, test.want, got)
		})
	}
}

func TestBankStatement_InferredType(t *testing.T) {
	type args struct {
		amount float64
	}
	tests := []struct {
		name string
		args args
		want TrxType
	}{
		{
			name: "when_amount_is_negative_then_return_debit",
			args: args{amount: -110000},
			want: Debit,
		},
		{
			name: "when_amount_is_positive_then_return_credit",
			args: args{amount: 220000},
			want: Credit,
		},
		{
			name: "when_amount_is_zero_then_return_credit",
			args: args{amount: 0},
			want: Credit,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := BankStatement{Amount: test.args.amount}
			got := b.InferredType()
			assert.Equal(t, test.want, got)
		})
	}
}
