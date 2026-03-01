package payrolls_test

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/urequire"
	grc20factory "gno.land/r/demo/defi/grc20factory"
	"gno.land/r/demo/payrolls"
)

func TestPayrollsGRC20(t *testing.T) {
	monthlyPAY := chain.NewCoin("gno.land/r/demo/defi/grc20factory.PAY", 5_000_0000)
	monthlyGRC20s := chain.NewCoins(monthlyPAY)
	namespace := "finance"
	monthlyPayrollsCoins, err := payrolls.Coins(cross, nil, monthlyGRC20s)
	urequire.NoError(t, err)

	tokenCreator := testutils.TestAddress("pay-creator")
	payrollsRealm := testing.NewCodeRealm("gno.land/r/demo/payrolls")
	creator := testutils.TestAddress("creator")
	beneficiary := testutils.TestAddress("beneficiary")

	testing.SetOriginCaller(tokenCreator)
	testing.SetRealm(testing.NewUserRealm(tokenCreator))

	fundAmount := int64(20_000_000)
	grc20factory.NewWithAdmin(cross, "Pay", "PAY", 6, 0, fundAmount, tokenCreator)

	testing.SetOriginCaller(creator)
	testing.SetRealm(testing.NewUserRealm(creator))

	grc20factory.Faucet(cross, "PAY")
	grc20factory.Approve(cross, "PAY", payrollsRealm.Address(), fundAmount)

	payrolls.FundGRC20Reg(cross, creator.String(), namespace, monthlyPAY.Denom, -1)

	payrollID := payrolls.Create(cross, namespace, "Salary Monthly", beneficiary,
		payrolls.DistribMonthlyContinuous(cross, monthlyPayrollsCoins),
		payrolls.BreakupCDI(cross, monthlyPayrollsCoins),
	)

	creatorBalance := grc20factory.BalanceOf("PAY", creator)
	urequire.Equal(t, int64(0), creatorBalance, "creator balance after setup")

	realmBalance := grc20factory.BalanceOf("PAY", payrollsRealm.Address())
	urequire.Equal(t, fundAmount, realmBalance, "realm balance after setup")

	skipHeight := int64(10000)
	testing.SkipHeights(skipHeight)

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Claim(cross, payrollID, address(""))
	}, "claim payroll")

	expectedBeneficiaryBalance := int64(964506)

	beneficiaryBalance := grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance, "beneficiary balance after claim")

	realmBalance = grc20factory.BalanceOf("PAY", payrollsRealm.Address())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance, "realm balance after claim")

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Pause(cross, payrollID)
	}, "pause payroll")

	testing.SkipHeights(skipHeight)

	urequire.AbortsWithMessage(t, "nothing to claim", func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Claim(cross, payrollID, address(""))
	})

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Resume(cross, payrollID)
	}, "resume payroll")

	testing.SkipHeights(skipHeight)

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(creator)
		testing.SetRealm(testing.NewUserRealm(creator))

		payrolls.Stop(cross, payrollID)
	}, "stop payroll")

	beneficiaryBalance = grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance, "beneficiary balance after stop")

	realmBalance = grc20factory.BalanceOf("PAY", payrollsRealm.Address())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance, "realm balance after claim all")

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.ClaimAll(cross, address(""))
	}, "claim all payrolls")

	expectedBeneficiaryBalance = int64(1969060)

	beneficiaryBalance = grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance, "beneficiary balance after claim all")

	realmBalance = grc20factory.BalanceOf("PAY", payrollsRealm.Address())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance, "realm balance after claim all")

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(creator)
		testing.SetRealm(testing.NewUserRealm(creator))
		payrolls.WithdrawFunds(cross, namespace, monthlyPayrollsCoins[0].Denom, -1, "")
	}, "withdraw funds")

	beneficiaryBalance = grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance, "beneficiary balance after withdraw")
	creatorBalance = grc20factory.BalanceOf("PAY", creator)
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, creatorBalance, "creator balance after withdraw")

	urequire.AbortsWithMessage(t, "nothing to claim", func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.ClaimAll(cross, address(""))
	}, "claim all with no funds")
}

