package payrolls_test

import (
	"chain"
	"chain/banker"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
	grc20factory "gno.land/r/demo/grc20factory"
	"gno.land/r/demo/payrolls"
)

func TestPayrollsGRC20(t *testing.T) {
	monthlyPAY := chain.NewCoin("gno.land/r/demo/grc20factory.PAY", 5_000_0000)
	monthlyGRC20s := chain.NewCoins(monthlyPAY)
	namespace := "finance"
	monthlyPayrollsCoins, err := payrolls.Coins(nil, monthlyGRC20s)
	urequire.NoError(t, err)

	tokenCreator := testutils.TestAddress("pay-creator")
	payrollsRealm := testing.NewCodeRealm("gno.land/r/demo/payrolls")
	creator := testutils.TestAddress("creator")
	beneficiary := testutils.TestAddress("beneficiary")

	std.TestSetOrigCaller(tokenCreator)
	std.TestSetRealm(testing.NewUserRealm(tokenCreator))

	fundAmount := uint64(20_000_000)
	grc20factory.New("Pay", "PAY", 6, 0, fundAmount)

	std.TestSetOrigCaller(creator)
	std.TestSetRealm(testing.NewUserRealm(creator))

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
		std.TestSetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Claim(payrollID, address(""))
	})

	expectedBeneficiaryBalance := uint64(964506)

	beneficiaryBalance := grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)

	realmBalance = grc20factory.BalanceOf("PAY", payrollsRealm.Addr())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance)

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Pause(payrollID)
	})

	std.TestSkipHeights(int64(skipHeight))

	urequire.PanicsWithMessage(t, "nothing to claim", func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Claim(payrollID, address(""))
	})

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Resume(payrollID)
	})

	std.TestSkipHeights(int64(skipHeight))

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(creator)
		std.TestSetRealm(testing.NewUserRealm(creator))

		payrolls.Stop(payrollID)
	})

	beneficiaryBalance = grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)

	realmBalance = grc20factory.BalanceOf("PAY", payrollsRealm.Addr())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance)

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(testing.NewUserRealm(beneficiary))
		payrolls.ClaimAll(address(""))
	})

	expectedBeneficiaryBalance = uint64(1969060)

	beneficiaryBalance = grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)

	realmBalance = grc20factory.BalanceOf("PAY", payrollsRealm.Addr())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance)

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(creator)
		std.TestSetRealm(testing.NewUserRealm(creator))
		payrolls.WithdrawFunds(namespace, monthlyPayrollsCoins[0].Denom, -1, "")
	})

	beneficiaryBalance = grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)
	creatorBalance = grc20factory.BalanceOf("PAY", creator)
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, creatorBalance)

	urequire.PanicsWithMessage(t, "nothing to claim", func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(testing.NewUserRealm(beneficiary))
		payrolls.ClaimAll(address(""))
	})
}

func TestPayrollsNative(t *testing.T) {
	monthlyUgnot := chain.NewCoin("ugnot", 5_000_0000)
	monthlyNatives := chain.NewCoins(monthlyUgnot)
	namespace := "native-finance"
	fundAmount := int64(20_000_000)
	monthlyPayrollsCoins, err := payrolls.Coins(monthlyNatives, nil)
	urequire.NoError(t, err)

	balanceOf := func(addr address) int64 {
		banker_ := banker.NewBanker(banker.BankerTypeReadonly)
		return banker_.GetCoins(addr).AmountOf(monthlyNatives[0].Denom)
	}

	payrollsRealm := testing.NewCodeRealm("gno.land/r/demo/payrolls")
	creator := testutils.TestAddress("creator")
	beneficiary := testutils.TestAddress("beneficiary")

	std.TestSetOrigCaller(creator)
	std.TestSetRealm(testing.NewUserRealm(creator))

	send := chain.NewCoins(chain.NewCoin("ugnot", fundAmount))
	std.TestIssueCoins(payrollsRealm.Addr(), send)
	std.TestSetOrigSend(send, nil)

	payrolls.FundNative(creator.String(), namespace)

	payrollID := payrolls.Create(namespace, "Salary Monthly", beneficiary,
		payrolls.DistribMonthlyContinuous(monthlyPayrollsCoins),
		payrolls.BreakupCDI(monthlyPayrollsCoins),
	)

	creatorBalance := balanceOf(creator)
	urequire.Equal(t, int64(0), creatorBalance)

	realmBalance := balanceOf(payrollsRealm.Addr())
	urequire.Equal(t, fundAmount, realmBalance)

	skipHeight := uint64(10000)
	std.TestSkipHeights(int64(skipHeight))

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Claim(payrollID, address(""))
	})

	expectedBeneficiaryBalance := int64(964506)

	beneficiaryBalance := balanceOf(beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)

	realmBalance = balanceOf(payrollsRealm.Addr())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance)

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Pause(payrollID)
	})

	std.TestSkipHeights(int64(skipHeight))

	urequire.PanicsWithMessage(t, "nothing to claim", func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Claim(payrollID, address(""))
	})

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Resume(payrollID)
	})

	std.TestSkipHeights(int64(skipHeight))

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(creator)
		std.TestSetRealm(testing.NewUserRealm(creator))

		payrolls.Stop(payrollID)
	})

	beneficiaryBalance = balanceOf(beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)

	realmBalance = balanceOf(payrollsRealm.Addr())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance)

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(testing.NewUserRealm(beneficiary))
		payrolls.ClaimAll(address(""))
	})

	expectedBeneficiaryBalance = int64(1969060)

	beneficiaryBalance = balanceOf(beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)

	realmBalance = balanceOf(payrollsRealm.Addr())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance)

	urequire.NotPanics(t, func() {
		std.TestSetOrigCaller(creator)
		std.TestSetRealm(testing.NewUserRealm(creator))
		payrolls.WithdrawFunds(namespace, monthlyPayrollsCoins[0].Denom, -1, "")
	})

	beneficiaryBalance = balanceOf(beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)
	creatorBalance = balanceOf(creator)
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, creatorBalance)

	urequire.PanicsWithMessage(t, "nothing to claim", func() {
		std.TestSetOrigCaller(beneficiary)
		std.TestSetRealm(testing.NewUserRealm(beneficiary))
		payrolls.ClaimAll(address(""))
	})
}
