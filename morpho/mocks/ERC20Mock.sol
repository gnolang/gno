// SPDX-License-Identifier: GPL-2.0-or-later
pragma solidity ^0.8.0;

import {IERC20} from "./interfaces/IERC20.sol";

contract ERC20Mock is IERC20 {
    uint256 public totalSupply;

    mapping(address account => uint256) public balanceOf;
    mapping(address account => mapping(address spender => uint256)) public allowance;

    function setBalance(address account, uint256 amount) public virtual {
        if (amount > balanceOf[account]) totalSupply += amount - balanceOf[account];
        else totalSupply -= balanceOf[account] - amount;

        balanceOf[account] = amount;
    }

    function approve(address spender, uint256 amount) public virtual returns (bool) {
        allowance[msg.sender][spender] = amount;

        emit Approval(msg.sender, spender, amount);

        return true;
    }

    function transfer(address to, uint256 amount) public virtual returns (bool) {
        require(balanceOf[msg.sender] >= amount, "insufficient balance");

        balanceOf[msg.sender] -= amount;
        balanceOf[to] += amount;

        emit Transfer(msg.sender, to, amount);

        return true;
    }

    function transferFrom(address from, address to, uint256 amount) public virtual returns (bool) {
        require(allowance[from][msg.sender] >= amount, "insufficient allowance");

        allowance[from][msg.sender] -= amount;

        require(balanceOf[from] >= amount, "insufficient balance");

        balanceOf[from] -= amount;
        balanceOf[to] += amount;

        emit Transfer(from, to, amount);

        return true;
    }
}
