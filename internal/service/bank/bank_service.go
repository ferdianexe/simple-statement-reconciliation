package bank

import (
	"context"
	"fmt"
	"strings"
)

const defaultPath = "testdata/"

func (svc *Service) GetBankStatementHistory(ctx context.Context, request []BankStatementParams) ([]BankStatement, error) {
	var bankStmts []BankStatement
	for _, b := range request {
		path := b.Path
		if path == "" {
			path = defaultPath + fmt.Sprintf("bank_%s.csv", strings.ToLower(b.BankName))
		}
		stmts, err := svc.resource.ParseBankStatementFromCSV(ctx, path, b.BankName, b.Start, b.End)
		if err != nil {
			return []BankStatement{}, fmt.Errorf("parsing bank statement %s: %w", path, err)
		}

		svcStmts := make([]BankStatement, 0, len(stmts))
		for _, s := range stmts {
			svcStmts = append(svcStmts, BankStatement{
				UniqueID: s.UniqueID,
				Amount:   s.Amount,
				Date:     s.Date,
				Bank:     s.Bank,
			})
		}
		bankStmts = append(bankStmts, svcStmts...)
	}
	return bankStmts, nil
}
