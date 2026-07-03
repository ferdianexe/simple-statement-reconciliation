package reconcile

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewUsecase(t *testing.T) {
	type args struct {
		bank        bankServiceManager
		transaction transactionServiceManager
	}
	tests := []struct {
		name string
		args func(ctrl *gomock.Controller) args
		want func(a args) *Usecase
	}{
		{
			name: "success",
			args: func(ctrl *gomock.Controller) args {
				return args{
					bank:        NewMockbankServiceManager(ctrl),
					transaction: NewMocktransactionServiceManager(ctrl),
				}
			},
			want: func(a args) *Usecase {
				return &Usecase{bank: a.bank, transaction: a.transaction}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			a := test.args(ctrl)

			got := NewUsecase(a.bank, a.transaction)

			assert.NotNil(t, got)
			assert.Equal(t, test.want(a), got)
		})
	}
}
