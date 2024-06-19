/// Withdraw a bid that was overbid.
function withdraw() external returns (bool) {
    uint amount = pendingReturns[msg.sender];
    if (amount > 0) {
        // It is important to set this to zero because the recipient
        // can call this function again as part of the receiving call
        // before `send` returns.
        pendingReturns[msg.sender] = 0;

        // msg.sender is not of type `address payable` and must be
        // explicitly converted using `payable(msg.sender)` in order
        // use the member function `send()`.
        if (!payable(msg.sender).send(amount)) {
            // No need to call throw here, just reset the amount owing
            pendingReturns[msg.sender] = amount;
            return false;
        }
    }
    return true;
}