package payrolls_test

import (
	"chain"
	"chain/banker"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/urequire"
	grc20factory "gno.land/r/demo/defi/grc20factory"
	"gno.land/r/demo/payrolls"
)

func TestPayrollsGRC20(t *testing.T) {
	monthlyPAY := chain.NewCoin("gno.land/r/demo/grc20factory.PAY", 5_000_0000)
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
	grc20factory.New(cross, "Pay", "PAY", 6, 0, fundAmount)

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
	urequire.Equal(t, uint64(0), creatorBalance)

	realmBalance := grc20factory.BalanceOf("PAY", payrollsRealm.Address())
	urequire.Equal(t, fundAmount, realmBalance)

	skipHeight := uint64(10000)
	testing.SkipHeights(int64(skipHeight))

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Claim(cross, payrollID, address(""))
	})

	expectedBeneficiaryBalance := int64(964506)

	beneficiaryBalance := grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)

	realmBalance = grc20factory.BalanceOf("PAY", payrollsRealm.Address())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance)

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Pause(cross, payrollID)
	})

	testing.SkipHeights(int64(skipHeight))

	urequire.PanicsWithMessage(t, "nothing to claim", func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Claim(cross, payrollID, address(""))
	})

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Resume(cross, payrollID)
	})

	testing.SkipHeights(int64(skipHeight))

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(creator)
		testing.SetRealm(testing.NewUserRealm(creator))

		payrolls.Stop(cross, payrollID)
	})

	beneficiaryBalance = grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)

	realmBalance = grc20factory.BalanceOf("PAY", payrollsRealm.Address())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance)

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.ClaimAll(cross, address(""))
	})

	expectedBeneficiaryBalance = int64(1969060)

	beneficiaryBalance = grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)

	realmBalance = grc20factory.BalanceOf("PAY", payrollsRealm.Address())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance)

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(creator)
		testing.SetRealm(testing.NewUserRealm(creator))
		payrolls.WithdrawFunds(cross, namespace, monthlyPayrollsCoins[0].Denom, -1, "")
	})

	beneficiaryBalance = grc20factory.BalanceOf("PAY", beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)
	creatorBalance = grc20factory.BalanceOf("PAY", creator)
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, creatorBalance)

	urequire.PanicsWithMessage(t, "nothing to claim", func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.ClaimAll(cross, address(""))
	})
}

func TestPayrollsNative(t *testing.T) {
	monthlyUgnot := chain.NewCoin("ugnot", 5_000_0000)
	monthlyNatives := chain.NewCoins(monthlyUgnot)
	namespace := "native-finance"
	fundAmount := int64(20_000_000)
	monthlyPayrollsCoins, err := payrolls.Coins(cross, monthlyNatives, nil)
	urequire.NoError(t, err)

	balanceOf := func(addr address) int64 {
		banker_ := banker.NewBanker(banker.BankerTypeReadonly)
		return banker_.GetCoins(addr).AmountOf(monthlyNatives[0].Denom)
	}

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

	creatorBalance := balanceOf(creator)
	urequire.Equal(t, int64(0), creatorBalance)

	realmBalance := balanceOf(payrollsRealm.Address())
	urequire.Equal(t, fundAmount, realmBalance)

	skipHeight := uint64(10000)
	testing.SkipHeights(int64(skipHeight))

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Claim(cross, payrollID, address(""))
	})

	expectedBeneficiaryBalance := int64(964506)

	beneficiaryBalance := balanceOf(beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)

	realmBalance = balanceOf(payrollsRealm.Address())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance)

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Pause(cross, payrollID)
	})

	testing.SkipHeights(int64(skipHeight))

	urequire.PanicsWithMessage(t, "nothing to claim", func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Claim(cross, payrollID, address(""))
	})

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.Resume(cross, payrollID)
	})

	testing.SkipHeights(int64(skipHeight))

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(creator)
		testing.SetRealm(testing.NewUserRealm(creator))

		payrolls.Stop(cross, payrollID)
	})

	beneficiaryBalance = balanceOf(beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)

	realmBalance = balanceOf(payrollsRealm.Address())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance)

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.ClaimAll(cross, address(""))
	})

	expectedBeneficiaryBalance = int64(1969060)

	beneficiaryBalance = balanceOf(beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)

	realmBalance = balanceOf(payrollsRealm.Address())
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, realmBalance)

	urequire.NotPanics(t, func() {
		testing.SetOriginCaller(creator)
		testing.SetRealm(testing.NewUserRealm(creator))
		payrolls.WithdrawFunds(cross, namespace, monthlyPayrollsCoins[0].Denom, -1, "")
	})

	beneficiaryBalance = balanceOf(beneficiary)
	urequire.Equal(t, expectedBeneficiaryBalance, beneficiaryBalance)
	creatorBalance = balanceOf(creator)
	urequire.Equal(t, fundAmount-expectedBeneficiaryBalance, creatorBalance)

	urequire.PanicsWithMessage(t, "nothing to claim", func() {
		testing.SetOriginCaller(beneficiary)
		testing.SetRealm(testing.NewUserRealm(beneficiary))
		payrolls.ClaimAll(cross, address(""))
	})
}
