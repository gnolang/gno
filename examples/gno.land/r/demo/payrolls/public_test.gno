package payrolls_test

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
	grc20factory "gno.land/r/demo/grc20factory"
	"gno.land/r/demo/payrolls"
)

func TestPayrollsGRC20(t *testing.T) {
	monthlyPAY := std.NewCoin("gno.land/r/demo/grc20factory.PAY", 5_000_0000)
	monthlyGRC20s := std.NewCoins(monthlyPAY)
	namespace := "finance"
	monthlyPayrollsCoins, err := payrolls.Coins(nil, monthlyGRC20s)
	urequire.NoError(t, err)

	tokenCreator := testutils.TestAddress("pay-creator")
	payrollsRealm := std.NewCodeRealm("gno.land/r/demo/payrolls")
	creator := testutils.TestAddress("creator")
	beneficiary := testutils.TestAddress("beneficiary")

	std.TestSetOrigCaller(tokenCreator)
	std.TestSetRealm(std.NewUserRealm(tokenCreator))

	fundAmount := uint64(20_000_000)
	grc20factory.New("Pay", "PAY", 6, 0, fundAmount)

	std.TestSetOrigCaller(creator)
	std.TestSetRealm(std.NewUserRealm(creator))

	grc20factory.Faucet("PAY")
	grc20factory.Approve("PAY", payrollsRealm.Addr(), fundAmount)

	payrolls.FundGRC20Reg(creator.String(), namespace, monthlyPAY.Denom, -1)

	payrollID := payrolls.Create(namespace, "Salary Monthly", beneficiary,
		payrolls.DistribMonthlyContinuous(monthlyPayrollsCoins),
		payrolls.BreakupCDI(monthlyPayrollsCoins),
	)

	creatorBalance := grc20factory.BalanceOf("PAY", creator)
	urequire.Equal(t, uint64(0), creatorBalance)

	realmBalance := grc20factory.BalanceOf("PAY", payrollsRealm.Addr())
	urequire.Equal(t, fundAmount, realmBalance)

	skipHeight := uint64(10000)
	std.TestSkipHeights(int64(skipHeight))

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(std.NewUserRealm(beneficiary))
		payrolls.Claim(payrollID, std.Address(""))
	})

	expectedBeneficiaryBalance := uint64(964506)

	beneficiaryBalance := grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)

	realmBalance = grc20factory.BalanceOf("PAY", payrollsRealm.Addr())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance)

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(std.NewUserRealm(beneficiary))
		payrolls.Pause(payrollID)
	})

	std.TestSkipHeights(int64(skipHeight))

	urequire.PanicsWithMessage(t, "nothing to claim", func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(std.NewUserRealm(beneficiary))
		payrolls.Claim(payrollID, std.Address(""))
	})

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(std.NewUserRealm(beneficiary))
		payrolls.Resume(payrollID)
	})

	std.TestSkipHeights(int64(skipHeight))

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(creator)
		std.TestSetRealm(std.NewUserRealm(creator))

		payrolls.Stop(payrollID)
	})

	beneficiaryBalance = grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)

	realmBalance = grc20factory.BalanceOf("PAY", payrollsRealm.Addr())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance)

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(std.NewUserRealm(beneficiary))
		payrolls.Claim(payrollID, std.Address(""))
	})

	expectedBeneficiaryBalance = uint64(1969060)

	beneficiaryBalance = grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)

	realmBalance = grc20factory.BalanceOf("PAY", payrollsRealm.Addr())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance)

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(creator)
		std.TestSetRealm(std.NewUserRealm(creator))
		payrolls.WithdrawFunds(namespace, monthlyPayrollsCoins[0].Denom, -1, "")
	})

	beneficiaryBalance = grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)
	creatorBalance = grc20factory.BalanceOf("PAY", creator)
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, creatorBalance)
}