func TestPayrollsNative(t *testing.T) {
	monthlyUgnot := chain.NewCoin("ugnot", 5_000_0000)
	monthlyNatives := chain.NewCoins(monthlyUgnot)
	namespace := "native-finance"
	fundAmount := int64(20_000_000)
	monthlyPayrollsCoins, err := payrolls.Coins(cross, monthlyNatives, nil)
	urequire.NoError(t, err)

	payrollsRealm := testing.NewCodeRealm("gno.land/r/demo/payrolls")
	creator := testutils.TestAddress("creator")
	beneficiary := testutils.TestAddress("beneficiary")

	testing.SetOriginCaller(creator)
	testing.SetRealm(testing.NewUserRealm(creator))

	send := chain.NewCoins(chain.NewCoin("ugnot", fundAmount))
	testing.IssueCoins(payrollsRealm.Address(), send)
	testing.SetOriginSend(send)

	payrolls.FundNative(cross, creator.String(), namespace)

	payrollID := payrolls.Create(cross, namespace, "Salary Monthly", beneficiary,
		payrolls.DistribMonthlyContinuous(cross, monthlyPayrollsCoins),
		payrolls.BreakupCDI(cross, monthlyPayrollsCoins),
	)

	creatorBalance := payrolls.BalanceOf(t, creator, monthlyNatives[0].Denom)
	urequire.Equal(t, int64(0), creatorBalance, "creator balance after setup")

	realmBalance := payrolls.BalanceOf(t, payrollsRealm.Address(), monthlyNatives[0].Denom)
	urequire.Equal(t, fundAmount, realmBalance, "realm balance after setup")

	skipHeight := uint64(10000)
	testing.SkipHeights(int64(skipHeight))

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Claim(cross, payrollID, address(""))
	})

	expectedBeneficiaryBalance := int64(964506)

	beneficiaryBalance := payrolls.BalanceOf(t, beneficiary, monthlyNatives[0].Denom)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance, "beneficiary balance after claim")

	realmBalance = payrolls.BalanceOf(t, payrollsRealm.Address(), monthlyNatives[0].Denom)
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance, "realm balance after claim")

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Pause(cross, payrollID)
	})

	testing.SkipHeights(int64(skipHeight))

	urequire.AbortsWithMessage(t, "nothing to claim", func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Claim(cross, payrollID, address(""))
	}, "claim paused payroll")

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Resume(cross, payrollID)
	}, "resume payroll")

	testing.SkipHeights(int64(skipHeight))

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(creator)
		testing.SetRealm(testing.NewUserRealm(creator))

		payrolls.Stop(cross, payrollID)
	}, "stop payroll")

	beneficiaryBalance = payrolls.BalanceOf(t, beneficiary, monthlyNatives[0].Denom)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance, "beneficiary balance after stop")

	realmBalance = payrolls.BalanceOf(t, payrollsRealm.Address(), monthlyNatives[0].Denom)
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance, "realm balance after stop")

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.ClaimAll(cross, address(""))
	}, "claim all payrolls")

	expectedBeneficiaryBalance = int64(1969060)

	beneficiaryBalance = payrolls.BalanceOf(t, beneficiary, monthlyNatives[0].Denom)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance, "beneficiary balance after claim all")

	realmBalance = payrolls.BalanceOf(t, payrollsRealm.Address(), monthlyNatives[0].Denom)
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance, "realm balance after claim all")

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(creator)
		testing.SetRealm(testing.NewUserRealm(creator))
		payrolls.WithdrawFunds(cross, namespace, monthlyPayrollsCoins[0].Denom, -1, "")
	}, "withdraw funds")

	beneficiaryBalance = payrolls.BalanceOf(t, beneficiary, monthlyNatives[0].Denom)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance, "beneficiary balance after withdraw")
	creatorBalance = payrolls.BalanceOf(t, creator, monthlyNatives[0].Denom)
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, creatorBalance, "creator balance after withdraw")

	urequire.AbortsWithMessage(t, "nothing to claim", func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.ClaimAll(cross, address(""))
	}, "claim all payrolls")
}
